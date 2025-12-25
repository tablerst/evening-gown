package admin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"evening-gown/internal/cache"
	"evening-gown/internal/logging"
	"evening-gown/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductsHandler struct {
	db    *gorm.DB
	cache *cache.PublicCache
}

func NewProductsHandler(db *gorm.DB, publicCache *cache.PublicCache) *ProductsHandler {
	return &ProductsHandler{db: db, cache: publicCache}
}


type productCreateRequest struct {
	Slug         string `json:"slug"`
	StyleNo      int    `json:"styleNo" binding:"required"`
	Season       string `json:"season" binding:"required"`
	Category     string `json:"category" binding:"required"`
	Availability string `json:"availability" binding:"required"`
	IsNew        *bool  `json:"isNew"`
	NewRank      *int   `json:"newRank"`

	CoverImageURL string `json:"coverImage"`
	CoverImageKey string `json:"coverImageKey"`
	HoverImageURL string `json:"hoverImage"`
	HoverImageKey string `json:"hoverImageKey"`

	Detail json.RawMessage `json:"detail"`
}

type productUpdateRequest struct {
	Slug         *string `json:"slug"`
	StyleNo      *int    `json:"styleNo"`
	Season       *string `json:"season"`
	Category     *string `json:"category"`
	Availability *string `json:"availability"`
	IsNew        *bool   `json:"isNew"`
	NewRank      *int    `json:"newRank"`

	CoverImageURL *string `json:"coverImage"`
	CoverImageKey *string `json:"coverImageKey"`
	HoverImageURL *string `json:"hoverImage"`
	HoverImageKey *string `json:"hoverImageKey"`

	Detail *json.RawMessage `json:"detail"`
}

func (h *ProductsHandler) List(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	q := h.db.WithContext(c.Request.Context()).Model(&model.Product{}).
		Where("deleted_at IS NULL")

	if status := strings.TrimSpace(c.Query("status")); status != "" {
		if status == "published" {
			q = q.Where("published_at IS NOT NULL")
		} else if status == "draft" {
			q = q.Where("published_at IS NULL")
		}
	}
	if isNew := strings.TrimSpace(c.Query("is_new")); isNew != "" {
		if isNew == "true" {
			q = q.Where("is_new = true")
		} else if isNew == "false" {
			q = q.Where("is_new = false")
		}
	}
	if season := strings.TrimSpace(c.Query("season")); season != "" {
		q = q.Where("season = ?", season)
	}
	if category := strings.TrimSpace(c.Query("category")); category != "" {
		q = q.Where("category = ?", category)
	}

	limit := parseIntQuery(c, "limit", 50)
	offset := parseIntQuery(c, "offset", 0)
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "admin products query count failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	var items []model.Product
	if err := q.Order("is_new desc, new_rank desc, id desc").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "admin products query list failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"items": items,
	})
}

func (h *ProductsHandler) Create(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	var req productCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = "style-" + strconv.Itoa(req.StyleNo)
	}

	isNew := false
	if req.IsNew != nil {
		isNew = *req.IsNew
	}
	newRank := 0
	if req.NewRank != nil {
		newRank = *req.NewRank
	}

	p := model.Product{
		Slug:          slug,
		StyleNo:       req.StyleNo,
		Season:        req.Season,
		Category:      req.Category,
		Availability:  req.Availability,
		IsNew:         isNew,
		NewRank:       newRank,
		CoverImageURL: strings.TrimSpace(req.CoverImageURL),
		CoverImageKey: strings.TrimSpace(req.CoverImageKey),
		HoverImageURL: strings.TrimSpace(req.HoverImageURL),
		HoverImageKey: strings.TrimSpace(req.HoverImageKey),
		PriceMode:     "negotiable",
		DetailJSON:    req.Detail,
	}

	if err := h.db.WithContext(c.Request.Context()).Create(&p).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, p)
}

