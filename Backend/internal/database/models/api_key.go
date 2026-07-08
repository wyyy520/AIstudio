package models

import "time"

type APIKey struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"userId"`
	Provider  string    `gorm:"size:32;not null;index" json:"provider"`
	Name      string    `gorm:"size:64" json:"name"`
	KeyPrefix string    `gorm:"size:8" json:"keyPrefix"`
	KeyHash   string    `gorm:"column:key_hash;size:512;not null" json:"-"`
	Status    string    `gorm:"size:32;default:active;not null" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (APIKey) TableName() string {
	return "api_keys"
}
