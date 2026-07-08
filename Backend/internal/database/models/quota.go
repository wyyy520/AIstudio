package models

import "time"

type Quota struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"userId"`
	ResourceType string    `gorm:"size:64;not null;index" json:"resourceType"`
	Limit        int64     `gorm:"default:-1" json:"limit"`
	Used         int64     `gorm:"default:0" json:"used"`
	PeriodStart  time.Time `json:"periodStart"`
	PeriodEnd    time.Time `json:"periodEnd"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (Quota) TableName() string {
	return "quotas"
}
