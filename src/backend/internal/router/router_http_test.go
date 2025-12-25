package router

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	jwtauth "evening-gown/internal/auth"
	"evening-gown/internal/bootstrap"
	"evening-gown/internal/cache"
	"evening-gown/internal/config"
	adminHandlers "evening-gown/internal/handler/admin"
	authHandlerPkg "evening-gown/internal/handler/auth"
	"evening-gown/internal/handler/health"
	publicHandlers "evening-gown/internal/handler/public"
	"evening-gown/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRouter_Recovery_LogsStackOnPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	old := slog.Default()
	defer slog.SetDefault(old)

	var buf bytes.Buffer
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))

	r := New(Dependencies{Health: health.New(nil, nil, nil)})
	r.GET("/__panic", func(c *gin.Context) { panic("boom") })

	resp := doRequest(t, r, http.MethodGet, "/__panic", nil, nil)
	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d got %d: %s", http.StatusInternalServerError, resp.Code, resp.Body.String())
	}

	out := buf.String()
	if !strings.Contains(out, "panic recovered") {
		t.Fatalf("expected panic log; got: %s", out)
	}
	if !strings.Contains(out, "\"stack\"") {
		t.Fatalf("expected stack in log; got: %s", out)
	}
	if !strings.Contains(out, "\"request_id\"") {
		t.Fatalf("expected request_id in log; got: %s", out)
	}
}

func TestRouter_Ping_ReturnsJSONAndRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := New(Dependencies{Health: health.New(nil, nil, nil)})
	resp := doRequest(t, r, http.MethodGet, "/ping", nil, map[string]string{"Accept": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
	if strings.TrimSpace(resp.Header().Get("X-Request-Id")) == "" {
		t.Fatalf("expected X-Request-Id header to be set")
	}

	var got map[string]any
	mustJSON(t, resp.Body.Bytes(), &got)
	if got["message"] != "pong" {
		t.Fatalf("expected message=pong, got %v", got["message"])
	}
}

func TestRouter_Healthz_WhenDepsDisabled_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := New(Dependencies{Health: health.New(nil, nil, nil)})
	resp := doRequest(t, r, http.MethodGet, "/healthz", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var got map[string]any
	mustJSON(t, resp.Body.Bytes(), &got)
	checks, _ := got["checks"].(map[string]any)
	if checks == nil {
		t.Fatalf("expected checks object")
	}
	if checks["postgres"] != "disabled" || checks["redis"] != "disabled" || checks["minio"] != "disabled" {
		t.Fatalf("unexpected checks: %#v", checks)
	}
}

func TestRouter_AuthDevTokenIssuer_IsGated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtCfg := config.JWTConfig{Secret: "test-secret", Issuer: "evening-gown", ExpiresIn: time.Hour}
	authHandler := authHandlerPkg.New(jwtCfg)

	// Disabled by default.
	{
		r := New(Dependencies{Auth: authHandler, EnableDevTokenIssuer: false})
		resp := doRequest(t, r, http.MethodPost, "/auth/token", []byte(`{"sub":"123"}`), jsonHeaders())
		if resp.Code != http.StatusNotFound {
			t.Fatalf("expected %d, got %d: %s", http.StatusNotFound, resp.Code, resp.Body.String())
		}
	}

	// Enabled.
	{
		r := New(Dependencies{Auth: authHandler, EnableDevTokenIssuer: true})
		resp := doRequest(t, r, http.MethodPost, "/auth/token", []byte(`{"sub":"123"}`), jsonHeaders())
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
		var got map[string]any
		mustJSON(t, resp.Body.Bytes(), &got)
		token, _ := got["token"].(string)
		if strings.TrimSpace(token) == "" {
			t.Fatalf("expected token in response")
		}

		verify := doRequest(t, r, http.MethodGet, "/auth/verify", nil, map[string]string{"Authorization": "Bearer " + token})
		if verify.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, verify.Code, verify.Body.String())
		}
		var v map[string]any
		mustJSON(t, verify.Body.Bytes(), &v)
		claims, _ := v["claims"].(map[string]any)
		if claims == nil || claims["sub"] != "123" {
			t.Fatalf("expected claims.sub=123, got %#v", claims)
		}
	}
}

