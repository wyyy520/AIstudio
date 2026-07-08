package models

import "time"

// Task represents an asynchronous task (workflow run, training job, etc.).
// This is the GORM persistence model for the tasks table.
type Task struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	TaskID      string     `gorm:"uniqueIndex;size:128;not null" json:"task_id"`
	ProjectID   string     `gorm:"index;size:128" json:"project_id"`
	WorkflowID  string     `gorm:"index;size:128" json:"workflow_id"`
	Type        string     `gorm:"size:32" json:"type"`
	Name        string     `gorm:"size:256" json:"name"`
	Status      string     `gorm:"size:32;not null;default:waiting" json:"status"`
	Progress    float64    `gorm:"default:0" json:"progress"`
	Priority    int        `gorm:"default:1" json:"priority"`
	Handler     string     `gorm:"size:128" json:"handler"`
	Result      string     `gorm:"type:text" json:"result,omitempty"`
	Error       string     `gorm:"type:text" json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
}

func (Task) TableName() string {
	return "tasks"
}