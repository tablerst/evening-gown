package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwtauth "evening-gown/internal/auth"
	"evening-gown/internal/bootstrap"
	"evening-gown/internal/config"
	"evening-gown/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAdminAuth_ErrorsAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtCfg := config.JWTConfig{Secret: "test-secret", Issuer: "evening-gown", ExpiresIn: time.Hour}
	jwtSvc, err := jwtauth.New(jwtCfg)
	if err != nil {
		t.Fatalf("jwt new: %v", err)
	}

	db := openTestDB(t)

	t.Run("unavailable when deps nil", func(t *testing.T) {
		r := gin.New()
		r.GET("/p", AdminAuth(nil, nil), func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/p", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected %d got %d", http.StatusServiceUnavailable, w.Code)
		}
	})

	t.Run("missing token", func(t *testing.T) {
		r := gin.New()
		r.GET("/p", AdminAuth(db, jwtSvc), func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/p", nil)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected %d got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("subject not numeric", func(t *testing.T) {
		// Token is valid JWT but subject is not a uint.
		tok, _, err := jwtSvc.IssueToken("abc")
		if err != nil {
			t.Fatalf("issue token: %v", err)
		}

		r := gin.New()
		r.GET("/p", AdminAuth(db, jwtSvc), func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/p", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected %d got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		tok, _, err := jwtSvc.IssueToken("999")
		if err != nil {
			t.Fatalf("issue token: %v", err)
		}

		r := gin.New()
		r.GET("/p", AdminAuth(db, jwtSvc), func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/p", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected %d got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("forbidden when not admin", func(t *testing.T) {
		user := model.User{Email: "u@example.com", PasswordHash: "x", Role: "staff", Status: "active"}
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}

		tok, _, err := jwtSvc.IssueToken("1")
		if err != nil {
			t.Fatalf("issue token: %v", err)
		}

		r := gin.New()
		r.GET("/p", AdminAuth(db, jwtSvc), func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/p", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected %d got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("forbidden when disabled", func(t *testing.T) {
		user := model.User{Email: "a@example.com", PasswordHash: "x", Role: "admin", Status: "disabled"}
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}

		tok, _, err := jwtSvc.IssueToken("2")
		if err != nil {
			t.Fatalf("issue token: %v", err)
		}

		r := gin.New()
		r.GET("/p", AdminAuth(db, jwtSvc), func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/p", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected %d got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("success sets user in context", func(t *testing.T) {
		user := model.User{Email: "admin@example.com", PasswordHash: "x", Role: "admin", Status: "active"}
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}

		tok, _, err := jwtSvc.IssueToken("3")
		if err != nil {
			t.Fatalf("issue token: %v", err)
		}

		r := gin.New()
		r.GET("/p", AdminAuth(db, jwtSvc), func(c *gin.Context) {
			if _, ok := c.Get(ContextUserKey); !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "missing user"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/p", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected %d got %d", http.StatusOK, w.Code)
		}
	})
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
