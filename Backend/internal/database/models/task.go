package models

import "time"

// Task represents an asynchronous task (workflow run, training job, etc.).
type Task struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ProjectID   uint      `gorm:"index;not null" json:"projectId"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Description string    `gorm:"size:1024" json:"description"`
	Status      string    `gorm:"size:32;default:pending" json:"status"` // pending, running, success, failed, cancelled
	Priority    int       `gorm:"default:0" json:"priority"`             // 0=low,1=normal,2=high,3=urgent
	Result      string    `gorm:"type:text" json:"result,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (Task) TableName() string {
	return "tasks"
}