package models

import "time"

// Plugin represents a registered plugin in the database.
type Plugin struct {
	ID          uint      `gorm:"primaryKey" json:"-"`
	PluginID    string    `gorm:"uniqueIndex;size:128;not null" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:128;not null" json:"name"`
	Version     string    `gorm:"size:32;not null" json:"version"`
	Author      string    `gorm:"size:128" json:"author"`
	Type        string    `gorm:"size:32" json:"category"`
	Description string    `gorm:"size:1024" json:"description"`
	Status      string    `gorm:"size:32;default:not_installed;index:idx_plugins_status_enabled" json:"status"`
	Enabled     bool      `gorm:"default:false;index:idx_plugins_status_enabled" json:"enabled"`
	Path        string    `gorm:"size:512" json:"path"`
	Source      string    `gorm:"size:32" json:"source"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (Plugin) TableName() string {
	return "plugins"
}