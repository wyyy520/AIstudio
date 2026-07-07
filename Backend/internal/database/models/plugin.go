package models

import "time"

// Plugin represents a registered plugin.
type Plugin struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Version     string    `gorm:"size:32;not null" json:"version"`
	Description string    `gorm:"size:1024" json:"description"`
	Status      string    `gorm:"size:32;default:installed" json:"status"` // installed, enabled, disabled, error
	Config      string    `gorm:"type:text" json:"config,omitempty"`       // JSON config
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (Plugin) TableName() string {
	return "plugins"
}