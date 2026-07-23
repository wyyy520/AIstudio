package service

import (
	"fmt"
	"log"
	"time"

	"github.com/aistudio/backend/internal/logcenter"
)

// LogLevel represents a log severity level.
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"

	DefaultPageSize = 20
	MaxPageSize     = 100
)

// LogEntry represents a single log entry (backward compatible).
type LogEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Level     LogLevel  `json:"level"`
	Source    string    `json:"source"`
	Message   string    `json:"message"`
	TaskID    string    `json:"taskId,omitempty"`
	Detail    string    `json:"detail,omitempty"`
}

// LogQuery represents filters for querying logs.
type LogQuery struct {
	Level   LogLevel `json:"level"`
	Source  string   `json:"source"`
	TaskID  string   `json:"taskId"`
	Start   string   `json:"start"`   // RFC3339
	End     string   `json:"end"`     // RFC3339
	Keyword string   `json:"keyword"` // search in message
	Page    int      `json:"page"`
	Size    int      `json:"size"`
}

// LogQueryResult represents paginated log results.
type LogQueryResult struct {
	Items []LogEntry `json:"items"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
}

// LogService provides application-level log management.
// Wraps the new logcenter.LogCenter for backward compatibility.
type LogService struct {
	center logcenter.Logger
}

// NewLogService creates a new LogService wrapping the LogCenter.
func NewLogService(center logcenter.Logger) *LogService {
	log.Printf("[log-service] initializing with LogCenter backend")
	return &LogService{center: center}
}

// Write adds a new log entry.
func (s *LogService) Write(level LogLevel, source, message string, opts ...LogOption) {
	entry := logcenter.Entry{
		Timestamp: time.Now(),
		Source:    source,
		Message:   message,
	}

	switch level {
	case LogLevelDebug:
		entry.Level = logcenter.LevelDebug
	case LogLevelInfo:
		entry.Level = logcenter.LevelInfo
	case LogLevelWarn:
		entry.Level = logcenter.LevelWarn
	case LogLevelError:
		entry.Level = logcenter.LevelError
	default:
		entry.Level = logcenter.LevelInfo
	}

	for _, opt := range opts {
		opt(&entry)
	}

	s.center.LogEntry(entry)
}

// Query returns filtered and paginated log entries.
func (s *LogService) Query(q LogQuery) (*LogQueryResult, error) {
	filter := logcenter.Filter{
		Search: q.Keyword,
		TaskID: q.TaskID,
		Limit:  q.Size,
		Offset: (q.Page - 1) * q.Size,
	}

	if q.Level != "" {
		filter.Levels = []logcenter.Level{logcenter.Level(q.Level)}
	}
	if q.Source != "" {
		filter.Sources = []string{q.Source}
	}
	if q.Start != "" {
		t, err := time.Parse(time.RFC3339, q.Start)
		if err == nil {
			filter.Since = &t
		}
	}
	if q.End != "" {
		t, err := time.Parse(time.RFC3339, q.End)
		if err == nil {
			filter.Until = &t
		}
	}

	entries := s.center.Query(filter)
	total := len(entries)

	// Convert to backward-compatible LogEntry
	items := make([]LogEntry, 0, len(entries))
	for i, e := range entries {
		items = append(items, LogEntry{
			ID:        int64(i + 1),
			Timestamp: e.Timestamp,
			Level:     LogLevel(e.Level),
			Source:    e.Source,
			Message:   e.Message,
			TaskID:    e.TaskID,
			Detail:    e.Detail,
		})
	}

	if q.Size <= 0 {
		q.Size = DefaultPageSize
	}

	return &LogQueryResult{
		Items: items,
		Total: int64(total),
		Page:  q.Page,
		Size:  q.Size,
	}, nil
}

// Writef writes a formatted log entry.
func (s *LogService) Writef(level LogLevel, source, format string, args ...interface{}) {
	s.Write(level, source, fmt.Sprintf(format, args...))
}

// Debug logs a debug message.
func (s *LogService) Debug(source, message string, opts ...LogOption) {
	s.Write(LogLevelDebug, source, message, opts...)
}

// Info logs an info message.
func (s *LogService) Info(source, message string, opts ...LogOption) {
	s.Write(LogLevelInfo, source, message, opts...)
}

// Warn logs a warning message.
func (s *LogService) Warn(source, message string, opts ...LogOption) {
	s.Write(LogLevelWarn, source, message, opts...)
}

// Error logs an error message.
func (s *LogService) Error(source, message string, opts ...LogOption) {
	s.Write(LogLevelError, source, message, opts...)
}

// GetLogCenter returns the underlying LogCenter for advanced use.
func (s *LogService) GetLogCenter() logcenter.Logger {
	return s.center
}

// ============================================================================
// Log Options
// ============================================================================

// LogOption is a function that modifies a log entry.
type LogOption func(entry *logcenter.Entry)

// WithTaskID adds a task ID to a log entry.
func WithTaskID(taskID string) LogOption {
	return func(entry *logcenter.Entry) {
		entry.TaskID = taskID
	}
}

// WithDetail adds a detail message to a log entry.
func WithDetail(detail string) LogOption {
	return func(entry *logcenter.Entry) {
		entry.Detail = detail
	}
}

// WithWorkflowID adds a workflow ID to a log entry.
func WithWorkflowID(workflowID string) LogOption {
	return func(entry *logcenter.Entry) {
		entry.WorkflowID = workflowID
	}
}

// WithNodeID adds a workflow node ID to a log entry.
func WithNodeID(nodeID string) LogOption {
	return func(entry *logcenter.Entry) {
		entry.NodeID = nodeID
	}
}

// WithRunID adds a run ID to a log entry.
func WithRunID(runID string) LogOption {
	return func(entry *logcenter.Entry) {
		entry.RunID = runID
	}
}

// WithRaw adds the original raw log to a log entry.
func WithRaw(raw string) LogOption {
	return func(entry *logcenter.Entry) {
		entry.Raw = raw
	}
}

// WithMetadata adds metadata to a log entry.
func WithMetadata(metadata map[string]interface{}) LogOption {
	return func(entry *logcenter.Entry) {
		entry.Metadata = metadata
	}
}

// ============================================================================
// Deprecated: Legacy log service for backward compatibility
// ============================================================================

// OldLogService is the legacy in-memory log service.
// Deprecated: Use LogService with LogCenter instead.
type OldLogService struct {
	// Kept for backward compatibility only
}