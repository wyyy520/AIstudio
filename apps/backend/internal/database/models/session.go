package models

import "time"

type Session struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index:idx_sessions_user_expires;not null" json:"userId"`
	Token        string    `gorm:"uniqueIndex;size:512;not null" json:"-"`
	RefreshToken string    `gorm:"size:512" json:"-"`
	DeviceInfo   string    `gorm:"size:256" json:"deviceInfo"`
	IPAddress    string    `gorm:"size:64" json:"ipAddress"`
	LastAccessAt time.Time `json:"lastAccessAt"`
	ExpiresAt    time.Time `gorm:"index:idx_sessions_expires;index:idx_sessions_user_expires" json:"expiresAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (Session) TableName() string {
	return "sessions"
}