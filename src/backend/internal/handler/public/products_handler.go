package public

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

const (
	publicProductsListTTL   = 5 * time.Minute
	publicProductDetailTTL  = 30 * time.Minute
	publicProductNotFoundTTL = 30 * time.Second
)

type productListItem struct {
	ID           uint   `json:"id"`
	StyleNo      int    `json:"styleNo"`
	Season       string `json:"season"`
	Category     string `json:"category"`
	Availability string `json:"availability"`
	CoverImage   string `json:"coverImage"`
	HoverImage   string `json:"hoverImage"`
	IsNew        bool   `json:"isNew"`

	PriceMode string `json:"priceMode"`
	PriceText string `json:"priceText"`
}

func (h *ProductsHandler) List(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	ctx := c.Request.Context()

	q := h.db.WithContext(c.Request.Context()).Model(&model.Product{}).
		Where("published_at IS NOT NULL").
		Where("deleted_at IS NULL")

	season := strings.TrimSpace(c.Query("season"))
	category := strings.TrimSpace(c.Query("category"))
	availability := strings.TrimSpace(c.Query("availability"))
	isNew := strings.TrimSpace(c.Query("is_new"))

	if season != "" {
		q = q.Where("season = ?", season)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if availability != "" {
		q = q.Where("availability = ?", availability)
	}
	if isNew != "" {
		if isNew == "true" {
			q = q.Where("is_new = true")
		} else if isNew == "false" {
			q = q.Where("is_new = false")
		}
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

	// Cache-aside with versioned key: after admin writes bump the products version,
	// the next read will bypass old cache entries.
	var cacheKey string
	if h.cache != nil {
		ver := h.cache.ProductsVersion(ctx)
		cacheKey = h.cache.ProductsListKey(ver, season, category, availability, isNew, limit, offset)
		if b, hit, _ := h.cache.GetJSONBytes(ctx, cacheKey); hit {
			c.Data(http.StatusOK, "application/json; charset=utf-8", b)
			return
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "public products query count failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	var products []model.Product
	if err := q.Select("id, style_no, season, category, availability, cover_image_url, cover_image_key, hover_image_url, hover_image_key, is_new, new_rank").
		Order("is_new desc, new_rank desc, id desc").Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "public products query list failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	items := make([]productListItem, 0, len(products))
	for _, p := range products {
		items = append(items, productListItem{
			ID:           p.ID,
			StyleNo:      p.StyleNo,
			Season:       p.Season,
			Category:     p.Category,
			Availability: p.Availability,
			CoverImage:   pickPublicImageURL(p.CoverImageKey, p.CoverImageURL),
			HoverImage:   pickPublicImageURL(p.HoverImageKey, p.HoverImageURL),
			IsNew:        p.IsNew,
			PriceMode:    "negotiable",
			PriceText:    "面议",
		})
	}

	resp := gin.H{"total": total, "items": items}
	if h.cache != nil && cacheKey != "" {
		b, err := json.Marshal(resp)
		if err == nil {
			ttl := cache.TTLWithKeyJitter(publicProductsListTTL, cacheKey, 0.2)
			h.cache.SetJSONBytes(ctx, cacheKey, b, ttl)
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ProductsHandler) Get(c *gin.Context) {
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

	var cacheKey string
	if h.cache != nil {
		ver := h.cache.ProductsVersion(ctx)
		cacheKey = h.cache.ProductDetailKey(ver, uint(id))
		if b, hit, isNF := h.cache.GetJSONBytes(ctx, cacheKey); hit {
			if isNF {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.Data(http.StatusOK, "application/json; charset=utf-8", b)
			return
		}
	}

	var p model.Product
	if err := h.db.WithContext(c.Request.Context()).
		Where("published_at IS NOT NULL").
		Where("deleted_at IS NULL").
		First(&p, uint(id)).Error; err != nil {
		if h.cache != nil && cacheKey != "" {
			ttl := cache.TTLWithKeyJitter(publicProductNotFoundTTL, cacheKey, 0.2)
			h.cache.SetNotFound(ctx, cacheKey, ttl)
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	resp := gin.H{
		"id":           p.ID,
		"slug":         p.Slug,
		"styleNo":      p.StyleNo,
		"season":       p.Season,
		"category":     p.Category,
		"availability": p.Availability,
		"coverImage":   pickPublicImageURL(p.CoverImageKey, p.CoverImageURL),
		"hoverImage":   pickPublicImageURL(p.HoverImageKey, p.HoverImageURL),
		"isNew":        p.IsNew,
		"priceMode":    "negotiable",
		"priceText":    "面议",
		"detail":       jsonOrNull(p.DetailJSON),
	}

	if h.cache != nil && cacheKey != "" {
		b, err := json.Marshal(resp)
		if err == nil {
			ttl := cache.TTLWithKeyJitter(publicProductDetailTTL, cacheKey, 0.2)
			h.cache.SetJSONBytes(ctx, cacheKey, b, ttl)
		}
	}

	c.JSON(http.StatusOK, resp)
}

func pickPublicImageURL(objectKey string, legacyURL string) string {
	key := strings.TrimSpace(strings.TrimPrefix(objectKey, "/"))
	if key != "" {
		return "/api/v1/assets/" + key
	}
	return strings.TrimSpace(legacyURL)
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

func jsonOrNull(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil
	}
	return v
}
