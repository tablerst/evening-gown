package public

import (
	"net/http"
	"strings"

	"evening-gown/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ContactsHandler struct {
	db *gorm.DB
}

func NewContactsHandler(db *gorm.DB) *ContactsHandler {
	return &ContactsHandler{db: db}
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         lead.ID,
		"created_at": lead.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	})
}
