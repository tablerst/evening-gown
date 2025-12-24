package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"evening-gown/internal/auth"
	"evening-gown/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const ContextUserKey = "auth.user"

func AdminAuth(db *gorm.DB, jwtSvc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil || jwtSvc == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "admin auth unavailable"})
			return
		}

		token := tokenFromRequest(c)
		claims, err := jwtSvc.ParseAdminToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		uid, err := strconv.ParseUint(strings.TrimSpace(claims.Subject), 10, 64)
		if err != nil || uid == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var user model.User
		// NOTE: model.User uses a manual DeletedAt *time.Time, not gorm.DeletedAt.
		// We must explicitly filter soft-deleted users.
		if err := db.WithContext(c.Request.Context()).Where("id = ? AND deleted_at IS NULL", uint(uid)).First(&user).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if user.Role != "admin" || user.Status != "active" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		// Force logout after password change.
		// Prefer comparing the password marker embedded in token (pwd_at) with the current
		// user.PasswordUpdatedAt in DB. This avoids ambiguity when password change and
		// token issuance happen within the same second (JWT iat is second precision).
		{
			dbPwdAt := int64(0)
			if user.PasswordUpdatedAt != nil {
				dbPwdAt = user.PasswordUpdatedAt.UTC().Unix()
			}

			if claims.PasswordUpdatedAt != 0 {
				if claims.PasswordUpdatedAt != dbPwdAt {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
			} else if user.PasswordUpdatedAt != nil {
				// Backward compatibility: older tokens may not have pwd_at.
				var issuedAt time.Time
				if claims.IssuedAt != nil {
					issuedAt = claims.IssuedAt.Time
				}
				if issuedAt.IsZero() || issuedAt.Before(user.PasswordUpdatedAt.UTC()) {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
			}
		}

		c.Set(ContextUserKey, user)
		EnrichLoggerWithAdmin(c, user)
		c.Next()
	}
}

func tokenFromRequest(c *gin.Context) string {
	if c == nil {
		return ""
	}

	authz := strings.TrimSpace(c.GetHeader("Authorization"))
	parts := strings.SplitN(authz, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}
