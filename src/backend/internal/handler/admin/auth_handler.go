package admin

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"evening-gown/internal/auth"
	"evening-gown/internal/middleware"
	"evening-gown/internal/model"
	"evening-gown/internal/security"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db     *gorm.DB
	jwtSvc *auth.Service
}

func NewAuthHandler(db *gorm.DB, jwtSvc *auth.Service) *AuthHandler {
	return &AuthHandler{db: db, jwtSvc: jwtSvc}
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	if h == nil || h.db == nil || h.jwtSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := strings.TrimSpace(req.Password)
	if email == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credentials"})
		return
	}

	var user model.User
	if err := h.db.WithContext(c.Request.Context()).Where("email = ? AND deleted_at IS NULL", email).First(&user).Error; err != nil {
		// Avoid leaking which part failed.
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if user.Role != "admin" || user.Status != "active" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if !security.CheckPassword(user.PasswordHash, password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	now := time.Now().UTC()
	_ = h.db.WithContext(c.Request.Context()).Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"last_login_at":       now,
		"failed_login_count":  0,
		"locked_until":        nil,
		"updated_at":          now,
	}).Error

	token, exp, err := h.jwtSvc.IssueToken(strconv.FormatUint(uint64(user.ID), 10))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "issue token failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"expires_at": exp.UTC().Format(time.RFC3339),
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	u, ok := c.Get(middleware.ContextUserKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, ok := u.(model.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}
