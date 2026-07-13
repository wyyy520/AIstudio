package security

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	jwtSecret     []byte
	jwtSecretOnce sync.Once
)

const DefaultJWTSecret = "aistudio-default-secret-change-in-production"

type AccessClaims struct {
	Sub      string `json:"sub"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	Sub  string `json:"sub"`
	Type string `json:"type"`
	jwt.RegisteredClaims
}

func SetJWTSecret(secret string) {
	jwtSecretOnce.Do(func() {
		if secret == "" {
			secret = DefaultJWTSecret
		}
		jwtSecret = []byte(secret)
	})
}

func IsDefaultSecret() bool {
	return string(jwtSecret) == DefaultJWTSecret
}

func ResetJWTSecret() {
	jwtSecretOnce = sync.Once{}
	jwtSecret = nil
}

func GenerateAccessToken(userID, username, role string, ttl time.Duration) (string, error) {
	if jwtSecret == nil {
		SetJWTSecret("")
	}
	now := time.Now()
	claims := AccessClaims{
		Sub:      userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return tokenStr, nil
}

func GenerateRefreshToken(userID string, ttl time.Duration) (string, error) {
	if jwtSecret == nil {
		SetJWTSecret("")
	}
	now := time.Now()
	claims := RefreshClaims{
		Sub:  userID,
		Type: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign refresh token: %w", err)
	}
	return tokenStr, nil
}

func ValidateAccessToken(tokenString string) (*AccessClaims, error) {
	if jwtSecret == nil {
		SetJWTSecret("")
	}
	claims := &AccessClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("validate access token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid access token")
	}
	return claims, nil
}

func ValidateRefreshToken(tokenString string) (*RefreshClaims, error) {
	if jwtSecret == nil {
		SetJWTSecret("")
	}
	claims := &RefreshClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("validate refresh token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}
	if claims.Type != "refresh" {
		return nil, fmt.Errorf("invalid token type: %s", claims.Type)
	}
	return claims, nil
}
