package models

import "time"

// Project represents an AI project.
type Project struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Description string    `gorm:"size:1024" json:"description"`
	OwnerID     uint      `gorm:"index:idx_projects_owner_status;not null" json:"ownerId"`
	Status      string    `gorm:"size:32;default:active;index:idx_projects_status;index:idx_projects_owner_status" json:"status"` // active, idle, running, error, archived
	CreatedAt   time.Time `gorm:"index:idx_projects_created_at" json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (Project) TableName() string {
	return "projects"
}