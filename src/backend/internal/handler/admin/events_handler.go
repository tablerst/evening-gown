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

type EventsHandler struct {
	db *gorm.DB
}

func NewEventsHandler(db *gorm.DB) *EventsHandler {
	return &EventsHandler{db: db}
}

func (h *EventsHandler) List(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	q := h.db.WithContext(c.Request.Context()).Model(&model.Event{})

	if et := strings.TrimSpace(c.Query("event_type")); et != "" {
		q = q.Where("event_type = ?", et)
	}
	if pid := strings.TrimSpace(c.Query("product_id")); pid != "" {
		if v, err := strconv.ParseUint(pid, 10, 64); err == nil && v > 0 {
			q = q.Where("product_id = ?", uint(v))
		}
	}

	if from := strings.TrimSpace(c.Query("from")); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			q = q.Where("occurred_at >= ?", t)
		}
	}
	if to := strings.TrimSpace(c.Query("to")); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			q = q.Where("occurred_at <= ?", t)
		}
	}

	limit := parseIntQuery(c, "limit", 100)
	offset := parseIntQuery(c, "offset", 0)
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	var items []model.Event
	if err := q.Order("occurred_at desc, id desc").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": total, "items": items})
}

func (h *EventsHandler) Get(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var e model.Event
	if err := h.db.WithContext(c.Request.Context()).First(&e, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, e)
}

func (h *EventsHandler) Delete(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res := h.db.WithContext(c.Request.Context()).Delete(&model.Event{}, uint(id))
	if res.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.Error.Error()})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
