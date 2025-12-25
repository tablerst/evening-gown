package bootstrap

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"evening-gown/internal/model"
	"evening-gown/internal/security"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrPostgresRequired = errors.New("postgres is required for backoffice features")

// AutoMigrate runs gorm AutoMigrate for all project models.
func AutoMigrate(db *gorm.DB) error {
	if db == nil {
		return ErrPostgresRequired
	}

	// Backward-compatible migration: style_no used to be integer.
	// We now store it as text to support values like "AB-001".
	if err := migrateProductsStyleNoToText(db); err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Product{},
		&model.AppSetting{},
		&model.UpdatePost{},
		&model.ContactLead{},
		&model.Event{},
	); err != nil {
		return err
	}

	// Ensure default product detail template exists.
	if err := ensureProductDetailTemplateSetting(db); err != nil {
		return err
	}

	return nil
}

func migrateProductsStyleNoToText(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	// Tests often use SQLite; the DO $$ block is Postgres-only.
	if db.Dialector == nil || db.Dialector.Name() != "postgres" {
		return nil
	}
	// Idempotent: only runs when the column exists and is integer.
	const q = `
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'public'
      AND table_name = 'products'
      AND column_name = 'style_no'
      AND data_type = 'integer'
  ) THEN
    ALTER TABLE products
      ALTER COLUMN style_no TYPE text
      USING style_no::text;
  END IF;
END $$;`
	return db.Exec(q).Error
}

func ensureProductDetailTemplateSetting(db *gorm.DB) error {
	if db == nil {
		return ErrPostgresRequired
	}

	// Insert-if-not-exists with a stable default.
	set := model.AppSetting{
		Key:      model.SettingKeyProductDetailTemplate,
		ValueJSON: model.DefaultProductDetailTemplate(),
	}
	// OnConflict DoNothing keeps user-customized template unchanged.
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&set).Error
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

	// JWT iat is second-precision. Truncate to seconds so AdminAuth comparison is stable.
	now := time.Now().UTC().Truncate(time.Second)
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
