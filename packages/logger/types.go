package logger

import "time"

type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelFatal LogLevel = "FATAL"
)

var AllLevels = []LogLevel{LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError, LogLevelFatal}

type LogEntry struct {
	Timestamp time.Time      `json:"timestamp"`
	Level     LogLevel       `json:"level"`
	Source    string         `json:"source"`
	Message   string         `json:"message"`
	Raw       string         `json:"raw,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	TaskID    string         `json:"taskId,omitempty"`
	RunID     string         `json:"runId,omitempty"`
}

type Filter struct {
	Levels  []LogLevel `json:"levels,omitempty"`
	Sources []string   `json:"sources,omitempty"`
	TaskID  string     `json:"taskId,omitempty"`
	RunID   string     `json:"runId,omitempty"`
	Search  string     `json:"search,omitempty"`
	Since   *time.Time `json:"since,omitempty"`
	Until   *time.Time `json:"until,omitempty"`
	Limit   int        `json:"limit,omitempty"`
	Offset  int        `json:"offset,omitempty"`
}
