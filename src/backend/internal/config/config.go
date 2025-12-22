package config

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config aggregates application configuration.
type Config struct {
	App      AppConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Minio    MinioConfig
	JWT      JWTConfig
	Admin    AdminConfig
	Dev      DevConfig
}

// AdminConfig controls bootstrap and login for the single super admin.
//
// Notes:
// - ADMIN_PASSWORD is only used for bootstrapping when no admin exists yet.
// - Do NOT put real credentials into git.
type AdminConfig struct {
	Email    string
	Password string
}

// DevConfig contains development-only toggles.
type DevConfig struct {
	// EnableDevTokenIssuer keeps legacy /auth/token endpoint enabled.
	// It is unsafe for production.
	EnableDevTokenIssuer bool
}

// AppConfig controls HTTP server settings.
type AppConfig struct {
	Host string
	Port string
}

// Addr returns host:port with sensible defaults if unset.
func (a AppConfig) Addr() string {
	host := defaultString(a.Host, "0.0.0.0")
	port := defaultString(a.Port, "8080")
	return net.JoinHostPort(host, port)
}

// PostgresConfig defines PostgreSQL connection tuning options.
type PostgresConfig struct {
	DSN             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
}

// RedisConfig defines Redis client options.
type RedisConfig struct {
	Addr        string
	Password    string
	DB          int
	PoolSize    int
	DialTimeout time.Duration
}

// MinioConfig defines MinIO/S3 compatible client options.
type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Region    string
	Bucket    string
	UseSSL    bool
}

// JWTConfig defines JSON Web Token signing and validation settings.
type JWTConfig struct {
	Secret    string
	Issuer    string
	Audience  string
	ExpiresIn time.Duration
}

// Load reads environment variables (optionally from .env) and returns a Config.
func Load() (Config, error) {
	// Attempt to load a local .env file for development. Missing file is ignored.
	if err := loadDotEnv(".env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("config: unable to load .env: %v", err)
	}

	cfg := Config{
		App: AppConfig{
			Host: getEnv("APP_HOST", "0.0.0.0"),
			Port: getEnv("APP_PORT", "8080"),
		},
		Postgres: PostgresConfig{
			DSN:             getEnv("POSTGRES_DSN", ""),
			MaxConns:        getInt32Env("POSTGRES_MAX_CONNS", 10),
			MinConns:        getInt32Env("POSTGRES_MIN_CONNS", 2),
			MaxConnLifetime: getDurationEnv("POSTGRES_MAX_CONN_LIFETIME", time.Hour),
		},
		Redis: RedisConfig{
			Addr:        getEnv("REDIS_ADDR", ""),
			Password:    getEnv("REDIS_PASSWORD", ""),
			DB:          getIntEnv("REDIS_DB", 0),
			PoolSize:    getIntEnv("REDIS_POOL_SIZE", 10),
			DialTimeout: getDurationEnv("REDIS_DIAL_TIMEOUT", 5*time.Second),
		},
		Minio: MinioConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", ""),
			AccessKey: getEnv("MINIO_ACCESS_KEY", ""),
			SecretKey: getEnv("MINIO_SECRET_KEY", ""),
			Region:    getEnv("MINIO_REGION", ""),
			Bucket:    getEnv("MINIO_BUCKET", ""),
			UseSSL:    getBoolEnv("MINIO_USE_SSL", false),
		},
		JWT: JWTConfig{
			Secret:    getEnv("JWT_SECRET", ""),
			Issuer:    getEnv("JWT_ISSUER", "evening-gown"),
			Audience:  getEnv("JWT_AUDIENCE", ""),
			ExpiresIn: getDurationEnv("JWT_EXPIRES_IN", 24*time.Hour),
		},
		Admin: AdminConfig{
			Email:    getEnv("ADMIN_EMAIL", ""),
			Password: getEnv("ADMIN_PASSWORD", ""),
		},
		Dev: DevConfig{
			EnableDevTokenIssuer: getBoolEnv("ENABLE_DEV_TOKEN_ISSUER", false),
		},
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		log.Printf("config: %s expects integer, got %q: %v (using %d)", key, raw, err, fallback)
		return fallback
	}
	return value
}

func getInt32Env(key string, fallback int32) int32 {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback
	}
	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		log.Printf("config: %s expects int32, got %q: %v (using %d)", key, raw, err, fallback)
		return fallback
	}
	return int32(value)
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		log.Printf("config: %s expects duration, got %q: %v (using %s)", key, raw, err, fallback)
		return fallback
	}
	return value
}

func getBoolEnv(key string, fallback bool) bool {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		log.Printf("config: %s expects bool, got %q: %v (using %t)", key, raw, err, fallback)
		return fallback
	}
	return value
}

// loadDotEnv loads key=value pairs from a dotenv file into environment variables.
//
// It keeps existing environment variables untouched, so real env overrides .env.
func loadDotEnv(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}

	f, err := os.Open(abs)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		idx := strings.IndexRune(line, '=')
		if idx <= 0 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if key == "" {
			continue
		}

		// Remove surrounding quotes.
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, val); err != nil {
			return fmt.Errorf("setenv %s: %w", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read dotenv: %w", err)
	}
	return nil
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
