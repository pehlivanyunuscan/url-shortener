package models

import (
	"time"

	"gorm.io/gorm"
)

type URL struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	OriginalURL string         `gorm:"not null;unique" json:"original_url"`
	ShortURL    string         `gorm:"not null;unique" json:"short_url"`
	CreatedAt   time.Time      `json:"created_at"`
	ExpiresAt   time.Time      `json:"expires_at"`
	UsageCount  int            `json:"usage_count"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at"`
}
