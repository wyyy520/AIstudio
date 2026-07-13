package models

import "time"

// TaskLog represents a persisted log entry for task execution.
// This replaces the in-memory-only log storage with database-backed persistence.
type TaskLog struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID    string    `gorm:"index:idx_task_logs_task;size:128;not null;default:''" json:"taskId"`
	Level     string    `gorm:"size:16;not null;index:idx_task_logs_level;default:INFO" json:"level"`
	Source    string    `gorm:"size:64;not null;index:idx_task_logs_source;default:''" json:"source"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	Detail    string    `gorm:"type:text" json:"detail,omitempty"`
	CreatedAt time.Time `gorm:"index:idx_task_logs_created_at" json:"timestamp"`
}

func (TaskLog) TableName() string {
	return "task_logs"
}