package security

import (
	"fmt"
	"time"
)

type SessionStore interface {
	Create(sessionID, userID, token, refreshToken string, expiresAt time.Time) error
	GetByToken(token string) (userID string, expiresAt time.Time, err error)
	DeleteByToken(token string) error
	DeleteByUserID(userID string) error
}

type SessionManager struct {
	store  SessionStore
	secret string
}

func NewSessionManager(store SessionStore, secret string) *SessionManager {
	return &SessionManager{store: store, secret: secret}
}

func (m *SessionManager) CreateSession(userID, deviceInfo, ipAddress string, ttl time.Duration) (accessToken, refreshToken string, err error) {
	accessToken, err = GenerateAccessToken(userID, "", "", ttl)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}
	refreshTTL := ttl * 2
	refreshToken, err = GenerateRefreshToken(userID, refreshTTL)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	sessionID, err := RandomHex(16)
	if err != nil {
		return "", "", fmt.Errorf("session id: %w", err)
	}
	if err := m.store.Create(sessionID, userID, accessToken, refreshToken, time.Now().Add(ttl)); err != nil {
		return "", "", fmt.Errorf("store session: %w", err)
	}
	return accessToken, refreshToken, nil
}

func (m *SessionManager) ValidateSession(token string) (string, error) {
	userID, expiresAt, err := m.store.GetByToken(token)
	if err != nil {
		return "", fmt.Errorf("session not found: %w", err)
	}
	if time.Now().After(expiresAt) {
		m.store.DeleteByToken(token)
		return "", fmt.Errorf("session expired")
	}
	return userID, nil
}

func (m *SessionManager) DeleteSession(token string) error {
	return m.store.DeleteByToken(token)
}

func (m *SessionManager) DeleteUserSessions(userID string) error {
	return m.store.DeleteByUserID(userID)
}

func (m *SessionManager) RefreshSession(refreshToken string, ttl time.Duration) (string, error) {
	claims, err := ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}
	accessToken, err := GenerateAccessToken(claims.Sub, "", "", ttl)
	if err != nil {
		return "", fmt.Errorf("generate access token: %w", err)
	}
	return accessToken, nil
}
