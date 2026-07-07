package models

import "time"

// User represents a platform user.
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Email     string    `gorm:"uniqueIndex;size:128;not null" json:"email"`
	Password  string    `gorm:"size:256;not null" json:"-"` // never expose in JSON
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (User) TableName() string {
	return "users"
}