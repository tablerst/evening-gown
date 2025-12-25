package admin

import (
	"net/http/httptest"
	"testing"
	"time"

	"evening-gown/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, err := db.DB()
	if err == nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}
	if err := db.AutoMigrate(&model.Product{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestAdminAssets_isKnownProductAsset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := openTestDB(t)
	h := &AssetsHandler{db: db}

	key := "products/1001/cover/2025/12/24/abc.webp"
	p := model.Product{
		Slug:          "style-1001",
		StyleNo:       "1001",
		Season:        "ss25",
		Category:      "gown",
		Availability:  "in_stock",
		CoverImageKey: key,
		HoverImageKey: "",
		CoverImageURL: "",
		HoverImageURL: "",
		PriceMode:     "negotiable",
		DetailJSON:    nil,
	}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/api/v1/admin/assets/"+key, nil)

	ok, err := h.isKnownProductAsset(c, key)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if !ok {
		t.Fatalf("expected ok=true")
	}

	// Wrong kind
	ok, err = h.isKnownProductAsset(c, "products/1001/other/2025/12/24/abc.webp")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if ok {
		t.Fatalf("expected ok=false")
	}

	// Not referenced in DB
	ok, err = h.isKnownProductAsset(c, "products/1001/cover/2025/12/24/other.webp")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if ok {
		t.Fatalf("expected ok=false")
	}

	// Deleted product should not authorize
	now := time.Now().UTC()
	if err := db.Model(&model.Product{}).Where("id = ?", p.ID).Update("deleted_at", &now).Error; err != nil {
		t.Fatalf("delete product: %v", err)
	}
	ok, err = h.isKnownProductAsset(c, key)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if ok {
		t.Fatalf("expected ok=false")
	}
}
