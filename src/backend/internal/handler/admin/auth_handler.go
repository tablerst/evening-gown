package admin

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"evening-gown/internal/auth"
	"evening-gown/internal/logging"
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

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type changePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	if h == nil || h.db == nil || h.jwtSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    "service_unavailable",
			"message": "service unavailable",
			"error":   "service unavailable",
		})
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
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "invalid_credentials",
			"message": "invalid credentials",
			"error":   "invalid credentials",
		})
		return
	}
	if user.Role != "admin" || user.Status != "active" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "invalid_credentials",
			"message": "invalid credentials",
			"error":   "invalid credentials",
		})
		return
	}
	if !security.CheckPassword(user.PasswordHash, password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "invalid_credentials",
			"message": "invalid credentials",
			"error":   "invalid credentials",
		})
		return
	}

	now := time.Now().UTC()
	_ = h.db.WithContext(c.Request.Context()).Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"last_login_at":       now,
		"failed_login_count":  0,
		"locked_until":        nil,
		"refresh_token_issued_at": now,
		"updated_at":          now,
	}).Error

	pwdAt := int64(0)
	if user.PasswordUpdatedAt != nil {
		pwdAt = user.PasswordUpdatedAt.UTC().Unix()
	}

	accessToken, accessExp, err := h.jwtSvc.IssueAdminToken(strconv.FormatUint(uint64(user.ID), 10), pwdAt)
	if err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "admin issue token failed", err, "user_id", user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "issue_token_failed",
			"message": "issue token failed",
			"error":   "issue token failed",
		})
		return
	}

	refreshToken, refreshExp, err := h.jwtSvc.IssueAdminRefreshToken(strconv.FormatUint(uint64(user.ID), 10), pwdAt)
	if err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "admin issue refresh token failed", err, "user_id", user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "issue_token_failed",
			"message": "issue token failed",
			"error":   "issue token failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":              accessToken,
		"expires_at":         accessExp.UTC().Format(time.RFC3339),
		"refresh_token":      refreshToken,
		"refresh_expires_at": refreshExp.UTC().Format(time.RFC3339),
	})
}

