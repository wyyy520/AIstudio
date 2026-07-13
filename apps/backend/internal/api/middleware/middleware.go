package middleware

import (
	"os"

	"github.com/aistudio/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

type Config struct {
	JWTSecret   string
	RateLimit   *RateLimitConfig
	CORS        *CORSConfig
	Development bool
}

func DefaultConfig() Config {
	return Config{
		JWTSecret:   "",
		RateLimit:   nil,
		CORS:        nil,
		Development: false,
	}
}

func Apply(r *gin.Engine, cfg Config) {
	secret := cfg.JWTSecret
	if secret == "" {
		secret = os.Getenv("AISTUDIO_JWT_SECRET")
	}
	if secret == "" {
		secret = os.Getenv("JWT_SECRET")
	}
	auth.SetJWTSecret(secret)

	if auth.IsDefaultSecret() {
		r.Use(func(c *gin.Context) {
			if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" {
				c.Header("X-Security-Warning", "JWT secret is using default value. Set AISTUDIO_JWT_SECRET environment variable for production.")
			}
			c.Next()
		})
	}

	if cfg.Development {
		r.Use(RecoveryWithConfig(RecoveryConfig{
			PrintStack:  true,
			LogEntry:    true,
			Development: true,
		}))
	} else {
		r.Use(Recovery())
	}

	r.Use(Logger())

	if cfg.CORS != nil {
		r.Use(CORSWithConfig(*cfg.CORS))
	} else {
		r.Use(CORS())
	}

	if cfg.RateLimit != nil {
		r.Use(RateLimitWithConfig(*cfg.RateLimit))
	} else {
		r.Use(RateLimit())
	}

	if !cfg.Development {
		r.Use(Auth())
	}

	r.Use(func(c *gin.Context) {
		c.Set("userID", "dev-user")
		c.Set("username", "developer")
		c.Set("userRole", "admin")
		c.Next()
	})}