package models

import "time"

type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Email        string     `gorm:"uniqueIndex;size:128;not null" json:"email"`
	PasswordHash string     `gorm:"column:password_hash;size:256;not null" json:"-"`
	Nickname     string     `gorm:"size:64" json:"nickname"`
	Avatar       string     `gorm:"size:512" json:"avatar"`
	Role         string     `gorm:"size:32;default:user;not null;index" json:"role"`
	Status       string     `gorm:"size:32;default:active;not null;index" json:"status"`
	LastLoginAt  *time.Time `json:"lastLoginAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func (User) TableName() string {
	return "users"
}
