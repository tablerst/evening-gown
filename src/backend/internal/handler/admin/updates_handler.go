package admin

import (
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

func isPublicCompanyUpdate(p model.UpdatePost) bool {
	return strings.TrimSpace(p.Type) == "company" && strings.TrimSpace(p.Status) == "published" && p.DeletedAt == nil
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

type updateUpdateRequest struct {
	Type   *string `json:"type"`   // company|industry
	Status *string `json:"status"` // draft|published|archived
	Tag    *string `json:"tag"`
	Title  *string `json:"title"`
	Summary *string `json:"summary"`
	Body   *string `json:"body"`
	RefCode *string `json:"ref"`
	PinnedRank *int `json:"pinnedRank"`
}

func (h *UpdatesHandler) List(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	q := h.db.WithContext(c.Request.Context()).Model(&model.UpdatePost{}).
		Where("deleted_at IS NULL")
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
		logging.ErrorWithStack(logging.FromGin(c), "admin updates query count failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	var items []model.UpdatePost
	if err := q.Order("pinned_rank desc, published_at desc, id desc").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "admin updates query list failed", err)
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
	if h.cache != nil && isPublicCompanyUpdate(post) {
		_, _ = h.cache.BumpUpdatesVersion(c.Request.Context())
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
	if err := h.db.WithContext(c.Request.Context()).
		Where("deleted_at IS NULL").
		First(&post, uint(id)).Error; err != nil {
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

	ctx := c.Request.Context()
	var before model.UpdatePost
	if err := h.db.WithContext(ctx).
		Where("id = ?", uint(id)).
		Where("deleted_at IS NULL").
		First(&before).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	var req updateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]any{}
	if req.Type != nil {
		if s := strings.TrimSpace(*req.Type); s != "" {
			updates["type"] = s
		}
	}
	if req.Status != nil {
		if s := strings.TrimSpace(*req.Status); s != "" {
			updates["status"] = s
			if s == "published" {
				now := time.Now().UTC()
				updates["published_at"] = &now
			}
			if s == "draft" {
				updates["published_at"] = nil
			}
		}
	}
	if req.Tag != nil {
		if s := strings.TrimSpace(*req.Tag); s != "" {
			updates["tag"] = s
		}
	}
	if req.Title != nil {
		if s := strings.TrimSpace(*req.Title); s != "" {
			updates["title"] = s
		}
	}
	if req.Summary != nil {
		if s := strings.TrimSpace(*req.Summary); s != "" {
			updates["summary"] = s
		}
	}
	if req.Body != nil {
		if s := strings.TrimSpace(*req.Body); s != "" {
			updates["body"] = s
		}
	}
	if req.RefCode != nil {
		if s := strings.TrimSpace(*req.RefCode); s != "" {
			updates["ref_code"] = s
		}
	}
	if req.PinnedRank != nil {
		updates["pinned_rank"] = *req.PinnedRank
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no updates"})
		return
	}

	if err := h.db.WithContext(ctx).Model(&model.UpdatePost{}).
		Where("id = ?", uint(id)).
		Where("deleted_at IS NULL").
		Updates(updates).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var after model.UpdatePost
	if err := h.db.WithContext(ctx).
		Where("id = ?", uint(id)).
		Where("deleted_at IS NULL").
		First(&after).Error; err == nil {
		if h.cache != nil && (isPublicCompanyUpdate(before) || isPublicCompanyUpdate(after)) {
			_, _ = h.cache.BumpUpdatesVersion(ctx)
		}
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

	ctx := c.Request.Context()
	var before model.UpdatePost
	if err := h.db.WithContext(ctx).
		Where("id = ?", uint(id)).
		Where("deleted_at IS NULL").
		First(&before).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	now := time.Now().UTC()
	res := h.db.WithContext(ctx).Model(&model.UpdatePost{}).
		Where("id = ?", uint(id)).
		Where("deleted_at IS NULL").
		Updates(map[string]any{
			"status":       "published",
			"published_at": &now,
		})
	if res.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.Error.Error()})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if h.cache != nil && strings.TrimSpace(before.Type) == "company" {
		_, _ = h.cache.BumpUpdatesVersion(ctx)
	}

	h.Get(c)
}

func (h *UpdatesHandler) Unpublish(c *gin.Context) {
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
	var before model.UpdatePost
	if err := h.db.WithContext(ctx).
		Where("id = ?", uint(id)).
		Where("deleted_at IS NULL").
		First(&before).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	res := h.db.WithContext(ctx).Model(&model.UpdatePost{}).
		Where("id = ?", uint(id)).
		Where("deleted_at IS NULL").
		Updates(map[string]any{
			"status":       "draft",
			"published_at": nil,
		})
	if res.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.Error.Error()})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if h.cache != nil && isPublicCompanyUpdate(before) {
		_, _ = h.cache.BumpUpdatesVersion(ctx)
	}

	h.Get(c)
}

func (h *UpdatesHandler) Delete(c *gin.Context) {
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
	var before model.UpdatePost
	if err := h.db.WithContext(ctx).
		Where("id = ?", uint(id)).
		Where("deleted_at IS NULL").
		First(&before).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	now := time.Now().UTC()
	res := h.db.WithContext(ctx).Model(&model.UpdatePost{}).
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
	if h.cache != nil && isPublicCompanyUpdate(before) {
		_, _ = h.cache.BumpUpdatesVersion(ctx)
	}

	c.Status(http.StatusNoContent)
}
