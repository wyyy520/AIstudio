package models

import "time"

// Workflow represents an AI workflow definition.
type Workflow struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ProjectID  uint      `gorm:"index;not null" json:"projectId"`
	Name       string    `gorm:"size:128;not null" json:"name"`
	Definition string    `gorm:"type:text" json:"definition"` // JSON definition of the workflow DAG
	Status     string    `gorm:"size:32;default:draft" json:"status"` // draft, active, running, completed, failed
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (Workflow) TableName() string {
	return "workflows"
}