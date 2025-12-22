package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	jwtauth "evening-gown/internal/auth"
	"evening-gown/internal/bootstrap"
	"evening-gown/internal/cache"
	"evening-gown/internal/config"
	"evening-gown/internal/database"
	adminHandlers "evening-gown/internal/handler/admin"
	authHandlerPkg "evening-gown/internal/handler/auth"
	"evening-gown/internal/handler/health"
	publicHandlers "evening-gown/internal/handler/public"
	"evening-gown/internal/middleware"
	"evening-gown/internal/router"
	"evening-gown/internal/storage"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Run wires dependencies and starts the HTTP server with graceful shutdown.
func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var db *gorm.DB
	if cfg.Postgres.DSN != "" {
		db, err = database.New(ctx, cfg.Postgres)
		if err != nil {
			return err
		}
		defer func() {
			if err := database.Close(db); err != nil {
				log.Printf("postgres close error: %v", err)
			}
		}()
	} else {
		log.Println("postgres disabled: POSTGRES_DSN not set")
	}

	// JWT service (shared by admin auth middleware).
	var jwtSvc *jwtauth.Service
	if cfg.JWT.Secret != "" {
		jwtSvc, err = jwtauth.New(cfg.JWT)
		if err != nil {
			return err
		}
	} else {
		log.Println("jwt disabled: JWT_SECRET not set")
	}

	var redisClient *redis.Client
	if cfg.Redis.Addr != "" {
		redisClient, err = cache.NewClient(ctx, cfg.Redis)
		if err != nil {
			return err
		}
		defer func() {
			if err := redisClient.Close(); err != nil {
				log.Printf("redis close error: %v", err)
			}
		}()
	} else {
		log.Println("redis disabled: REDIS_ADDR not set")
	}

	minioClient, err := storage.NewClient(ctx, cfg.Minio)
	if err != nil {
		return err
	}
	if minioClient == nil {
		log.Println("minio disabled: MINIO_ENDPOINT not set")
	}

	// Legacy auth handler (dev-only token issuer / verify helper).
	var authHandler *authHandlerPkg.Handler
	if cfg.JWT.Secret != "" {
		authHandler = authHandlerPkg.New(cfg.JWT)
	}

	healthHandler := health.New(db, redisClient, minioClient)

	deps := router.Dependencies{Health: healthHandler, Auth: authHandler, EnableDevTokenIssuer: cfg.Dev.EnableDevTokenIssuer}

	// Business APIs require Postgres.
	if db != nil {
		if err := bootstrap.AutoMigrate(db); err != nil {
			return err
		}
		if err := bootstrap.EnsureSingleAdmin(db, cfg.Admin.Email, cfg.Admin.Password); err != nil {
			return err
		}

		deps.Public.Products = publicHandlers.NewProductsHandler(db)
		deps.Public.Updates = publicHandlers.NewUpdatesHandler(db)
		deps.Public.Contacts = publicHandlers.NewContactsHandler(db)
		deps.Public.Events = publicHandlers.NewEventsHandler(db)

		deps.Admin.Auth = adminHandlers.NewAuthHandler(db, jwtSvc)
		deps.Admin.Products = adminHandlers.NewProductsHandler(db)
		deps.Admin.Updates = adminHandlers.NewUpdatesHandler(db)
		deps.Admin.Contacts = adminHandlers.NewContactsHandler(db)
		deps.Admin.Events = adminHandlers.NewEventsHandler(db)
		deps.Admin.AuthMiddleware = middleware.AdminAuth(db, jwtSvc)
	} else {
		log.Println("business APIs disabled: postgres not configured")
	}

	r := router.New(deps)

	srv := &http.Server{
		Addr:    cfg.App.Addr(),
		Handler: r,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}()

	log.Printf("HTTP server listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
