package models

import "time"

type Permission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Role      string    `gorm:"size:32;not null;index" json:"role"`
	Resource  string    `gorm:"size:64;not null" json:"resource"`
	Action    string    `gorm:"size:32;not null" json:"action"`
	CreatedAt time.Time `json:"createdAt"`
}

func (Permission) TableName() string {
	return "permissions"
}
