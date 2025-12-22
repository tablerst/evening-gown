package public

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"evening-gown/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductsHandler struct {
	db *gorm.DB
}

func NewProductsHandler(db *gorm.DB) *ProductsHandler {
	return &ProductsHandler{db: db}
}

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

	q := h.db.WithContext(c.Request.Context()).Model(&model.Product{}).
		Where("published_at IS NOT NULL").
		Where("deleted_at IS NULL")

	if season := strings.TrimSpace(c.Query("season")); season != "" {
		q = q.Where("season = ?", season)
	}
	if category := strings.TrimSpace(c.Query("category")); category != "" {
		q = q.Where("category = ?", category)
	}
	if availability := strings.TrimSpace(c.Query("availability")); availability != "" {
		q = q.Where("availability = ?", availability)
	}
	if isNew := strings.TrimSpace(c.Query("is_new")); isNew != "" {
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

	var total int64
	if err := q.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	var products []model.Product
	if err := q.Select("id, style_no, season, category, availability, cover_image_url, hover_image_url, is_new, new_rank").
		Order("is_new desc, new_rank desc, id desc").Limit(limit).Offset(offset).Find(&products).Error; err != nil {
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
			CoverImage:   p.CoverImageURL,
			HoverImage:   p.HoverImageURL,
			IsNew:        p.IsNew,
			PriceMode:    "negotiable",
			PriceText:    "面议",
		})
	}

	c.JSON(http.StatusOK, gin.H{"total": total, "items": items})
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
		Where("published_at IS NOT NULL").
		Where("deleted_at IS NULL").
		First(&p, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           p.ID,
		"slug":         p.Slug,
		"styleNo":      p.StyleNo,
		"season":       p.Season,
		"category":     p.Category,
		"availability": p.Availability,
		"coverImage":   p.CoverImageURL,
		"hoverImage":   p.HoverImageURL,
		"isNew":        p.IsNew,
		"priceMode":    "negotiable",
		"priceText":    "面议",
		"detail":       jsonOrNull(p.DetailJSON),
	})
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
