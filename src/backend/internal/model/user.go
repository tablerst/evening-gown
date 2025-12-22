package model

import "time"

// User represents a backoffice user.
//
// This project uses a single super admin (no multi-tenant).
type User struct {
	ID uint `gorm:"primaryKey" json:"id"`

	Email        string `gorm:"type:text;uniqueIndex;not null" json:"email"`
	PasswordHash string `gorm:"type:text;not null" json:"-"`

	Role   string `gorm:"type:text;not null;default:admin" json:"role"`   // admin
	Status string `gorm:"type:text;not null;default:active" json:"status"` // active|disabled|locked

	FailedLoginCount  int        `gorm:"not null;default:0" json:"failedLoginCount"`
	LockedUntil       *time.Time `gorm:"" json:"lockedUntil,omitempty"`
	LastLoginAt       *time.Time `gorm:"" json:"lastLoginAt,omitempty"`
	PasswordUpdatedAt *time.Time `gorm:"" json:"passwordUpdatedAt,omitempty"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}
