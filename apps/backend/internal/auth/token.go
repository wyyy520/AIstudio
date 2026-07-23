package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

var jwtHeader = base64URLEncode([]byte(`{"alg":"HS256","typ":"JWT"}`))

var (
	jwtSecret     []byte
	jwtSecretOnce sync.Once
)

// DefaultJWTSecret is the fallback secret used only when no secret is configured.
// This MUST be overridden in production environments.
const DefaultJWTSecret = "aistudio-default-secret-change-in-production"

func SetJWTSecret(secret string) {
	jwtSecretOnce.Do(func() {
		if secret == "" {
			secret = DefaultJWTSecret
		}
		jwtSecret = []byte(secret)
	})
}

// MustNotUseDefaultSecret panics if the default JWT secret is still configured.
// Call this at startup in production to prevent insecure deployments.
func MustNotUseDefaultSecret() {
	if IsDefaultSecret() {
		panic("FATAL: JWT secret is using the default value. Set AISTUDIO_JWT_SECRET or jwt.secret config before starting the server.")
	}
}

// IsDefaultSecret returns true if the current JWT secret is the default (insecure) value.
func IsDefaultSecret() bool {
	return string(jwtSecret) == DefaultJWTSecret
}

func ResetJWTSecret() {
	jwtSecretOnce = sync.Once{}
	jwtSecret = nil
}

type AccessClaims struct {
	Sub      string `json:"sub"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
}

type RefreshClaims struct {
	Sub  string `json:"sub"`
	Exp  int64  `json:"exp"`
	Iat  int64  `json:"iat"`
	Type string `json:"type"`
}

type TokenManager struct {
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewTokenManager(accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (m *TokenManager) GenerateAccessToken(userID, username, role string) (string, error) {
	now := time.Now()
	claims := AccessClaims{
		Sub:      userID,
		Username: username,
		Role:     role,
		Exp:      now.Add(m.accessTTL).Unix(),
		Iat:      now.Unix(),
	}
	return encodeJWT(claims)
}

func (m *TokenManager) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := RefreshClaims{
		Sub:  userID,
		Exp:  now.Add(m.refreshTTL).Unix(),
		Iat:  now.Unix(),
		Type: "refresh",
	}
	return encodeJWT(claims)
}

func (m *TokenManager) ValidateAccessToken(tokenString string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	if err := decodeJWT(tokenString, claims); err != nil {
		return nil, err
	}
	if time.Now().Unix() > claims.Exp {
		return nil, ErrTokenExpired
	}
	return claims, nil
}

func (m *TokenManager) ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	if err := decodeJWT(tokenString, claims); err != nil {
		return nil, err
	}
	if claims.Type != "refresh" {
		return nil, ErrTokenInvalid
	}
	if time.Now().Unix() > claims.Exp {
		return nil, ErrTokenExpired
	}
	return claims, nil
}

func encodeJWT(claims interface{}) (string, error) {
	if jwtSecret == nil {
		SetJWTSecret("")
	}
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}
	payload := base64URLEncode(payloadBytes)
	signingInput := jwtHeader + "." + payload
	signature := sign(signingInput)
	return signingInput + "." + signature, nil
}

func decodeJWT(tokenString string, claims interface{}) error {
	if jwtSecret == nil {
		SetJWTSecret("")
	}
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return ErrTokenInvalid
	}
	signingInput := parts[0] + "." + parts[1]
	expectedSig := sign(signingInput)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return ErrTokenInvalid
	}
	payloadBytes, err := base64URLDecode(parts[1])
	if err != nil {
		return ErrTokenInvalid
	}
	if err := json.Unmarshal(payloadBytes, claims); err != nil {
		return ErrTokenInvalid
	}
	return nil
}

func sign(input string) string {
	mac := hmac.New(sha256.New, jwtSecret)
	mac.Write([]byte(input))
	return base64URLEncode(mac.Sum(nil))
}

func base64URLEncode(data []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(data), "=")
}

func base64URLDecode(s string) ([]byte, error) {
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}
