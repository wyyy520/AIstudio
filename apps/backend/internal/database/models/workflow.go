package models

import "time"

// Workflow represents an AI workflow definition.
// Only metadata is stored in the database — the actual workflow data
// lives in workflow.json on the filesystem (Single Source of Truth).
type Workflow struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ProjectID     uint      `gorm:"index:idx_workflows_project;not null" json:"projectId"`
	Name          string    `gorm:"size:128;not null" json:"name"`
	SchemaVersion string    `gorm:"size:32;default:2.0.0" json:"schemaVersion"`
	Version       int       `gorm:"default:1" json:"version"`
	Tags          string    `gorm:"size:512" json:"tags,omitempty"` // comma-separated tags
	Path          string    `gorm:"size:1024;not null" json:"path"` // absolute path to workflow.json
	Status        string    `gorm:"size:32;default:active;index:idx_workflows_status" json:"status"` // active, archived
	LastOpenedAt  time.Time `json:"lastOpenedAt,omitempty"`
	CreatedAt     time.Time `gorm:"index:idx_workflows_created_at" json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (Workflow) TableName() string {
	return "workflows"
}