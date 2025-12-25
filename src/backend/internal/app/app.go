package app

import (
	"context"
	"errors"
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
	"evening-gown/internal/logging"
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

	logger, closeLogger, err := logging.Init(cfg.Log)
	if err != nil {
		return err
	}
	defer func() {
		if closeLogger == nil {
			return
		}
		if err := closeLogger(); err != nil {
			logger.Warn("log close error", "err", err)
		}
	}()

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
				logger.Warn("postgres close error", "err", err)
			}
		}()
	} else {
		logger.Info("postgres disabled: POSTGRES_DSN not set")
	}

	// JWT service (shared by admin auth middleware).
	var jwtSvc *jwtauth.Service
	if cfg.JWT.Secret != "" {
		jwtSvc, err = jwtauth.New(cfg.JWT)
		if err != nil {
			return err
		}
	} else {
		logger.Info("jwt disabled: JWT_SECRET not set")
	}

	var redisClient *redis.Client
	if cfg.Redis.Addr != "" {
		redisClient, err = cache.NewClient(ctx, cfg.Redis)
		if err != nil {
			return err
		}
		defer func() {
			if err := redisClient.Close(); err != nil {
				logger.Warn("redis close error", "err", err)
			}
		}()
	} else {
		logger.Info("redis disabled: REDIS_ADDR not set")
	}

	minioClient, err := storage.NewClient(ctx, cfg.Minio)
	if err != nil {
		return err
	}
	if minioClient == nil {
		logger.Info("minio disabled: MINIO_ENDPOINT not set")
	}

	// Legacy auth handler (dev-only token issuer / verify helper).
	var authHandler *authHandlerPkg.Handler
	if cfg.JWT.Secret != "" {
		authHandler = authHandlerPkg.New(cfg.JWT)
	}

	healthHandler := health.New(db, redisClient, minioClient)
	publicCache := cache.NewPublicCache(redisClient)

	deps := router.Dependencies{Health: healthHandler, Auth: authHandler, EnableDevTokenIssuer: cfg.Dev.EnableDevTokenIssuer}
	if minioClient != nil {
		deps.Public.Assets = publicHandlers.NewAssetsHandler(db, minioClient, cfg.Minio, publicCache)
	}

	// Business APIs require Postgres.
	if db != nil {
		if err := bootstrap.AutoMigrate(db); err != nil {
			return err
		}
		if err := bootstrap.EnsureSingleAdmin(db, cfg.Admin.Email, cfg.Admin.Password); err != nil {
			return err
		}

		deps.Public.Products = publicHandlers.NewProductsHandler(db, publicCache)
		deps.Public.Updates = publicHandlers.NewUpdatesHandler(db, publicCache)
		deps.Public.Contacts = publicHandlers.NewContactsHandler(db)
		deps.Public.Events = publicHandlers.NewEventsHandler(db)

		deps.Admin.Auth = adminHandlers.NewAuthHandler(db, jwtSvc)
		if minioClient != nil {
			deps.Admin.Assets = adminHandlers.NewAssetsHandler(db, minioClient, cfg.Minio)
		}
		deps.Admin.Uploads = adminHandlers.NewUploadsHandler(minioClient, cfg.Minio, cfg.Upload)
		deps.Admin.Products = adminHandlers.NewProductsHandler(db, publicCache)
		deps.Admin.Updates = adminHandlers.NewUpdatesHandler(db, publicCache)
		deps.Admin.Contacts = adminHandlers.NewContactsHandler(db)
		deps.Admin.Events = adminHandlers.NewEventsHandler(db)
		deps.Admin.AuthMiddleware = middleware.AdminAuth(db, jwtSvc)
	} else {
		logger.Info("business APIs disabled: postgres not configured")
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
			logger.Error("server shutdown error", "err", err)
		}
	}()

	logger.Info("HTTP server listening", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
