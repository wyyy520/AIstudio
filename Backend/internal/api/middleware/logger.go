package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// LogEntry represents a structured log entry for API requests.
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	Duration  string `json:"duration_ms"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Error     string `json:"error,omitempty"`
	Source    string `json:"source"`
}

// loggerOutput is the output writer for structured logs.
// Defaults to stdout, can be overridden for testing.
var loggerOutput = os.Stdout

// SetLoggerOutput allows overriding the output writer.
func SetLoggerOutput(w *os.File) {
	loggerOutput = w
}

// Logger returns a Gin middleware that logs structured JSON entries
// for every HTTP request. It records:
//   - Timestamp (RFC3339)
//   - Request method and path
//   - Response status code
//   - Request duration in milliseconds
//   - Client IP address
//   - User agent
//   - Authenticated user ID (if available)
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Process request
		c.Next()

		// Post-request data
		duration := time.Since(start)
		status := c.Writer.Status()
		durationMs := float64(duration.Microseconds()) / 1000.0

		entry := LogEntry{
			Timestamp: time.Now().Format(time.RFC3339Nano),
			Level:     logLevelFromStatus(status),
			Method:    method,
			Path:      path,
			Status:    status,
			Duration:  fmt.Sprintf("%.2f", durationMs),
			IP:        clientIP,
			UserAgent: userAgent,
			Source:    "api",
		}

		// Include user ID if authenticated
		if userID, exists := c.Get("userID"); exists {
			if uid, ok := userID.(string); ok {
				entry.UserID = uid
			}
		}

		// Include error message if present
		if len(c.Errors) > 0 {
			errMsgs := make([]string, len(c.Errors))
			for i, err := range c.Errors {
				errMsgs[i] = err.Error()
			}
			entry.Error = strings.Join(errMsgs, "; ")
		}

		// Output structured JSON log
		line, err := json.Marshal(entry)
		if err != nil {
			log.Printf("[middleware] failed to marshal log entry: %v", err)
			return
		}

		fmt.Fprintln(loggerOutput, string(line))
	}
}

// logLevelFromStatus maps HTTP status codes to log levels.
func logLevelFromStatus(status int) string {
	switch {
	case status >= 500:
		return "ERROR"
	case status >= 400:
		return "WARN"
	case status >= 300:
		return "INFO"
	default:
		return "INFO"
	}
}