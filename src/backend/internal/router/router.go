package router

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"

	"evening-gown/internal/handler/auth"
	"evening-gown/internal/handler/health"
)

// Dependencies groups handlers required by the router.
type Dependencies struct {
	Health *health.Handler
	Auth   *auth.Handler
}

// New builds a gin.Engine with common middleware and routes.
func New(deps Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Request ID (X-Request-Id). Useful for tracing and logs.
	r.Use(requestid.New())

	// CORS: allow-all by default for development; customize via env if needed.
	//
	// Set CORS_ALLOW_ORIGINS to comma-separated origins to restrict.
	// Example: https://example.com,https://admin.example.com
	allowOrigins := splitCommaEnv("CORS_ALLOW_ORIGINS")
	if len(allowOrigins) == 0 {
		r.Use(cors.Default())
	} else {
		cfg := cors.Config{
			AllowOrigins:     allowOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-Id"},
			ExposeHeaders:    []string{"X-Request-Id"},
			AllowCredentials: true,
		}
		r.Use(cors.New(cfg))
	}

	// pprof (debug only).
	// Enable by setting ENABLE_PPROF=true.
	if strings.EqualFold(os.Getenv("ENABLE_PPROF"), "true") {
		pprof.Register(r)
	}

	if deps.Health != nil {
		r.GET("/ping", deps.Health.Ping)
		r.GET("/healthz", deps.Health.Health)
	}

	if deps.Auth != nil {
		authGroup := r.Group("/auth")
		authGroup.POST("/token", deps.Auth.IssueToken)
		authGroup.GET("/verify", deps.Auth.VerifyToken)
	}

	return r
}

func splitCommaEnv(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
