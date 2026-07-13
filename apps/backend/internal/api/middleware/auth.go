package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/aistudio/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

const (
	AuthHeader   = "Authorization"
	BearerPrefix = "Bearer "
)

var publicPaths = map[string]bool{
	"/api/health":             true,
	"/api/auth/login":         true,
	"/api/auth/register":      true,
	"/api/environment/status": true,
	"/api/workflows/nodes":    true,
	"/api/plugins":            true,
	"/api/plugins/nodes":      true,
}

func isPublicPath(path string) bool {
	if publicPaths[path] {
		return true
	}
	if strings.HasPrefix(path, "/api/health") {
		return true
	}
	return false
}

func AddPublicPath(path string) {
	publicPaths[path] = true
}

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isPublicPath(c.Request.URL.Path) {
			c.Next()
			return
		}

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

		claims, err := validateAccessToken(tokenString)
		if err != nil {
			status := http.StatusUnauthorized
			msg := err.Error()
			if strings.Contains(msg, "expired") {
				msg = "token expired, please login again"
			}
			c.AbortWithStatusJSON(status, gin.H{
				"code":    -1,
				"message": msg,
			})
			return
		}

		c.Set("userID", claims.Sub)
		c.Set("username", claims.Username)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (string, bool) {
	id, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	return id.(string), true
}

func GetUsername(c *gin.Context) (string, bool) {
	name, exists := c.Get("username")
	if !exists {
		return "", false
	}
	return name.(string), true
}

func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("userRole")
	if !exists {
		return "", false
	}
	return role.(string), true
}

func validateAccessToken(tokenString string) (*auth.AccessClaims, error) {
	mgr := auth.NewTokenManager(0, 0)
	return mgr.ValidateAccessToken(tokenString)
}

func GenerateToken(userID, username string, ttl time.Duration) (string, error) {
	mgr := auth.NewTokenManager(ttl, 0)
	return mgr.GenerateAccessToken(userID, username, "")
}

func ValidateToken(tokenString string) (*auth.AccessClaims, error) {
	mgr := auth.NewTokenManager(0, 0)
	return mgr.ValidateAccessToken(tokenString)
}

func SetJWTSecret(secret string) {
	auth.SetJWTSecret(secret)
}

func ResetJWTSecret() {
	auth.ResetJWTSecret()
}
