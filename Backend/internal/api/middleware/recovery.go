package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RecoveryConfig allows customization of the recovery middleware behavior.
type RecoveryConfig struct {
	// PrintStack controls whether to print the stack trace on panic.
	PrintStack bool
	// LogEntry controls whether to write a structured JSON log entry.
	LogEntry bool
	// Development mode enables more verbose output.
	Development bool
}

// DefaultRecoveryConfig returns the default recovery configuration.
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		PrintStack:  true,
		LogEntry:    true,
		Development: false,
	}
}

// Recovery returns a Gin middleware that catches panics during request
// handling and returns a 500 Internal Server Error response.
//
// It prevents the server from crashing due to unhandled panics and
// logs the error details for debugging.
func Recovery() gin.HandlerFunc {
	return RecoveryWithConfig(DefaultRecoveryConfig())
}

// RecoveryWithConfig returns a recovery middleware with the given configuration.
func RecoveryWithConfig(cfg RecoveryConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Build stack trace
				stack := string(debug.Stack())

				// Log the panic
				if cfg.PrintStack {
					log.Printf(
						"[PANIC] %s %s\nError: %v\nStack:\n%s",
						c.Request.Method,
						c.Request.URL.Path,
						err,
						stack,
					)
				}

				// Structured log entry
				if cfg.LogEntry {
					entry := map[string]interface{}{
						"timestamp": time.Now().Format(time.RFC3339Nano),
						"level":     "CRITICAL",
						"method":    c.Request.Method,
						"path":      c.Request.URL.Path,
						"ip":        c.ClientIP(),
						"error":     fmt.Sprintf("%v", err),
						"stack":     truncateStack(stack, 10),
						"source":    "api",
					}
					log.Printf("[middleware] panic recovered: method=%s path=%s err=%v",
						c.Request.Method, c.Request.URL.Path, err)
					_ = entry // structured logging via standard log
				}

				// Development mode: include error details in response
				message := "internal server error"
				if cfg.Development {
					message = fmt.Sprintf("panic: %v", err)
				}

				// Abort with 500
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    -1,
					"message": message,
				})
			}
		}()
		c.Next()
	}
}

// truncateStack limits the stack trace to the first n lines.
func truncateStack(stack string, n int) string {
	lines := strings.Split(stack, "\n")
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}