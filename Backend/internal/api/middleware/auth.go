package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// JWT header and payload constants
const (
	AuthHeader = "Authorization"
	BearerPrefix = "Bearer "
)

// --- JWT Implementation (HS256, no external dependencies) ---

// jwtHeader is the base64url-encoded JWT header.
var jwtHeader = base64URLEncode([]byte(`{"alg":"HS256","typ":"JWT"}`))

// jwtSecret is the HMAC key used for signing tokens.
var (
	jwtSecret     []byte
	jwtSecretOnce sync.Once
)

// JWTClaims represents the standard JWT claims.
type JWTClaims struct {
	Sub      string `json:"sub"`      // user identifier
	Username string `json:"username"` // custom: username
	Exp      int64  `json:"exp"`      // expiration timestamp
	Iat      int64  `json:"iat"`      // issued at
}

// SetJWTSecret sets the JWT signing secret (called during middleware init).
func SetJWTSecret(secret string) {
	jwtSecretOnce.Do(func() {
		if secret == "" {
			secret = "aistudio-default-secret-change-in-production"
		}
		jwtSecret = []byte(secret)
	})
}

// ResetJWTSecret resets the JWT secret state for testing purposes.
func ResetJWTSecret() {
	jwtSecretOnce = sync.Once{}
	jwtSecret = nil
}

// GenerateToken creates a signed JWT token string.
// A non-positive TTL creates an already-expired token (useful for testing).
func GenerateToken(userID, username string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		Sub:      userID,
		Username: username,
		Exp:      now.Add(ttl).Unix(),
		Iat:      now.Unix(),
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

// ValidateToken parses and validates a JWT token.
// Returns the claims if valid.
func ValidateToken(tokenString string) (*JWTClaims, error) {
	if jwtSecret == nil {
		SetJWTSecret("")
	}

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	// Verify signature
	signingInput := parts[0] + "." + parts[1]
	expectedSig := sign(signingInput)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, errors.New("invalid token signature")
	}

	// Decode payload
	payloadBytes, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("unmarshal claims: %w", err)
	}

	// Check expiration
	if time.Now().Unix() > claims.Exp {
		return nil, errors.New("token expired")
	}

	return &claims, nil
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
	// Add padding
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.URLEncoding.DecodeString(s)
}

// --- Auth Middleware ---

// publicPaths are routes that do not require authentication.
var publicPaths = map[string]bool{
	"/api/health":          true,
	"/api/auth/login":      true,
	"/api/environment/status": true,
}

// isPublicPath checks if a path is excluded from authentication.
func isPublicPath(path string) bool {
	// Exact match
	if publicPaths[path] {
		return true
	}
	// Prefix match for OPTIONS preflight
	if strings.HasPrefix(path, "/api/health") {
		return true
	}
	return false
}

// AddPublicPath registers an additional route as public (no auth required).
// Intended for tests that register ad-hoc routes.
func AddPublicPath(path string) {
	publicPaths[path] = true
}

// Auth returns a Gin middleware that enforces JWT Bearer authentication.
// Routes in publicPaths are allowed without a token.
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for public paths
		if isPublicPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Skip for OPTIONS (CORS preflight)
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		authHeader := c.GetHeader(AuthHeader)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "missing authorization header",
			})
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "invalid authorization format, expected 'Bearer <token>'",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := ValidateToken(tokenString)
		if err != nil {
			status := http.StatusUnauthorized
			msg := err.Error()
			if strings.Contains(msg, "expired") {
				status = http.StatusUnauthorized
				msg = "token expired, please login again"
			}
			c.AbortWithStatusJSON(status, gin.H{
				"code":    -1,
				"message": msg,
			})
			return
		}

		// Store user info in context for downstream handlers
		c.Set("userID", claims.Sub)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// GetUserID extracts the authenticated user ID from the Gin context.
func GetUserID(c *gin.Context) (string, bool) {
	id, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	return id.(string), true
}

// GetUsername extracts the authenticated username from the Gin context.
func GetUsername(c *gin.Context) (string, bool) {
	name, exists := c.Get("username")
	if !exists {
		return "", false
	}
	return name.(string), true
}

// Ensure JWT functions are accessible.
var _ = GenerateToken