package model

import "time"

// UpdatePost represents a company update (later can be extended to industry news).
type UpdatePost struct {
	ID uint `gorm:"primaryKey" json:"id"`

	Type   string `gorm:"type:text;not null;default:company" json:"type"`   // company|industry
	Status string `gorm:"type:text;not null;default:draft" json:"status"` // draft|published|archived

	Tag     string `gorm:"type:text;not null;default:''" json:"tag"`
	Title   string `gorm:"type:text;not null" json:"title"`
	Summary string `gorm:"type:text;not null;default:''" json:"summary"`
	Body    string `gorm:"type:text;not null;default:''" json:"body"`
	RefCode string `gorm:"type:text;not null;default:''" json:"refCode"`

	PinnedRank int `gorm:"not null;default:0" json:"pinnedRank"`

	PublishedAt *time.Time `gorm:"index" json:"publishedAt,omitempty"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}
