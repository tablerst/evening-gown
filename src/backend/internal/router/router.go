package router

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"

	adminHandlers "evening-gown/internal/handler/admin"
	"evening-gown/internal/handler/auth"
	"evening-gown/internal/handler/health"
	publicHandlers "evening-gown/internal/handler/public"
)

// Dependencies groups handlers required by the router.
type Dependencies struct {
	Health *health.Handler
	Auth   *auth.Handler

	// Dev toggles
	EnableDevTokenIssuer bool

	// Public website APIs (no auth)
	Public struct {
		Products *publicHandlers.ProductsHandler
		Updates  *publicHandlers.UpdatesHandler
		Contacts *publicHandlers.ContactsHandler
		Events   *publicHandlers.EventsHandler
	}

	// Admin backoffice APIs (JWT-protected)
	Admin struct {
		Auth     *adminHandlers.AuthHandler
		Products *adminHandlers.ProductsHandler
		Updates  *adminHandlers.UpdatesHandler
		Contacts *adminHandlers.ContactsHandler
		Events   *adminHandlers.EventsHandler
		// Middleware applied to protected admin routes.
		AuthMiddleware gin.HandlerFunc
	}
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
		// Be explicit instead of relying on cors.Default() so behavior remains stable
		// across gin-contrib/cors versions.
		cfg := cors.Config{
			AllowAllOrigins: true,
			AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:    []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-Id"},
			ExposeHeaders:   []string{"X-Request-Id"},
			MaxAge:          12 * time.Hour,
		}
		r.Use(cors.New(cfg))
	} else {
		cfg := cors.Config{
			AllowOrigins:     allowOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-Id"},
			ExposeHeaders:    []string{"X-Request-Id"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
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
		// Legacy dev-only endpoint (UNSAFE): issues tokens for arbitrary subjects.
		if deps.EnableDevTokenIssuer {
			authGroup.POST("/token", deps.Auth.IssueToken)
		}
		authGroup.GET("/verify", deps.Auth.VerifyToken)
	}

	// Public website APIs (no auth)
	if deps.Public.Products != nil || deps.Public.Updates != nil || deps.Public.Contacts != nil || deps.Public.Events != nil {
		api := r.Group("/api/v1")
		if deps.Public.Products != nil {
			api.GET("/products", deps.Public.Products.List)
			api.GET("/products/:id", deps.Public.Products.Get)
		}
		if deps.Public.Updates != nil {
			api.GET("/updates", deps.Public.Updates.List)
			api.GET("/updates/:id", deps.Public.Updates.Get)
		}
		if deps.Public.Contacts != nil {
			api.POST("/contacts", deps.Public.Contacts.Create)
		}
		if deps.Public.Events != nil {
			api.POST("/events", deps.Public.Events.Create)
		}
	}

	// Admin backoffice APIs (JWT-protected)
	if deps.Admin.Auth != nil || deps.Admin.Products != nil || deps.Admin.Updates != nil || deps.Admin.Contacts != nil || deps.Admin.Events != nil {
		admin := r.Group("/api/v1/admin")
		if deps.Admin.Auth != nil {
			// Login is unprotected.
			admin.POST("/auth/login", deps.Admin.Auth.Login)
		}
		// Protected admin routes.
		if deps.Admin.AuthMiddleware != nil {
			admin.Use(deps.Admin.AuthMiddleware)
		}
		if deps.Admin.Auth != nil {
			admin.GET("/me", deps.Admin.Auth.Me)
		}
		if deps.Admin.Products != nil {
			admin.GET("/products", deps.Admin.Products.List)
			admin.POST("/products", deps.Admin.Products.Create)
			admin.GET("/products/:id", deps.Admin.Products.Get)
			admin.PATCH("/products/:id", deps.Admin.Products.Update)
			admin.POST("/products/:id/publish", deps.Admin.Products.Publish)
			admin.POST("/products/:id/unpublish", deps.Admin.Products.Unpublish)
			admin.DELETE("/products/:id", deps.Admin.Products.Delete)
		}
		if deps.Admin.Updates != nil {
			admin.GET("/updates", deps.Admin.Updates.List)
			admin.POST("/updates", deps.Admin.Updates.Create)
			admin.GET("/updates/:id", deps.Admin.Updates.Get)
			admin.PATCH("/updates/:id", deps.Admin.Updates.Update)
			admin.POST("/updates/:id/publish", deps.Admin.Updates.Publish)
			admin.POST("/updates/:id/unpublish", deps.Admin.Updates.Unpublish)
			admin.DELETE("/updates/:id", deps.Admin.Updates.Delete)
		}
		if deps.Admin.Contacts != nil {
			admin.GET("/contacts", deps.Admin.Contacts.List)
			admin.GET("/contacts/:id", deps.Admin.Contacts.Get)
			admin.PATCH("/contacts/:id", deps.Admin.Contacts.Update)
			admin.DELETE("/contacts/:id", deps.Admin.Contacts.Delete)
		}
		if deps.Admin.Events != nil {
			admin.GET("/events", deps.Admin.Events.List)
			admin.GET("/events/:id", deps.Admin.Events.Get)
			admin.DELETE("/events/:id", deps.Admin.Events.Delete)
		}
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
