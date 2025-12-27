package admin

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"evening-gown/internal/cache"
	"evening-gown/internal/logging"
	"evening-gown/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ContactsHandler struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewContactsHandler(db *gorm.DB) *ContactsHandler {
	return NewContactsHandlerWithRedis(db, nil)
}

func NewContactsHandlerWithRedis(db *gorm.DB, rdb *redis.Client) *ContactsHandler {
	return &ContactsHandler{db: db, rdb: rdb}
}

// UnreadCount returns the number of contact leads that are still "new".
//
// Query params:
// - force=true: recompute from DB and overwrite Redis counter.
func (h *ContactsHandler) UnreadCount(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	ctx := c.Request.Context()
	force := strings.EqualFold(strings.TrimSpace(c.Query("force")), "true")

	// Fast path: Redis counter (if present and not forcing).
	if h.rdb != nil && !force {
		val, err := h.rdb.Get(ctx, cache.AdminContactsNewCountKey).Int64()
		if err == nil && val >= 0 {
			c.JSON(http.StatusOK, gin.H{
				"count": val,
				"status": "new",
				"asOf": time.Now().UTC().Format(time.RFC3339),
				"source": "redis",
			})
			return
		}
		// On redis.Nil or parse error, fall back to DB recompute.
	}

	count, err := h.countNewLeads(ctx)
	if err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "admin contacts unread-count db failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	if h.rdb != nil {
		// Counter key is a durable value; no TTL.
		_ = h.rdb.Set(ctx, cache.AdminContactsNewCountKey, count, 0).Err()
	}

	c.JSON(http.StatusOK, gin.H{
		"count": count,
		"status": "new",
		"asOf": time.Now().UTC().Format(time.RFC3339),
		"source": "db",
	})
}

func (h *ContactsHandler) List(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	q := h.db.WithContext(c.Request.Context()).Model(&model.ContactLead{})
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
		logging.ErrorWithStack(logging.FromGin(c), "admin contacts query count failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	var items []model.ContactLead
	if err := q.Order("id desc").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "admin contacts query list failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": total, "items": items})
}

func (h *ContactsHandler) Get(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var lead model.ContactLead
	if err := h.db.WithContext(c.Request.Context()).First(&lead, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, lead)
}

type contactUpdateRequest struct {
	Status string `json:"status" binding:"required"` // new|contacted|closed
}

func (h *ContactsHandler) Update(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req contactUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	st := strings.TrimSpace(req.Status)
	if st != "new" && st != "contacted" && st != "closed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	ctx := c.Request.Context()

	var before model.ContactLead
	if err := h.db.WithContext(ctx).First(&before, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	if strings.TrimSpace(before.Status) == st {
		// No-op update; return existing record.
		c.JSON(http.StatusOK, before)
		return
	}

	if err := h.db.WithContext(ctx).Model(&model.ContactLead{}).Where("id = ?", uint(id)).Update("status", st).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Keep the Redis counter strongly consistent by applying a delta.
	if h.rdb != nil {
		delta := int64(0)
		if strings.TrimSpace(before.Status) == "new" && st != "new" {
			delta = -1
		} else if strings.TrimSpace(before.Status) != "new" && st == "new" {
			delta = 1
		}
		if delta != 0 {
			if err := h.applyNewLeadsDelta(ctx, delta); err != nil {
				// Best-effort: do not fail the request.
				logging.ErrorWithStack(logging.FromGin(c), "admin contacts unread-count delta failed", err)
			}
		}
	}

	var lead model.ContactLead
	if err := h.db.WithContext(ctx).First(&lead, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, lead)
}

func (h *ContactsHandler) Delete(c *gin.Context) {
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
	var before model.ContactLead
	if err := h.db.WithContext(ctx).First(&before, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	res := h.db.WithContext(ctx).Delete(&model.ContactLead{}, uint(id))
	if res.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": res.Error.Error()})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	if h.rdb != nil && strings.TrimSpace(before.Status) == "new" {
		if err := h.applyNewLeadsDelta(ctx, -1); err != nil {
			logging.ErrorWithStack(logging.FromGin(c), "admin contacts unread-count delta failed", err)
		}
	}

	c.Status(http.StatusNoContent)
}

func (h *ContactsHandler) countNewLeads(ctx context.Context) (int64, error) {
	var count int64
	err := h.db.WithContext(ctx).Model(&model.ContactLead{}).Where("status = ?", "new").Count(&count).Error
	return count, err
}

func (h *ContactsHandler) applyNewLeadsDelta(ctx context.Context, delta int64) error {
	if h == nil || h.rdb == nil || delta == 0 {
		return nil
	}

	// If the counter key is missing (e.g., Redis eviction), reconcile from DB.
	if exists, err := h.rdb.Exists(ctx, cache.AdminContactsNewCountKey).Result(); err == nil && exists == 0 {
		count, err := h.countNewLeads(ctx)
		if err != nil {
			return err
		}
		return h.rdb.Set(ctx, cache.AdminContactsNewCountKey, count, 0).Err()
	}

	val, err := h.rdb.IncrBy(ctx, cache.AdminContactsNewCountKey, delta).Result()
	if err != nil {
		// If counter update fails, callers will fall back to DB on next count read.
		return err
	}
	if val >= 0 {
		return nil
	}

	// Counter went negative (unexpected). Reconcile from DB.
	count, err := h.countNewLeads(ctx)
	if err != nil {
		return err
	}
	return h.rdb.Set(ctx, cache.AdminContactsNewCountKey, count, 0).Err()
}

