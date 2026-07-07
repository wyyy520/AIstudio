package middleware

import (
	"os"

	"github.com/gin-gonic/gin"
)

// Config holds all middleware configuration options.
type Config struct {
	// JWTSecret is the key used for signing JWT tokens.
	// If empty, a default development secret is used.
	JWTSecret string

	// RateLimit configures the per-IP rate limiting.
	// If nil, rate limiting uses default values (100 req/min/IP).
	RateLimit *RateLimitConfig

	// CORS configures Cross-Origin Resource Sharing.
	// If nil, CORS uses default origins (localhost + Tauri).
	CORS *CORSConfig

	// Development mode enables verbose error messages in responses.
	Development bool
}

// DefaultConfig returns the default middleware configuration.
func DefaultConfig() Config {
	return Config{
		JWTSecret:   "",
		RateLimit:   nil, // uses defaults
		CORS:        nil, // uses defaults
		Development: false,
	}
}

// Apply registers all middleware on the Gin router using the provided config.
//
// Order of middleware execution:
// 1. Recovery    - catch panics early
// 2. Logger      - log all requests
// 3. CORS        - handle CORS headers early
// 4. RateLimit   - throttle requests
// 5. Auth        - authenticate requests
//
// Usage:
//
//	import "github.com/aistudio/backend/internal/api/middleware"
//
//	cfg := middleware.DefaultConfig()
//	cfg.JWTSecret = os.Getenv("JWT_SECRET")
//	cfg.Development = true
//	middleware.Apply(router, cfg)
func Apply(r *gin.Engine, cfg Config) {
	// Initialize JWT secret
	secret := cfg.JWTSecret
	if secret == "" {
		secret = os.Getenv("JWT_SECRET")
	}
	SetJWTSecret(secret)

	// 1. Recovery (outermost: catches panics from all inner middleware)
	if cfg.Development {
		r.Use(RecoveryWithConfig(RecoveryConfig{
			PrintStack:  true,
			LogEntry:    true,
			Development: true,
		}))
	} else {
		r.Use(Recovery())
	}

	// 2. Logger (logs every request)
	r.Use(Logger())

	// 3. CORS (handle preflight before auth)
	if cfg.CORS != nil {
		r.Use(CORSWithConfig(*cfg.CORS))
	} else {
		r.Use(CORS())
	}

	// 4. Rate Limit (throttle before auth to reduce auth server load)
	if cfg.RateLimit != nil {
		r.Use(RateLimitWithConfig(*cfg.RateLimit))
	} else {
		r.Use(RateLimit())
	}

	// 5. Auth (innermost: validates JWT tokens)
	r.Use(Auth())
}