// Refresh exchanges a refresh token for a new access token (and rotates refresh token).
// Route: POST /api/v1/admin/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	if h == nil || h.db == nil || h.jwtSvc == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    "service_unavailable",
			"message": "service unavailable",
			"error":   "service unavailable",
		})
		return
	}

	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rt := strings.TrimSpace(req.RefreshToken)
	if rt == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "invalid_refresh_token",
			"message": "invalid refresh_token",
			"error":   "invalid refresh_token",
		})
		return
	}

	claims, err := h.jwtSvc.ParseAdminRefreshToken(rt)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "unauthorized",
			"message": "unauthorized",
			"error":   "unauthorized",
		})
		return
	}

	uid, err := strconv.ParseUint(strings.TrimSpace(claims.Subject), 10, 64)
	if err != nil || uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "unauthorized",
			"message": "unauthorized",
			"error":   "unauthorized",
		})
		return
	}

	var user model.User
	if err := h.db.WithContext(c.Request.Context()).Where("id = ? AND deleted_at IS NULL", uint(uid)).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "unauthorized",
			"message": "unauthorized",
			"error":   "unauthorized",
		})
		return
	}
	if user.Role != "admin" || user.Status != "active" {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    "forbidden",
			"message": "forbidden",
			"error":   "forbidden",
		})
		return
	}

	// Enforce password-change logout (same logic as middleware).
	{
		dbPwdAt := int64(0)
		if user.PasswordUpdatedAt != nil {
			dbPwdAt = user.PasswordUpdatedAt.UTC().Unix()
		}
		if claims.PasswordUpdatedAt != 0 {
			if claims.PasswordUpdatedAt != dbPwdAt {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    "unauthorized",
					"message": "unauthorized",
					"error":   "unauthorized",
				})
				return
			}
		} else if user.PasswordUpdatedAt != nil {
			var issuedAt time.Time
			if claims.IssuedAt != nil {
				issuedAt = claims.IssuedAt.Time
			}
			if issuedAt.IsZero() || issuedAt.Before(user.PasswordUpdatedAt.UTC()) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    "unauthorized",
					"message": "unauthorized",
					"error":   "unauthorized",
				})
				return
			}
		}
	}

	// Refresh token rotation guard: reject refresh tokens older than the latest issued marker.
	if user.RefreshTokenIssuedAt != nil {
		if claims.IssuedAt == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "unauthorized",
				"message": "unauthorized",
				"error":   "unauthorized",
			})
			return
		}
		if claims.IssuedAt.Time.UTC().Unix() < user.RefreshTokenIssuedAt.UTC().Unix() {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "unauthorized",
				"message": "unauthorized",
				"error":   "unauthorized",
			})
			return
		}
	}

	now := time.Now().UTC()
	_ = h.db.WithContext(c.Request.Context()).Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"refresh_token_issued_at": now,
		"updated_at":             now,
	}).Error

	pwdAt := int64(0)
	if user.PasswordUpdatedAt != nil {
		pwdAt = user.PasswordUpdatedAt.UTC().Unix()
	}

	accessToken, accessExp, err := h.jwtSvc.IssueAdminToken(strconv.FormatUint(uint64(user.ID), 10), pwdAt)
	if err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "admin issue token failed", err, "user_id", user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "issue_token_failed",
			"message": "issue token failed",
			"error":   "issue token failed",
		})
		return
	}

	refreshToken, refreshExp, err := h.jwtSvc.IssueAdminRefreshToken(strconv.FormatUint(uint64(user.ID), 10), pwdAt)
	if err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "admin issue refresh token failed", err, "user_id", user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "issue_token_failed",
			"message": "issue token failed",
			"error":   "issue token failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":              accessToken,
		"expires_at":         accessExp.UTC().Format(time.RFC3339),
		"refresh_token":      refreshToken,
		"refresh_expires_at": refreshExp.UTC().Format(time.RFC3339),
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	u, ok := c.Get(middleware.ContextUserKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "unauthorized",
			"message": "unauthorized",
			"error":   "unauthorized",
		})
		return
	}
	user, ok := u.(model.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "unauthorized",
			"message": "unauthorized",
			"error":   "unauthorized",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	})
}

// ChangePassword allows an authenticated admin to change their password.
// After success, old tokens will be rejected by AdminAuth middleware via PasswordUpdatedAt.
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	if h == nil || h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    "service_unavailable",
			"message": "service unavailable",
			"error":   "service unavailable",
		})
		return
	}

	u, ok := c.Get(middleware.ContextUserKey)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "unauthorized",
			"message": "unauthorized",
			"error":   "unauthorized",
		})
		return
	}
	user, ok := u.(model.User)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "unauthorized",
			"message": "unauthorized",
			"error":   "unauthorized",
		})
		return
	}

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oldPassword := strings.TrimSpace(req.OldPassword)
	newPassword := strings.TrimSpace(req.NewPassword)
	if oldPassword == "" || newPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	if oldPassword == newPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new password must be different"})
		return
	}

	// Verify old password.
	if !security.CheckPassword(user.PasswordHash, oldPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "old password is incorrect"})
		return
	}

	hash, err := security.HashPassword(newPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// JWT iat is second-precision. Truncate to seconds so AdminAuth comparison is stable.
	now := time.Now().UTC().Truncate(time.Second)
	// Ensure the marker actually changes even if the request happens within the same second
	// (common in tests / fast machines). AdminAuth compares pwd_at by equality.
	if user.PasswordUpdatedAt != nil {
		prev := user.PasswordUpdatedAt.UTC().Truncate(time.Second)
		if !now.After(prev) {
			now = prev.Add(time.Second)
		}
	}
	if err := h.db.WithContext(c.Request.Context()).Model(&model.User{}).Where("id = ? AND deleted_at IS NULL", user.ID).Updates(map[string]any{
		"password_hash":       hash,
		"password_updated_at": &now,
		"updated_at":          now,
	}).Error; err != nil {
		logging.ErrorWithStack(logging.FromGin(c), "admin change password failed", err, "user_id", user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "change_password_failed",
			"message": "change password failed",
			"error":   "change password failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok": true,
	})
}
