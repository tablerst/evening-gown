package bootstrap

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"evening-gown/internal/model"
	"evening-gown/internal/security"

	"gorm.io/gorm"
)

var ErrPostgresRequired = errors.New("postgres is required for backoffice features")

// AutoMigrate runs gorm AutoMigrate for all project models.
func AutoMigrate(db *gorm.DB) error {
	if db == nil {
		return ErrPostgresRequired
	}
	return db.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.UpdatePost{},
		&model.ContactLead{},
		&model.Event{},
	)
}

// EnsureSingleAdmin creates the first super admin user if no admin exists.
//
// It enforces a "single super admin" policy on creation, but does not delete
// extra rows if they already exist.
func EnsureSingleAdmin(db *gorm.DB, email, password string) error {
	if db == nil {
		return ErrPostgresRequired
	}

	var adminCount int64
	if err := db.Model(&model.User{}).Where("role = ? AND deleted_at IS NULL", "admin").Count(&adminCount).Error; err != nil {
		return fmt.Errorf("count admin users: %w", err)
	}
	if adminCount > 0 {
		return nil
	}

	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	if email == "" || password == "" {
		return fmt.Errorf("no admin exists yet; set ADMIN_EMAIL and ADMIN_PASSWORD to bootstrap")
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	now := time.Now().UTC()
	user := model.User{
		Email:          email,
		PasswordHash:   hash,
		Role:           "admin",
		Status:         "active",
		PasswordUpdatedAt: &now,
	}
	if err := db.Create(&user).Error; err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}

	return nil
}
