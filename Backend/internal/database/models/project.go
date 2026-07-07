package models

import "time"

// Project represents an AI project.
type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Description string    `gorm:"size:1024" json:"description"`
	OwnerID     uint      `gorm:"index;not null" json:"ownerId"`
	Status      string    `gorm:"size:32;default:active" json:"status"` // active, idle, running, error, archived
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (Project) TableName() string {
	return "projects"
}