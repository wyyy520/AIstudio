package auth

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Authenticator struct {
	users    *UserManager
	tokens   *TokenManager
	sessions *SessionManager
	perms    *PermissionManager
	quotas   *QuotaManager
	apikeys  *APIKeyManager
}

type Manager struct {
	*Authenticator
	users *UserManager
}

func NewManager(db *gorm.DB, jwtSecret string, accessTTL, refreshTTL time.Duration) *Manager {
	users := NewUserManager(db)
	tokens := NewTokenManager(accessTTL, refreshTTL)
	sessions := NewSessionManager(db)
	perms := NewPermissionManager(db)
	quotas := NewQuotaManager(db)
	apikeys := NewAPIKeyManager(db, jwtSecret)

	authenticator := NewAuthenticator(users, tokens, sessions, perms, quotas, apikeys)
	return &Manager{
		Authenticator: authenticator,
		users:         users,
	}
}

func (m *Manager) UserManager() *UserManager {
	return m.users
}

func NewAuthenticator(
	users *UserManager,
	tokens *TokenManager,
	sessions *SessionManager,
	perms *PermissionManager,
	quotas *QuotaManager,
	apikeys *APIKeyManager,
) *Authenticator {
	return &Authenticator{
		users:    users,
		tokens:   tokens,
		sessions: sessions,
		perms:    perms,
		quotas:   quotas,
		apikeys:  apikeys,
	}
}

type LoginParams struct {
	Username   string
	Password   string
	DeviceInfo string
	IPAddress  string
}

type LoginResult struct {
	User         UserInfo `json:"user"`
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	ExpiresIn    int64    `json:"expiresIn"`
}

type UserInfo struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Role      string `json:"role"`
	Status    string `json:"status"`
}

func (a *Authenticator) Login(params LoginParams) (*LoginResult, error) {
	user, err := a.users.GetByUsername(params.Username)
	if err != nil {
		return nil, ErrInvalidCredential
	}

	if err := a.users.ValidatePassword(user, params.Password); err != nil {
		return nil, err
	}

	uid := fmt.Sprintf("%d", user.ID)

	accessToken, err := a.tokens.GenerateAccessToken(uid, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := a.tokens.GenerateRefreshToken(uid)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	sessionTTL := 7 * 24 * time.Hour
	_, err = a.sessions.Create(CreateSessionParams{
		UserID:       user.ID,
		Token:        accessToken,
		RefreshToken: refreshToken,
		DeviceInfo:   params.DeviceInfo,
		IPAddress:    params.IPAddress,
		TTL:          sessionTTL,
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	a.users.UpdateLastLogin(user.ID)

	accessTTL := time.Hour * 24

	return &LoginResult{
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Role:     user.Role,
			Status:   user.Status,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTTL.Seconds()),
	}, nil
}

func (a *Authenticator) Logout(token string) error {
	return a.sessions.DeleteByToken(token)
}

func (a *Authenticator) RefreshAccessToken(refreshToken string) (string, error) {
	_, err := a.tokens.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	session, err := a.sessions.GetByToken(refreshToken)
	if err != nil {
		return "", ErrSessionExpired
	}

	user, err := a.users.GetByID(session.UserID)
	if err != nil {
		return "", err
	}

	uid := fmt.Sprintf("%d", user.ID)
	accessToken, err := a.tokens.GenerateAccessToken(uid, user.Username, user.Role)
	if err != nil {
		return "", fmt.Errorf("generate access token: %w", err)
	}

	a.sessions.UpdateLastAccess(session.ID)

	return accessToken, nil
}

func (a *Authenticator) ValidateToken(token string) (*AccessClaims, error) {
	return a.tokens.ValidateAccessToken(token)
}

func (a *Authenticator) Users() *UserManager            { return a.users }
func (a *Authenticator) Tokens() *TokenManager           { return a.tokens }
func (a *Authenticator) Sessions() *SessionManager       { return a.sessions }
func (a *Authenticator) Permissions() *PermissionManager { return a.perms }
func (a *Authenticator) Quotas() *QuotaManager           { return a.quotas }
func (a *Authenticator) APIKeys() *APIKeyManager         { return a.apikeys }