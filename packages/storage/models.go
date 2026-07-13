package storage

import "time"

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Email        string    `gorm:"uniqueIndex;size:128;not null" json:"email"`
	PasswordHash string    `gorm:"column:password_hash;size:256;not null" json:"-"`
	Role         string    `gorm:"size:32;default:user;not null;index" json:"role"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (User) TableName() string { return "users" }

type Session struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"userId"`
	Token        string    `gorm:"uniqueIndex;size:512;not null" json:"-"`
	RefreshToken string    `gorm:"size:512" json:"-"`
	ExpiresAt    time.Time `gorm:"index" json:"expiresAt"`
	IP           string    `gorm:"size:64" json:"ip"`
	UserAgent    string    `gorm:"size:256" json:"userAgent"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (Session) TableName() string { return "sessions" }

type APIKey struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"index;not null" json:"userId"`
	Name       string     `gorm:"size:64" json:"name"`
	Key        string     `gorm:"size:512;not null" json:"-"`
	Provider   string     `gorm:"size:32;not null;index" json:"provider"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	IsActive   bool       `gorm:"default:true;index" json:"isActive"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

func (APIKey) TableName() string { return "api_keys" }

type Permission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"userId"`
	Resource  string    `gorm:"size:64;not null" json:"resource"`
	Action    string    `gorm:"size:32;not null" json:"action"`
	Effect    string    `gorm:"size:16;default:allow" json:"effect"`
	CreatedAt time.Time `json:"createdAt"`
}

func (Permission) TableName() string { return "permissions" }

type Quota struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"userId"`
	Resource  string    `gorm:"size:64;not null" json:"resource"`
	Total     int64     `gorm:"default:-1" json:"total"`
	Used      int64     `gorm:"default:0" json:"used"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Quota) TableName() string { return "quotas" }

type Project struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"userId"`
	Name         string    `gorm:"size:128;not null" json:"name"`
	Description  string    `gorm:"size:1024" json:"description"`
	Path         string    `gorm:"size:1024" json:"path"`
	Language     string    `gorm:"size:32" json:"language"`
	Status       string    `gorm:"size:32;default:active;index" json:"status"`
	Tags         string    `gorm:"size:512" json:"tags"`
	LastOpenedAt *time.Time `json:"lastOpenedAt"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (Project) TableName() string { return "projects" }

type Plugin struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	Version   string    `gorm:"size:32;not null" json:"version"`
	Author    string    `gorm:"size:128" json:"author"`
	Type      string    `gorm:"size:32" json:"type"`
	Status    string    `gorm:"size:32;default:not_installed;index" json:"status"`
	Enabled   bool      `gorm:"default:false" json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
}

func (Plugin) TableName() string { return "plugins" }

type Task struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"index;not null" json:"userId"`
	Type        string     `gorm:"size:32;index" json:"type"`
	Status      string     `gorm:"size:32;not null;default:pending;index" json:"status"`
	Priority    int        `gorm:"default:1" json:"priority"`
	Progress    float64    `gorm:"default:0" json:"progress"`
	CreatedAt   time.Time  `gorm:"index" json:"createdAt"`
	StartedAt   *time.Time `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt"`
}

func (Task) TableName() string { return "tasks" }

type TaskLog struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID    string    `gorm:"index;size:128;not null" json:"taskId"`
	Level     string    `gorm:"size:16;not null;default:INFO" json:"level"`
	Source    string    `gorm:"size:64" json:"source"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	Detail    string    `gorm:"type:text" json:"detail"`
	CreatedAt time.Time `gorm:"index" json:"timestamp"`
}

func (TaskLog) TableName() string { return "task_logs" }
