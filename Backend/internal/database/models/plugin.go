package models

import "time"

// Plugin represents a registered plugin in the database.
type Plugin struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	PluginID    string    `gorm:"uniqueIndex;size:128;not null" json:"plugin_id"`
	Name        string    `gorm:"uniqueIndex;size:128;not null" json:"name"`
	Version     string    `gorm:"size:32;not null" json:"version"`
	Author      string    `gorm:"size:128" json:"author"`
	Type        string    `gorm:"size:32" json:"type"`
	Description string    `gorm:"size:1024" json:"description"`
	Status      string    `gorm:"size:32;default:not_installed" json:"status"`
	Enabled     bool      `gorm:"default:false" json:"enabled"`
	Path        string    `gorm:"size:512" json:"path"`
	Source      string    `gorm:"size:32" json:"source"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Plugin) TableName() string {
	return "plugins"
}