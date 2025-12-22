package public

import (
	"encoding/json"
	"net/http"
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

type eventCreateRequest struct {
	EventType string `json:"event_type" binding:"required"`
	OccurredAt string `json:"occurred_at"` // RFC3339 optional

	SessionID string `json:"session_id"`
	AnonID    string `json:"anon_id"`

	ProductID *uint `json:"product_id"`

	PageURL  string `json:"page_url"`
	Referrer string `json:"referrer"`

	UTMSource   string `json:"utm_source"`
	UTMMedium   string `json:"utm_medium"`
	UTMCampaign string `json:"utm_campaign"`
	UTMContent  string `json:"utm_content"`
	UTMTerm     string `json:"utm_term"`

	Payload json.RawMessage `json:"payload"`
}

func (h *EventsHandler) Create(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	var req eventCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	occurred := time.Now().UTC()
	if strings.TrimSpace(req.OccurredAt) != "" {
		if t, err := time.Parse(time.RFC3339, strings.TrimSpace(req.OccurredAt)); err == nil {
			occurred = t.UTC()
		}
	}

	e := model.Event{
		EventType:   strings.TrimSpace(req.EventType),
		OccurredAt:  occurred,
		SessionID:   strings.TrimSpace(req.SessionID),
		AnonID:      strings.TrimSpace(req.AnonID),
		ProductID:   req.ProductID,
		PageURL:     strings.TrimSpace(req.PageURL),
		Referrer:    strings.TrimSpace(req.Referrer),
		UTMSource:   strings.TrimSpace(req.UTMSource),
		UTMMedium:   strings.TrimSpace(req.UTMMedium),
		UTMCampaign: strings.TrimSpace(req.UTMCampaign),
		UTMContent:  strings.TrimSpace(req.UTMContent),
		UTMTerm:     strings.TrimSpace(req.UTMTerm),
		Payload:     req.Payload,
	}

	if err := h.db.WithContext(c.Request.Context()).Create(&e).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"ok": true, "id": e.ID})
}
