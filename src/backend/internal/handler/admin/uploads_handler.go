package admin

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"evening-gown/internal/config"
	"evening-gown/internal/model"
	"evening-gown/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type UploadsHandler struct {
	minioClient *minio.Client
	minioCfg    config.MinioConfig
	maxBytes    int64
}

func NewUploadsHandler(minioClient *minio.Client, minioCfg config.MinioConfig, uploadCfg config.UploadConfig) *UploadsHandler {
	maxBytes := uploadCfg.MaxImageUploadBytes
	if maxBytes <= 0 {
		maxBytes = 1048576
	}
	return &UploadsHandler{minioClient: minioClient, minioCfg: minioCfg, maxBytes: maxBytes}
}

// UploadImage accepts an already-compressed webp image and uploads it to MinIO.
//
// Form fields:
// - file: image/webp
// - kind: cover|hover
// - styleNo: int
func (h *UploadsHandler) UploadImage(c *gin.Context) {
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

	// Apply request body limit before parsing multipart.
	maxBody := h.maxBytes + 64*1024 // allow some multipart overhead
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBody)

	kind := strings.TrimSpace(c.PostForm("kind"))
	if kind != "cover" && kind != "hover" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid kind"})
		return
	}

	styleNoRaw := strings.TrimSpace(c.PostForm("styleNo"))
	styleNo, err := model.NormalizeStyleNo(styleNoRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid styleNo"})
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	if fh.Size <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file"})
		return
	}
	if fh.Size > h.maxBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error":    "file too large",
			"maxBytes": h.maxBytes,
		})
		return
	}

	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read file"})
		return
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	headCT := strings.TrimSpace(fh.Header.Get("Content-Type"))
	detectCT := ""
	if n > 0 {
		detectCT = http.DetectContentType(buf[:n])
	}
	if headCT != "image/webp" && detectCT != "image/webp" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "only image/webp is accepted",
			"contentType": headCT,
		})
		return
	}

	ctx := c.Request.Context()
	now := time.Now().UTC()
	objectKey := fmt.Sprintf(
		"products/%s/%s/%04d/%02d/%02d/%s.webp",
		styleNo,
		kind,
		now.Year(),
		now.Month(),
		now.Day(),
		uuid.NewString(),
	)

	r := io.MultiReader(bytes.NewReader(buf[:n]), f)
	if err := storage.PutObject(ctx, h.minioClient, h.minioCfg, objectKey, r, fh.Size, "image/webp"); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	assetURL := "/api/v1/assets/" + strings.TrimPrefix(objectKey, "/")

	c.JSON(http.StatusOK, gin.H{
		"url":         assetURL,
		"objectKey":   objectKey,
		"contentType": "image/webp",
		"size":        fh.Size,
	})
}
