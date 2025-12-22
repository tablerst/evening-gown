package model

import (
	"encoding/json"
	"time"
)

// Product is a B2B catalog item. Public APIs must not expose exact prices.
type Product struct {
	ID uint `gorm:"primaryKey" json:"id"`

	Slug    string `gorm:"type:text;uniqueIndex" json:"slug"`
	StyleNo int    `gorm:"uniqueIndex;not null" json:"styleNo"`

	Season       string `gorm:"type:text;not null" json:"season"`       // ss25|fw25
	Category     string `gorm:"type:text;not null" json:"category"`     // gown|couture|bridal
	Availability string `gorm:"type:text;not null" json:"availability"` // in_stock|preorder|archived

	IsNew   bool `gorm:"not null;default:false" json:"isNew"`
	NewRank int  `gorm:"not null;default:0" json:"newRank"`

	CoverImageURL string `gorm:"type:text;not null;default:''" json:"coverImage"`
	HoverImageURL string `gorm:"type:text;not null;default:''" json:"hoverImage"`

	// Negotiation-first pricing.
	PriceMode string `gorm:"type:text;not null;default:negotiable" json:"priceMode"` // negotiable

	// DetailJSON stores product details and configurable options.
	// Suggested structure:
	// - title_i18n, description_i18n
	// - specs[]
	// - option_groups[]
	DetailJSON json.RawMessage `gorm:"type:jsonb" json:"detail"`

	PublishedAt *time.Time `gorm:"index" json:"publishedAt,omitempty"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}
