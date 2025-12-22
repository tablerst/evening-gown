package middleware

import (
	"net/http"
	"strconv"
	"strings"

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
		claims, err := jwtSvc.ParseToken(token)
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
		if err := db.WithContext(c.Request.Context()).First(&user, uint(uid)).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if user.Role != "admin" || user.Status != "active" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.Set(ContextUserKey, user)
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
