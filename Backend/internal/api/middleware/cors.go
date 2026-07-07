package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// DefaultAllowedOrigins lists the allowed origins.
// Includes localhost development and Tauri desktop app.
var DefaultAllowedOrigins = []string{
	"http://localhost:5173",  // Vite dev server
	"http://localhost:3000",  // Alternative dev port
	"http://localhost:8080",  // Alternative dev port
	"tauri://localhost",      // Tauri desktop app
	"https://tauri.localhost", // Tauri production
	"capacitor://localhost",   // Capacitor mobile app
	"http://localhost",        // Generic localhost
}

// CORSConfig allows customization of CORS behavior.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposeHeaders    []string
	MaxAge           string
	AllowCredentials bool
}

// DefaultCORSConfig returns the default CORS configuration.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: DefaultAllowedOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-CSRF-Token",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"Authorization",
		},
		MaxAge:           "86400", // 24 hours
		AllowCredentials: true,
	}
}

// originAllowed checks if the given origin is in the allowed list.
// Returns the origin if allowed, or an empty string if not.
func originAllowed(origin string, allowed []string) string {
	if origin == "" {
		return ""
	}
	for _, allowedOrigin := range allowed {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return origin
		}
		// Wildcard subdomain support: http://*.example.com
		if strings.HasPrefix(allowedOrigin, "*.") {
			suffix := allowedOrigin[1:] // skip "*"
			if strings.HasSuffix(origin, suffix) {
				return origin
			}
		}
	}
	return ""
}

// CORS handles Cross-Origin Resource Sharing with configurable origins.
// Supports localhost development and Tauri desktop app origins.
func CORS() gin.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns a CORS middleware with the given configuration.
func CORSWithConfig(cfg CORSConfig) gin.HandlerFunc {
	// Default to wildcard if no origins specified
	if len(cfg.AllowedOrigins) == 0 {
		cfg.AllowedOrigins = []string{"*"}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Determine the allowed origin
		allowedOrigin := originAllowed(origin, cfg.AllowedOrigins)
		if allowedOrigin == "" && len(cfg.AllowedOrigins) == 1 && cfg.AllowedOrigins[0] == "*" {
			allowedOrigin = "*"
		}

		// Set CORS headers
		if allowedOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowedOrigin)
		}

		if cfg.AllowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))

		if len(cfg.ExposeHeaders) > 0 {
			c.Header("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ", "))
		}

		if cfg.MaxAge != "" {
			c.Header("Access-Control-Max-Age", cfg.MaxAge)
		}

		// Handle preflight
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent) // 204
			return
		}

		c.Next()
	}
}