func TestRouter_CORS_DefaultAndRestricted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Default: allow-all.
	{
		t.Setenv("CORS_ALLOW_ORIGINS", "")
		r := New(Dependencies{Health: health.New(nil, nil, nil)})
		resp := doRequest(t, r, http.MethodGet, "http://api.local/ping", nil, map[string]string{
			"Origin": "https://web.local",
		})
		if resp.Header().Get("Access-Control-Allow-Origin") == "" {
			t.Fatalf("expected Access-Control-Allow-Origin to be set; headers=%v", resp.Header())
		}
	}

	// Restricted.
	{
		t.Setenv("CORS_ALLOW_ORIGINS", "https://a.example,https://b.example")
		r := New(Dependencies{Health: health.New(nil, nil, nil)})

		okResp := doRequest(t, r, http.MethodGet, "http://api.local/ping", nil, map[string]string{
			"Origin": "https://a.example",
		})
		if okResp.Header().Get("Access-Control-Allow-Origin") != "https://a.example" {
			t.Fatalf("expected allow-origin to echo request origin, got %q", okResp.Header().Get("Access-Control-Allow-Origin"))
		}

		denyResp := doRequest(t, r, http.MethodGet, "http://api.local/ping", nil, map[string]string{
			"Origin": "https://not-allowed.example",
		})
		if denyResp.Header().Get("Access-Control-Allow-Origin") != "" {
			t.Fatalf("expected no allow-origin for disallowed origin, got %q", denyResp.Header().Get("Access-Control-Allow-Origin"))
		}
	}
}

func TestRouter_Pprof_IsGatedByEnv(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Disabled.
	{
		t.Setenv("ENABLE_PPROF", "false")
		r := New(Dependencies{})
		resp := doRequest(t, r, http.MethodGet, "/debug/pprof/", nil, nil)
		if resp.Code != http.StatusNotFound {
			t.Fatalf("expected %d, got %d", http.StatusNotFound, resp.Code)
		}
	}

	// Enabled.
	{
		t.Setenv("ENABLE_PPROF", "true")
		r := New(Dependencies{})
		resp := doRequest(t, r, http.MethodGet, "/debug/pprof/", nil, nil)
		if resp.Code == http.StatusNotFound {
			t.Fatalf("expected pprof route to be registered")
		}
	}
}

