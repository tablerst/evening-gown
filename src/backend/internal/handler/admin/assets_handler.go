package admin

import (
	"net/http"
	"path"
	"strings"

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
}

func NewAssetsHandler(db *gorm.DB, minioClient *minio.Client, minioCfg config.MinioConfig) *AssetsHandler {
	return &AssetsHandler{db: db, minioClient: minioClient, minioCfg: minioCfg}
}

// Get streams an object from MinIO through the application for admin usage.
//
// Route: GET /api/v1/admin/assets/*key
//
// Notes:
// - Protected by AdminAuth middleware.
// - Intended for backoffice preview (draft products). Public website must use /api/v1/assets/*key.
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

	// Safety: currently only allow product assets.
	if !strings.HasPrefix(cleanKey, "products/") {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// Prevent "arbitrary object read" via admin assets endpoint:
	// only allow keys that are referenced by an existing (non-deleted) product.
	// This matches the admin UI workflow where draft previews load the key stored on the product.
	if ok, err := h.isKnownProductAsset(c, cleanKey); err != nil || !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

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

	headers := map[string]string{
		// Avoid long-term caching for draft assets.
		"Cache-Control": "private, max-age=60",
	}
	if strings.TrimSpace(stat.ETag) != "" {
		c.Header("ETag", stat.ETag)
	}

	c.DataFromReader(http.StatusOK, stat.Size, contentType, obj, headers)
}

func (h *AssetsHandler) isKnownProductAsset(c *gin.Context, objectKey string) (bool, error) {
	if h == nil || h.db == nil || c == nil {
		return false, nil
	}

	objectKey = strings.TrimSpace(strings.TrimPrefix(objectKey, "/"))
	if objectKey == "" {
		return false, nil
	}

	// Expected: products/{styleNo}/{kind}/...
	parts := strings.Split(objectKey, "/")
	if len(parts) < 3 {
		return false, nil
	}
	if parts[0] != "products" {
		return false, nil
	}

	styleNo, err := model.NormalizeStyleNo(strings.TrimSpace(parts[1]))
	if err != nil {
		return false, nil
	}

	kind := strings.TrimSpace(parts[2])
	if kind != "cover" && kind != "hover" {
		return false, nil
	}

	var cnt int64
	err = h.db.WithContext(c.Request.Context()).Model(&model.Product{}).
		Where("style_no = ?", styleNo).
		Where("deleted_at IS NULL").
		Where("(cover_image_key = ? OR hover_image_key = ?)", objectKey, objectKey).
		Limit(1).
		Count(&cnt).Error
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}
