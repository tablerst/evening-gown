package admin

import (
	"encoding/json"
	"net/http"

	"evening-gown/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SettingsHandler struct {
	db *gorm.DB
}

func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	return &SettingsHandler{db: db}
}

type productDetailTemplateResponse struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

// GetProductDetailTemplate returns the current product detail template.
// Route: GET /api/v1/admin/settings/product-detail-template
func (h *SettingsHandler) GetProductDetailTemplate(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	var s model.AppSetting
	err := h.db.WithContext(c.Request.Context()).
		Where("key = ?", model.SettingKeyProductDetailTemplate).
		First(&s).Error
	if err != nil {
		// Fallback to default template.
		c.JSON(http.StatusOK, productDetailTemplateResponse{
			Key:   model.SettingKeyProductDetailTemplate,
			Value: model.DefaultProductDetailTemplate(),
		})
		return
	}

	val := s.ValueJSON
	if len(val) == 0 {
		val = model.DefaultProductDetailTemplate()
	}
	// Ensure stored value is valid JSON object; otherwise fallback.
	var anyV any
	if err := json.Unmarshal(val, &anyV); err != nil {
		val = model.DefaultProductDetailTemplate()
	} else {
		if _, ok := anyV.(map[string]any); !ok {
			val = model.DefaultProductDetailTemplate()
		}
	}

	c.JSON(http.StatusOK, productDetailTemplateResponse{Key: s.Key, Value: val})
}

type putProductDetailTemplateRequest struct {
	Value json.RawMessage `json:"value" binding:"required"`
}

// PutProductDetailTemplate replaces the current product detail template.
// Route: PUT /api/v1/admin/settings/product-detail-template
func (h *SettingsHandler) PutProductDetailTemplate(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	var req putProductDetailTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate: must be a JSON object.
	var anyV any
	if err := json.Unmarshal(req.Value, &anyV); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid value"})
		return
	}
	if _, ok := anyV.(map[string]any); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "value must be a JSON object"})
		return
	}

	set := model.AppSetting{Key: model.SettingKeyProductDetailTemplate, ValueJSON: req.Value}
	if err := h.db.WithContext(c.Request.Context()).Save(&set).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, productDetailTemplateResponse{Key: set.Key, Value: set.ValueJSON})
}
