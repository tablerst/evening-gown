package public

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"evening-gown/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpdatesHandler struct {
	db *gorm.DB
}

func NewUpdatesHandler(db *gorm.DB) *UpdatesHandler {
	return &UpdatesHandler{db: db}
}

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

	// Only show company updates for now.
	q := h.db.WithContext(c.Request.Context()).Model(&model.UpdatePost{}).
		Where("type = ?", "company").
		Where("status = ?", "published")

	var total int64
	if err := q.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	var posts []model.UpdatePost
	if err := q.Order("pinned_rank desc, published_at desc, id desc").Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
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

	c.JSON(http.StatusOK, gin.H{"total": total, "items": items})
}

func (h *UpdatesHandler) Get(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var p model.UpdatePost
	if err := h.db.WithContext(c.Request.Context()).
		Where("type = ?", "company").
		Where("status = ?", "published").
		First(&p, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	date := ""
	if p.PublishedAt != nil {
		date = p.PublishedAt.UTC().Format(time.RFC3339)
	}
	c.JSON(http.StatusOK, gin.H{
		"id":    p.ID,
		"type":  p.Type,
		"date":  date,
		"tag":   p.Tag,
		"title": p.Title,
		"body":  p.Body,
		"ref":   p.RefCode,
	})
}

func firstNonEmpty(a, b string) string {
	a = strings.TrimSpace(a)
	if a != "" {
		return a
	}
	return strings.TrimSpace(b)
}
