package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var DefaultAllowedOrigins = []string{
	"tauri://localhost",
	"https://tauri.localhost",
	"capacitor://localhost",
}

func resolveAllowedOrigins() []string {
	if raw := os.Getenv("CORS_ALLOWED_ORIGINS"); raw != "" {
		origins := strings.Split(raw, ",")
		for i, o := range origins {
			origins[i] = strings.TrimSpace(o)
		}
		return origins
	}

	origins := make([]string, len(DefaultAllowedOrigins))
	copy(origins, DefaultAllowedOrigins)

	env := os.Getenv("AISTUDIO_ENV")
	if env == "" {
		env = "development"
	}
	if env == "development" {
		origins = append(origins,
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:3000",
			"http://localhost:8080",
			"http://localhost",
		)
	}

	return origins
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposeHeaders    []string
	MaxAge           string
	AllowCredentials bool
}

func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: resolveAllowedOrigins(),
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
		MaxAge:           "86400",
		AllowCredentials: true,
	}
}

func originAllowed(origin string, allowed []string) string {
	if origin == "" {
		return ""
	}
	for _, allowedOrigin := range allowed {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return origin
		}
		if strings.HasPrefix(allowedOrigin, "*.") {
			suffix := allowedOrigin[1:]
			if strings.HasSuffix(origin, suffix) {
				return origin
			}
		}
	}
	return ""
}

func CORS() gin.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

func CORSWithConfig(cfg CORSConfig) gin.HandlerFunc {
	if len(cfg.AllowedOrigins) == 0 {
		cfg.AllowedOrigins = []string{"*"}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		allowedOrigin := originAllowed(origin, cfg.AllowedOrigins)
		if allowedOrigin == "" && len(cfg.AllowedOrigins) == 1 && cfg.AllowedOrigins[0] == "*" {
			allowedOrigin = "*"
		}

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

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
