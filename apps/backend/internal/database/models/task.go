package models

import "time"

// Task represents an asynchronous task (workflow run, training job, etc.).
// This is the GORM persistence model for the tasks table.
type Task struct {
	ID          uint       `gorm:"primaryKey" json:"-"`
	TaskID      string     `gorm:"uniqueIndex;size:128;not null" json:"id"`
	ProjectID   string     `gorm:"index:idx_tasks_project;size:128" json:"projectId"`
	WorkflowID  string     `gorm:"index:idx_tasks_workflow;size:128" json:"workflowId"`
	Type        string     `gorm:"size:32;index:idx_tasks_type_status" json:"type"`
	Name        string     `gorm:"size:256" json:"name"`
	Status      string     `gorm:"size:32;not null;default:waiting;index:idx_tasks_status;index:idx_tasks_type_status" json:"status"`
	Progress    float64    `gorm:"default:0" json:"progress"`
	Priority    int        `gorm:"default:1;index:idx_tasks_priority" json:"priority"`
	Handler     string     `gorm:"size:128" json:"handler"`
	Result      string     `gorm:"type:text" json:"result,omitempty"`
	Error       string     `gorm:"type:text" json:"error,omitempty"`
	CreatedAt   time.Time  `gorm:"index:idx_tasks_created_at" json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	StartTime   *time.Time `json:"startedAt,omitempty"`
	EndTime     *time.Time `json:"completedAt,omitempty"`
}

func (Task) TableName() string {
	return "tasks"
}