package public

import (
	"net/http"
	"strings"

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

type contactCreateRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Wechat  string `json:"wechat"`
	Message string `json:"message"`

	SourcePage string `json:"source_page"`
	UTMSource  string `json:"utm_source"`
	UTMMedium  string `json:"utm_medium"`
	UTMCampaign string `json:"utm_campaign"`
	UTMContent  string `json:"utm_content"`
	UTMTerm     string `json:"utm_term"`
}

func (h *ContactsHandler) Create(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	var req contactCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	phone := strings.TrimSpace(req.Phone)
	wechat := strings.TrimSpace(req.Wechat)
	if phone == "" && wechat == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone or wechat is required"})
		return
	}

	lead := model.ContactLead{
		Name:       strings.TrimSpace(req.Name),
		Phone:      phone,
		Wechat:     wechat,
		Message:    strings.TrimSpace(req.Message),
		SourcePage: strings.TrimSpace(req.SourcePage),
		UTMSource:  strings.TrimSpace(req.UTMSource),
		UTMMedium:  strings.TrimSpace(req.UTMMedium),
		UTMCampaign: strings.TrimSpace(req.UTMCampaign),
		UTMContent:  strings.TrimSpace(req.UTMContent),
		UTMTerm:     strings.TrimSpace(req.UTMTerm),
		Status:     "new",
	}

	if err := h.db.WithContext(c.Request.Context()).Create(&lead).Error; err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "public contacts create failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}

	// Keep admin "new leads" counter strongly consistent.
	// Best-effort: do not fail the user submission if Redis is unavailable.
	if h.rdb != nil {
		ctx := c.Request.Context()
		if exists, err := h.rdb.Exists(ctx, cache.AdminContactsNewCountKey).Result(); err == nil && exists == 0 {
			// Counter key missing (e.g., Redis restart/eviction). Reconcile from DB.
			var newLeads int64
			if err := h.db.WithContext(ctx).Model(&model.ContactLead{}).Where("status = ?", "new").Count(&newLeads).Error; err == nil {
				_ = h.rdb.Set(ctx, cache.AdminContactsNewCountKey, newLeads, 0).Err()
			}
		} else {
			_ = h.rdb.Incr(ctx, cache.AdminContactsNewCountKey).Err()
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         lead.ID,
		"created_at": lead.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}
