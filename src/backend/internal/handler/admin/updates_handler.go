package admin

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

type updateUpsertRequest struct {
	Type   string `json:"type"`   // company|industry
	Status string `json:"status"` // draft|published|archived
	Tag    string `json:"tag"`
	Title  string `json:"title" binding:"required"`
	Summary string `json:"summary"`
	Body   string `json:"body"`
	RefCode string `json:"ref"`
	PinnedRank *int `json:"pinnedRank"`
}

func (h *UpdatesHandler) List(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	q := h.db.WithContext(c.Request.Context()).Model(&model.UpdatePost{})
	if t := strings.TrimSpace(c.Query("type")); t != "" {
		q = q.Where("type = ?", t)
	}
	if st := strings.TrimSpace(c.Query("status")); st != "" {
		q = q.Where("status = ?", st)
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

	var items []model.UpdatePost
	if err := q.Order("pinned_rank desc, published_at desc, id desc").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": total, "items": items})
}

func (h *UpdatesHandler) Create(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	var req updateUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	typeVal := strings.TrimSpace(req.Type)
	if typeVal == "" {
		typeVal = "company"
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "draft"
	}

	post := model.UpdatePost{
		Type:     typeVal,
		Status:   status,
		Tag:      strings.TrimSpace(req.Tag),
		Title:    strings.TrimSpace(req.Title),
		Summary:  strings.TrimSpace(req.Summary),
		Body:     strings.TrimSpace(req.Body),
		RefCode:  strings.TrimSpace(req.RefCode),
		PinnedRank: 0,
	}
	if req.PinnedRank != nil {
		post.PinnedRank = *req.PinnedRank
	}
	if status == "published" {
		now := time.Now().UTC()
		post.PublishedAt = &now
	}

	if err := h.db.WithContext(c.Request.Context()).Create(&post).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post)
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

	var post model.UpdatePost
	if err := h.db.WithContext(c.Request.Context()).First(&post, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, post)
}

func (h *UpdatesHandler) Update(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]any{}
	if s := strings.TrimSpace(req.Type); s != "" {
		updates["type"] = s
	}
	if s := strings.TrimSpace(req.Status); s != "" {
		updates["status"] = s
		if s == "published" {
			now := time.Now().UTC()
			updates["published_at"] = &now
		}
		if s == "draft" {
			updates["published_at"] = nil
		}
	}
	if s := strings.TrimSpace(req.Tag); s != "" {
		updates["tag"] = s
	}
	if s := strings.TrimSpace(req.Title); s != "" {
		updates["title"] = s
	}
	if s := strings.TrimSpace(req.Summary); s != "" {
		updates["summary"] = s
	}
	if s := strings.TrimSpace(req.Body); s != "" {
		updates["body"] = s
	}
	if s := strings.TrimSpace(req.RefCode); s != "" {
		updates["ref_code"] = s
	}
	if req.PinnedRank != nil {
		updates["pinned_rank"] = *req.PinnedRank
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no updates"})
		return
	}

	if err := h.db.WithContext(c.Request.Context()).Model(&model.UpdatePost{}).Where("id = ?", uint(id)).Updates(updates).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.Get(c)
}

func (h *UpdatesHandler) Publish(c *gin.Context) {
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
	if err := h.db.WithContext(c.Request.Context()).Model(&model.UpdatePost{}).Where("id = ?", uint(id)).Updates(map[string]any{
		"status":       "published",
		"published_at": &now,
	}).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.Get(c)
}
