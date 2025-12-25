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

type UpdatesHandler struct {
	db    *gorm.DB
	cache *cache.PublicCache
}

func NewUpdatesHandler(db *gorm.DB, publicCache *cache.PublicCache) *UpdatesHandler {
	return &UpdatesHandler{db: db, cache: publicCache}
}

const (
	publicUpdatesListTTL    = 5 * time.Minute
	publicUpdateDetailTTL   = 30 * time.Minute
	publicUpdateNotFoundTTL = 30 * time.Second
)

type updateItem struct {
	ID    uint   `json:"id"`
	Date  string `json:"date"`
	Tag   string `json:"tag"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Ref   string `json:"ref,omitempty"`
}

func (h *UpdatesHandler) List(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	ctx := c.Request.Context()

	limit := parseIntQuery(c, "limit", 3)
	offset := parseIntQuery(c, "offset", 0)
	if limit <= 0 {
		limit = 3
	}
	if limit > 50 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var cacheKey string
	if h.cache != nil {
		ver := h.cache.UpdatesVersion(ctx)
		cacheKey = h.cache.UpdatesListKey(ver, limit, offset)
		if b, hit, _ := h.cache.GetJSONBytes(ctx, cacheKey); hit {
			c.Data(http.StatusOK, "application/json; charset=utf-8", b)
			return
		}
	}

	// Only show company updates for now.
	q := h.db.WithContext(c.Request.Context()).Model(&model.UpdatePost{}).
		Where("type = ?", "company").
		Where("status = ?", "published").
		Where("deleted_at IS NULL")

	var total int64
	if err := q.Count(&total).Error; err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "public updates query count failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	var posts []model.UpdatePost
	if err := q.Order("pinned_rank desc, published_at desc, id desc").Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "public updates query list failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	items := make([]updateItem, 0, len(posts))
	for _, p := range posts {
		date := ""
		if p.PublishedAt != nil {
			date = p.PublishedAt.UTC().Format("2006-01-02")
		}
		items = append(items, updateItem{
			ID:    p.ID,
			Date:  date,
			Tag:   p.Tag,
			Title: p.Title,
			Body:  firstNonEmpty(p.Body, p.Summary),
			Ref:   p.RefCode,
		})
	}

	resp := gin.H{"total": total, "items": items}
	if h.cache != nil && cacheKey != "" {
		b, err := json.Marshal(resp)
		if err == nil {
			ttl := cache.TTLWithKeyJitter(publicUpdatesListTTL, cacheKey, 0.2)
			h.cache.SetJSONBytes(ctx, cacheKey, b, ttl)
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *UpdatesHandler) Get(c *gin.Context) {
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
		ver := h.cache.UpdatesVersion(ctx)
		cacheKey = h.cache.UpdateDetailKey(ver, uint(id))
		if b, hit, isNF := h.cache.GetJSONBytes(ctx, cacheKey); hit {
			if isNF {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.Data(http.StatusOK, "application/json; charset=utf-8", b)
			return
		}
	}

	var p model.UpdatePost
	if err := h.db.WithContext(c.Request.Context()).
		Where("type = ?", "company").
		Where("status = ?", "published").
		Where("deleted_at IS NULL").
		First(&p, uint(id)).Error; err != nil {
		if h.cache != nil && cacheKey != "" {
			ttl := cache.TTLWithKeyJitter(publicUpdateNotFoundTTL, cacheKey, 0.2)
			h.cache.SetNotFound(ctx, cacheKey, ttl)
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	date := ""
	if p.PublishedAt != nil {
		date = p.PublishedAt.UTC().Format(time.RFC3339)
	}
	resp := gin.H{
		"id":    p.ID,
		"type":  p.Type,
		"date":  date,
		"tag":   p.Tag,
		"title": p.Title,
		"body":  p.Body,
		"ref":   p.RefCode,
	}

	if h.cache != nil && cacheKey != "" {
		b, err := json.Marshal(resp)
		if err == nil {
			ttl := cache.TTLWithKeyJitter(publicUpdateDetailTTL, cacheKey, 0.2)
			h.cache.SetJSONBytes(ctx, cacheKey, b, ttl)
		}
	}

	c.JSON(http.StatusOK, resp)
}

func firstNonEmpty(a, b string) string {
	a = strings.TrimSpace(a)
	if a != "" {
		return a
	}
	return strings.TrimSpace(b)
}
