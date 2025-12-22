package model

import "time"

// ContactLead is a "contact us" submission (anonymous, no auth on public website).
type ContactLead struct {
	ID uint `gorm:"primaryKey" json:"id"`

	Name    string `gorm:"type:text;not null;default:''" json:"name"`
	Phone   string `gorm:"type:text;not null;default:''" json:"phone"`
	Wechat  string `gorm:"type:text;not null;default:''" json:"wechat"`
	Message string `gorm:"type:text;not null;default:''" json:"message"`

	SourcePage  string `gorm:"type:text;not null;default:''" json:"sourcePage"`
	UTMSource   string `gorm:"type:text;not null;default:''" json:"utmSource"`
	UTMMedium   string `gorm:"type:text;not null;default:''" json:"utmMedium"`
	UTMCampaign string `gorm:"type:text;not null;default:''" json:"utmCampaign"`
	UTMContent  string `gorm:"type:text;not null;default:''" json:"utmContent"`
	UTMTerm     string `gorm:"type:text;not null;default:''" json:"utmTerm"`

	Status string `gorm:"type:text;not null;default:new" json:"status"` // new|contacted|closed

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