func TestRouter_PublicAndAdmin_APIs_EndToEnd(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openTestDB(t)

	jwtCfg := config.JWTConfig{Secret: "test-secret", Issuer: "evening-gown", ExpiresIn: time.Hour}
	jwtSvc, err := jwtauth.New(jwtCfg)
	if err != nil {
		t.Fatalf("create jwt service: %v", err)
	}

	adminEmail := "admin@example.com"
	adminPassword := "passw0rd123" // >= 10 chars
	if err := bootstrap.EnsureSingleAdmin(db, adminEmail, adminPassword); err != nil {
		t.Fatalf("ensure admin: %v", err)
	}

	publicCache := cache.NewPublicCache(nil)

	deps := Dependencies{Health: health.New(db, nil, nil), Auth: authHandlerPkg.New(jwtCfg), EnableDevTokenIssuer: true}
	deps.Public.Products = publicHandlers.NewProductsHandler(db, publicCache)
	deps.Public.Updates = publicHandlers.NewUpdatesHandler(db, publicCache)
	deps.Public.Contacts = publicHandlers.NewContactsHandler(db)
	deps.Public.Events = publicHandlers.NewEventsHandler(db)

	deps.Admin.Auth = adminHandlers.NewAuthHandler(db, jwtSvc)
	deps.Admin.Products = adminHandlers.NewProductsHandler(db, publicCache)
	deps.Admin.Updates = adminHandlers.NewUpdatesHandler(db, publicCache)
	deps.Admin.Contacts = adminHandlers.NewContactsHandler(db)
	deps.Admin.Events = adminHandlers.NewEventsHandler(db)
	deps.Admin.AuthMiddleware = middleware.AdminAuth(db, jwtSvc)

	r := New(deps)

	// Public: contacts validation.
	{
		resp := doRequest(t, r, http.MethodPost, "/api/v1/contacts", []byte(`{"name":"a"}`), jsonHeaders())
		if resp.Code != http.StatusBadRequest {
			t.Fatalf("expected %d, got %d: %s", http.StatusBadRequest, resp.Code, resp.Body.String())
		}
	}

	// Public: create contact.
	var contactID uint
	{
		resp := doRequest(t, r, http.MethodPost, "/api/v1/contacts", []byte(`{"name":"Alice","phone":"13800000000","message":"hi"}`), jsonHeaders())
		if resp.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d: %s", http.StatusCreated, resp.Code, resp.Body.String())
		}
		var got map[string]any
		mustJSON(t, resp.Body.Bytes(), &got)
		contactID = mustUintFromJSONNumber(t, got["id"])
		if contactID == 0 {
			t.Fatalf("expected id in response")
		}
	}

	// Public: create event.
	var eventID uint
	{
		resp := doRequest(t, r, http.MethodPost, "/api/v1/events", []byte(`{"event_type":"view","page_url":"/"}`), jsonHeaders())
		if resp.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d: %s", http.StatusCreated, resp.Code, resp.Body.String())
		}
		var got map[string]any
		mustJSON(t, resp.Body.Bytes(), &got)
		eventID = mustUintFromJSONNumber(t, got["id"])
		if eventID == 0 {
			t.Fatalf("expected event id")
		}
	}

	// Admin: login.
	var adminToken string
	{
		resp := doRequest(t, r, http.MethodPost, "/api/v1/admin/auth/login", []byte(`{"email":"`+adminEmail+`","password":"`+adminPassword+`"}`), jsonHeaders())
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
		var got map[string]any
		mustJSON(t, resp.Body.Bytes(), &got)
		adminToken, _ = got["token"].(string)
		if strings.TrimSpace(adminToken) == "" {
			t.Fatalf("expected token")
		}
	}

	// Admin: me (authorized).
	{
		resp := doRequest(t, r, http.MethodGet, "/api/v1/admin/me", nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
	}

	// Admin: list contacts includes created contact.
	{
		resp := doRequest(t, r, http.MethodGet, "/api/v1/admin/contacts", nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
		var got map[string]any
		mustJSON(t, resp.Body.Bytes(), &got)
		items, _ := got["items"].([]any)
		if len(items) == 0 {
			t.Fatalf("expected at least 1 contact")
		}
	}

	// Admin: update contact status.
	{
		path := "/api/v1/admin/contacts/" + strconv.FormatUint(uint64(contactID), 10)
		resp := doRequest(t, r, http.MethodPatch, path, []byte(`{"status":"contacted"}`), withAuth(jsonHeaders(), adminToken))
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
	}

	// Admin: get contact.
	{
		path := "/api/v1/admin/contacts/" + strconv.FormatUint(uint64(contactID), 10)
		resp := doRequest(t, r, http.MethodGet, path, nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
	}

	// Admin: create product (draft).
	var productID uint
	{
		resp := doRequest(t, r, http.MethodPost, "/api/v1/admin/products", []byte(`{"styleNo":1001,"season":"ss25","category":"gown","availability":"in_stock","isNew":true}`), withAuth(jsonHeaders(), adminToken))
		if resp.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d: %s", http.StatusCreated, resp.Code, resp.Body.String())
		}
		var got map[string]any
		mustJSON(t, resp.Body.Bytes(), &got)
		productID = mustUintFromJSONNumber(t, got["id"])
		if productID == 0 {
			t.Fatalf("expected product id")
		}
	}

	// Public list should still be empty (not published).
	{
		resp := doRequest(t, r, http.MethodGet, "/api/v1/products", nil, nil)
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
		var got map[string]any
		mustJSON(t, resp.Body.Bytes(), &got)
		total := mustUintFromJSONNumber(t, got["total"])
		if total != 0 {
			t.Fatalf("expected total=0, got %#v", got["total"])
		}
	}

	// Publish product.
	{
		path := "/api/v1/admin/products/" + strconv.FormatUint(uint64(productID), 10) + "/publish"
		resp := doRequest(t, r, http.MethodPost, path, nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
	}

	// Public list should now include it.
	{
		resp := doRequest(t, r, http.MethodGet, "/api/v1/products?limit=10", nil, nil)
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
		var got map[string]any
		mustJSON(t, resp.Body.Bytes(), &got)
		total := mustUintFromJSONNumber(t, got["total"])
		if total != 1 {
			t.Fatalf("expected total=1, got %#v", got["total"])
		}
	}

	// Unpublish product.
	{
		path := "/api/v1/admin/products/" + strconv.FormatUint(uint64(productID), 10) + "/unpublish"
		resp := doRequest(t, r, http.MethodPost, path, nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
	}

	// Public list should be empty again.
	{
		resp := doRequest(t, r, http.MethodGet, "/api/v1/products", nil, nil)
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
		var got map[string]any
		mustJSON(t, resp.Body.Bytes(), &got)
		total := mustUintFromJSONNumber(t, got["total"])
		if total != 0 {
			t.Fatalf("expected total=0, got %#v", got["total"])
		}
	}

	// Admin: delete product.
	{
		path := "/api/v1/admin/products/" + strconv.FormatUint(uint64(productID), 10)
		resp := doRequest(t, r, http.MethodDelete, path, nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if resp.Code != http.StatusNoContent {
			t.Fatalf("expected %d, got %d: %s", http.StatusNoContent, resp.Code, resp.Body.String())
		}
		getResp := doRequest(t, r, http.MethodGet, path, nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if getResp.Code != http.StatusNotFound {
			t.Fatalf("expected %d, got %d: %s", http.StatusNotFound, getResp.Code, getResp.Body.String())
		}
	}

	// Admin: create update (draft).
	var updateID uint
	{
		resp := doRequest(t, r, http.MethodPost, "/api/v1/admin/updates", []byte(`{"type":"company","status":"draft","title":"Hello","body":"World"}`), withAuth(jsonHeaders(), adminToken))
		if resp.Code != http.StatusCreated {
			t.Fatalf("expected %d, got %d: %s", http.StatusCreated, resp.Code, resp.Body.String())
		}
		var got map[string]any
		mustJSON(t, resp.Body.Bytes(), &got)
		updateID = mustUintFromJSONNumber(t, got["id"])
		if updateID == 0 {
			t.Fatalf("expected update id")
		}
	}

	// Public updates should be empty (not published).
	{
		resp := doRequest(t, r, http.MethodGet, "/api/v1/updates", nil, nil)
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
		var got map[string]any
		mustJSON(t, resp.Body.Bytes(), &got)
		total := mustUintFromJSONNumber(t, got["total"])
		if total != 0 {
			t.Fatalf("expected total=0, got %#v", got["total"])
		}
	}

	// Publish update.
	{
		path := "/api/v1/admin/updates/" + strconv.FormatUint(uint64(updateID), 10) + "/publish"
		resp := doRequest(t, r, http.MethodPost, path, nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
	}

	// Public updates should now include it.
	{
		resp := doRequest(t, r, http.MethodGet, "/api/v1/updates", nil, nil)
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
		var got map[string]any
		mustJSON(t, resp.Body.Bytes(), &got)
		total := mustUintFromJSONNumber(t, got["total"])
		if total != 1 {
			t.Fatalf("expected total=1, got %#v", got["total"])
		}
	}

	// Unpublish update.
	{
		path := "/api/v1/admin/updates/" + strconv.FormatUint(uint64(updateID), 10) + "/unpublish"
		resp := doRequest(t, r, http.MethodPost, path, nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
	}

	// Public updates should be empty again.
	{
		resp := doRequest(t, r, http.MethodGet, "/api/v1/updates", nil, nil)
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
		var got map[string]any
		mustJSON(t, resp.Body.Bytes(), &got)
		total := mustUintFromJSONNumber(t, got["total"])
		if total != 0 {
			t.Fatalf("expected total=0, got %#v", got["total"])
		}
	}

	// Admin: delete update.
	{
		path := "/api/v1/admin/updates/" + strconv.FormatUint(uint64(updateID), 10)
		resp := doRequest(t, r, http.MethodDelete, path, nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if resp.Code != http.StatusNoContent {
			t.Fatalf("expected %d, got %d: %s", http.StatusNoContent, resp.Code, resp.Body.String())
		}
		getResp := doRequest(t, r, http.MethodGet, path, nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if getResp.Code != http.StatusNotFound {
			t.Fatalf("expected %d, got %d: %s", http.StatusNotFound, getResp.Code, getResp.Body.String())
		}
	}

	// Admin: get & delete event.
	{
		getPath := "/api/v1/admin/events/" + strconv.FormatUint(uint64(eventID), 10)
		getResp := doRequest(t, r, http.MethodGet, getPath, nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if getResp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, getResp.Code, getResp.Body.String())
		}
		delResp := doRequest(t, r, http.MethodDelete, getPath, nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if delResp.Code != http.StatusNoContent {
			t.Fatalf("expected %d, got %d: %s", http.StatusNoContent, delResp.Code, delResp.Body.String())
		}
		after := doRequest(t, r, http.MethodGet, getPath, nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if after.Code != http.StatusNotFound {
			t.Fatalf("expected %d, got %d: %s", http.StatusNotFound, after.Code, after.Body.String())
		}
	}

	// Admin: delete contact.
	{
		path := "/api/v1/admin/contacts/" + strconv.FormatUint(uint64(contactID), 10)
		resp := doRequest(t, r, http.MethodDelete, path, nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if resp.Code != http.StatusNoContent {
			t.Fatalf("expected %d, got %d: %s", http.StatusNoContent, resp.Code, resp.Body.String())
		}
		getResp := doRequest(t, r, http.MethodGet, path, nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if getResp.Code != http.StatusNotFound {
			t.Fatalf("expected %d, got %d: %s", http.StatusNotFound, getResp.Code, getResp.Body.String())
		}
	}

	// Admin: change password should force logout (old token becomes invalid).
	newAdminPassword := "passw0rd456" // >= 10 chars
	{
		resp := doRequest(t, r, http.MethodPatch, "/api/v1/admin/me/password", []byte(`{"oldPassword":"`+adminPassword+`","newPassword":"`+newAdminPassword+`"}`), withAuth(jsonHeaders(), adminToken))
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
	}
	{
		resp := doRequest(t, r, http.MethodGet, "/api/v1/admin/me", nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if resp.Code != http.StatusUnauthorized {
			t.Fatalf("expected %d, got %d: %s", http.StatusUnauthorized, resp.Code, resp.Body.String())
		}
	}
	{
		resp := doRequest(t, r, http.MethodPost, "/api/v1/admin/auth/login", []byte(`{"email":"`+adminEmail+`","password":"`+newAdminPassword+`"}`), jsonHeaders())
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
		var got map[string]any
		mustJSON(t, resp.Body.Bytes(), &got)
		adminToken, _ = got["token"].(string)
		if strings.TrimSpace(adminToken) == "" {
			t.Fatalf("expected token")
		}
	}
	{
		resp := doRequest(t, r, http.MethodGet, "/api/v1/admin/me", nil, map[string]string{"Authorization": "Bearer " + adminToken})
		if resp.Code != http.StatusOK {
			t.Fatalf("expected %d, got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
		}
	}
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := bootstrap.AutoMigrate(db); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return db
}

func doRequest(t *testing.T, h http.Handler, method, path string, body []byte, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}

	req := httptest.NewRequest(method, path, r)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func mustJSON(t *testing.T, raw []byte, out any) {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(out); err != nil {
		t.Fatalf("decode json: %v; body=%s", err, string(raw))
	}
}

func jsonHeaders() map[string]string {
	return map[string]string{"Content-Type": "application/json", "Accept": "application/json"}
}

func withAuth(h map[string]string, token string) map[string]string {
	out := map[string]string{}
	for k, v := range h {
		out[k] = v
	}
	out["Authorization"] = "Bearer " + token
	return out
}

func mustUintFromJSONNumber(t *testing.T, v any) uint {
	t.Helper()
	switch x := v.(type) {
	case json.Number:
		u64, err := strconv.ParseUint(x.String(), 10, 64)
		if err != nil {
			t.Fatalf("parse uint from json.Number %q: %v", x.String(), err)
		}
		return uint(u64)
	case float64:
		if x < 0 {
			return 0
		}
		return uint(x)
	case int:
		if x < 0 {
			return 0
		}
		return uint(x)
	case uint:
		return x
	default:
		t.Fatalf("unexpected number type %T (%v)", v, v)
		return 0
	}
}
