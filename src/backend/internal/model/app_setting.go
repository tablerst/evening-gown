package model

import (
	"encoding/json"
	"time"
)

// AppSetting stores small JSON config blobs keyed by a stable string.
// It is used for backoffice-configurable behavior (e.g. product detail templates).
type AppSetting struct {
	Key string `gorm:"primaryKey;type:text" json:"key"`

	// ValueJSON holds arbitrary JSON payload.
	ValueJSON json.RawMessage `gorm:"type:jsonb;not null" json:"value"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
