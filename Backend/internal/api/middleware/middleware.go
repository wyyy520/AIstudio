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
		secret = os.Getenv("JWT_SECRET")
	}
	if secret == "" {
		secret = "aistudio-default-secret-change-in-production"
	}
	auth.SetJWTSecret(secret)

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

	r.Use(Auth())
}
