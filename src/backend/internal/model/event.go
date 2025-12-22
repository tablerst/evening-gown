package model

import (
	"encoding/json"
	"time"
)

// Event stores anonymous intent metadata for analytics (e.g., poster generated, share click).
// No poster bytes/images are stored.
type Event struct {
	ID uint `gorm:"primaryKey" json:"id"`

	EventType  string    `gorm:"type:text;not null" json:"eventType"`
	OccurredAt time.Time `gorm:"not null;index" json:"occurredAt"`

	SessionID string `gorm:"type:text;not null;default:''" json:"sessionId"`
	AnonID    string `gorm:"type:text;not null;default:''" json:"anonId"`

	UserID    *uint `gorm:"index" json:"userId,omitempty"`
	ProductID *uint `gorm:"index" json:"productId,omitempty"`

	PageURL  string `gorm:"type:text;not null;default:''" json:"pageUrl"`
	Referrer string `gorm:"type:text;not null;default:''" json:"referrer"`

	UTMSource   string `gorm:"type:text;not null;default:''" json:"utmSource"`
	UTMMedium   string `gorm:"type:text;not null;default:''" json:"utmMedium"`
	UTMCampaign string `gorm:"type:text;not null;default:''" json:"utmCampaign"`
	UTMContent  string `gorm:"type:text;not null;default:''" json:"utmContent"`
	UTMTerm     string `gorm:"type:text;not null;default:''" json:"utmTerm"`

	Payload json.RawMessage `gorm:"type:jsonb" json:"payload"`

	CreatedAt time.Time `json:"createdAt"`
}