func (h *ProductsHandler) Get(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var p model.Product
	if err := h.db.WithContext(c.Request.Context()).
		Where("deleted_at IS NULL").
		First(&p, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *ProductsHandler) Update(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	ctx := c.Request.Context()

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var before model.Product
	if err := h.db.WithContext(ctx).
		Where("id = ?", uint(id)).
		Where("deleted_at IS NULL").
		First(&before).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	wasPublished := before.PublishedAt != nil

	var req productUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]any{}
	if req.Slug != nil {
		if s := strings.TrimSpace(*req.Slug); s != "" {
			updates["slug"] = s
		}
	}
	if req.StyleNo != nil {
		if *req.StyleNo != 0 {
			updates["style_no"] = *req.StyleNo
		}
	}
	if req.Season != nil {
		if s := strings.TrimSpace(*req.Season); s != "" {
			updates["season"] = s
		}
	}
	if req.Category != nil {
		if s := strings.TrimSpace(*req.Category); s != "" {
			updates["category"] = s
		}
	}
	if req.Availability != nil {
		if s := strings.TrimSpace(*req.Availability); s != "" {
			updates["availability"] = s
		}
	}
	if req.IsNew != nil {
		updates["is_new"] = *req.IsNew
	}
	if req.NewRank != nil {
		updates["new_rank"] = *req.NewRank
	}
	if req.CoverImageURL != nil {
		updates["cover_image_url"] = strings.TrimSpace(*req.CoverImageURL)
	}
	if req.CoverImageKey != nil {
		updates["cover_image_key"] = strings.TrimSpace(*req.CoverImageKey)
	}
	if req.HoverImageURL != nil {
		updates["hover_image_url"] = strings.TrimSpace(*req.HoverImageURL)
	}
	if req.HoverImageKey != nil {
		updates["hover_image_key"] = strings.TrimSpace(*req.HoverImageKey)
	}
	if req.Detail != nil {
		updates["detail_json"] = *req.Detail
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no updates"})
		return
	}

	if err := h.db.WithContext(ctx).Model(&model.Product{}).
		Where("id = ?", uint(id)).
		Where("deleted_at IS NULL").
		Updates(updates).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if wasPublished && h.cache != nil {
		_, _ = h.cache.BumpProductsVersion(ctx)
	}

	h.Get(c)
}

func (h *ProductsHandler) Publish(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	now := time.Now().UTC()
	ctx := c.Request.Context()
	res := h.db.WithContext(ctx).Model(&model.Product{}).
		Where("id = ?", uint(id)).
		Where("deleted_at IS NULL").
		Update("published_at", &now)
	if res.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.Error.Error()})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if h.cache != nil {
		_, _ = h.cache.BumpProductsVersion(ctx)
	}

	h.Get(c)
}

func (h *ProductsHandler) Unpublish(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx := c.Request.Context()
	res := h.db.WithContext(ctx).Model(&model.Product{}).
		Where("id = ?", uint(id)).
		Where("deleted_at IS NULL").
		Update("published_at", nil)
	if res.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.Error.Error()})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if h.cache != nil {
		_, _ = h.cache.BumpProductsVersion(ctx)
	}

	h.Get(c)
}

func parseIntQuery(c *gin.Context, key string, fallback int) int {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func (h *ProductsHandler) Delete(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx := c.Request.Context()
	var before model.Product
	if err := h.db.WithContext(ctx).
		Where("id = ?", uint(id)).
		Where("deleted_at IS NULL").
		First(&before).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	wasPublished := before.PublishedAt != nil

	now := time.Now().UTC()
	res := h.db.WithContext(ctx).Model(&model.Product{}).
		Where("id = ?", uint(id)).
		Where("deleted_at IS NULL").
		Update("deleted_at", &now)
	if res.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.Error.Error()})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if wasPublished && h.cache != nil {
		_, _ = h.cache.BumpProductsVersion(ctx)
	}

	c.Status(http.StatusNoContent)
}
