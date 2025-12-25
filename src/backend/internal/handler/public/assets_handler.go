package public

import (
	"net/http"
	"path"
	"strings"
	"time"

	"evening-gown/internal/cache"
	"evening-gown/internal/config"
	"evening-gown/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type AssetsHandler struct {
	db         *gorm.DB
	minioClient *minio.Client
	minioCfg    config.MinioConfig
	cache       *cache.PublicCache
}

func NewAssetsHandler(db *gorm.DB, minioClient *minio.Client, minioCfg config.MinioConfig, publicCache *cache.PublicCache) *AssetsHandler {
	return &AssetsHandler{db: db, minioClient: minioClient, minioCfg: minioCfg, cache: publicCache}
}

const publicAssetAllowTTL = 15 * time.Minute

// Get streams an object from MinIO through the application.
//
// Route: GET /api/v1/assets/*key
//
// Notes:
// - Intended for public website consumption (published products).
// - Keeps MinIO buckets private; browsers never talk to MinIO directly.
func (h *AssetsHandler) Get(c *gin.Context) {
	if h == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}
	if h.minioClient == nil || strings.TrimSpace(h.minioCfg.Endpoint) == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "minio disabled"})
		return
	}
	if strings.TrimSpace(h.minioCfg.Bucket) == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "minio bucket not configured"})
		return
	}

	key := strings.TrimPrefix(c.Param("key"), "/")
	key = strings.TrimSpace(key)
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key"})
		return
	}
	if strings.Contains(key, "\\") || strings.Contains(key, "\x00") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key"})
		return
	}

	cleaned := path.Clean("/" + key)
	if strings.HasPrefix(cleaned, "/..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key"})
		return
	}
	cleanKey := strings.TrimPrefix(cleaned, "/")
	if cleanKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key"})
		return
	}

	// Optional safety: only allow product assets for now.
	if !strings.HasPrefix(cleanKey, "products/") {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// Prevent unauthorized reads of draft/backoffice-managed images.
	// Only allow assets that are referenced by a published product.
	// NOTE: Admin backoffice can fetch draft assets via /api/v1/admin/assets/*key.
	if h.db != nil {
		ctx := c.Request.Context()
		productsVer := int64(0)
		if h.cache != nil {
			productsVer = h.cache.ProductsVersion(ctx)
			allowKey := h.cache.AssetAllowKey(productsVer, cleanKey)
			if v, hit := h.cache.BoolFromCache(ctx, allowKey); hit {
				if !v {
					c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
					return
				}
				goto allowed
			}
			// Cache miss: fallthrough to DB check.
			ok, err := h.isPublishedProductAsset(c, cleanKey)
			if err != nil || !ok {
				h.cache.SetBool(ctx, allowKey, false, cache.TTLWithKeyJitter(publicAssetAllowTTL, allowKey, 0.2))
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			h.cache.SetBool(ctx, allowKey, true, cache.TTLWithKeyJitter(publicAssetAllowTTL, allowKey, 0.2))
			goto allowed
		}

		ok, err := h.isPublishedProductAsset(c, cleanKey)
		if err != nil || !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
	}

allowed:

	ctx := c.Request.Context()

	stat, err := h.minioClient.StatObject(ctx, h.minioCfg.Bucket, cleanKey, minio.StatObjectOptions{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	obj, err := h.minioClient.GetObject(ctx, h.minioCfg.Bucket, cleanKey, minio.GetObjectOptions{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	defer obj.Close()

	contentType := strings.TrimSpace(stat.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Cache aggressively: object keys are content-addressed-ish (include uuid/date),
	// so updates generate new keys and won't break caches.
	headers := map[string]string{
		"Cache-Control": "public, max-age=31536000, immutable",
	}
	if strings.TrimSpace(stat.ETag) != "" {
		c.Header("ETag", stat.ETag)
	}

	c.DataFromReader(http.StatusOK, stat.Size, contentType, obj, headers)
}

func (h *AssetsHandler) isPublishedProductAsset(c *gin.Context, objectKey string) (bool, error) {
	// products/{styleNo}/...
	parts := strings.Split(objectKey, "/")
	if len(parts) < 2 {
		return false, nil
	}
	styleNo, err := model.NormalizeStyleNo(strings.TrimSpace(parts[1]))
	if err != nil {
		return false, nil
	}

	objectKey = strings.TrimSpace(strings.TrimPrefix(objectKey, "/"))
	if objectKey == "" {
		return false, nil
	}

	var cnt int64
	err = h.db.WithContext(c.Request.Context()).Model(&model.Product{}).
		Where("style_no = ?", styleNo).
		Where("published_at IS NOT NULL").
		Where("deleted_at IS NULL").
		Where("(cover_image_key = ? OR hover_image_key = ?)", objectKey, objectKey).
		Limit(1).
		Count(&cnt).Error
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}
