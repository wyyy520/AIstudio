package auth

import (
	"fmt"
	"time"

	"github.com/aistudio/backend/internal/database/models"
	"gorm.io/gorm"
)

type SessionManager struct {
	db *gorm.DB
}

func NewSessionManager(db *gorm.DB) *SessionManager {
	return &SessionManager{db: db}
}

type CreateSessionParams struct {
	UserID       uint
	Token        string
	RefreshToken string
	DeviceInfo   string
	IPAddress    string
	TTL          time.Duration
}

func (m *SessionManager) Create(params CreateSessionParams) (*models.Session, error) {
	now := time.Now()
	session := &models.Session{
		UserID:       params.UserID,
		Token:        params.Token,
		RefreshToken: params.RefreshToken,
		DeviceInfo:   params.DeviceInfo,
		IPAddress:    params.IPAddress,
		LastAccessAt: now,
		ExpiresAt:    now.Add(params.TTL),
	}
	if err := m.db.Create(session).Error; err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return session, nil
}

func (m *SessionManager) GetByToken(token string) (*models.Session, error) {
	var session models.Session
	if err := m.db.Where("token = ?", token).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSessionExpired
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	if time.Now().After(session.ExpiresAt) {
		m.db.Delete(&session)
		return nil, ErrSessionExpired
	}
	return &session, nil
}

func (m *SessionManager) GetByUserID(userID uint) ([]models.Session, error) {
	var sessions []models.Session
	if err := m.db.Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Order("last_access_at DESC").Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("get user sessions: %w", err)
	}
	return sessions, nil
}

func (m *SessionManager) UpdateLastAccess(sessionID uint) error {
	return m.db.Model(&models.Session{}).Where("id = ?", sessionID).
		Update("last_access_at", time.Now()).Error
}

func (m *SessionManager) DeleteByToken(token string) error {
	result := m.db.Where("token = ?", token).Delete(&models.Session{})
	if result.Error != nil {
		return fmt.Errorf("delete session: %w", result.Error)
	}
	return nil
}

func (m *SessionManager) DeleteByUserID(userID uint) error {
	return m.db.Where("user_id = ?", userID).Delete(&models.Session{}).Error
}

func (m *SessionManager) CleanupExpired() error {
	return m.db.Where("expires_at < ?", time.Now()).Delete(&models.Session{}).Error
}
