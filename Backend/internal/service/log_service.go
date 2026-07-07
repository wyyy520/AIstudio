package service

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// LogLevel represents a log severity level.
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

// LogEntry represents a single log entry.
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
// Stores logs in memory with capacity limits.
type LogService struct {
	mu       sync.RWMutex
	entries  []LogEntry
	nextID   int64
	maxSize  int
}

// NewLogService creates a new LogService.
func NewLogService() *LogService {
	log.Println("[log-service] initializing in-memory log store (max 10000 entries)")
	return &LogService{
		entries: make([]LogEntry, 0, 1000),
		nextID:  1,
		maxSize: 10000,
	}
}

// Write adds a new log entry.
func (s *LogService) Write(level LogLevel, source, message string, opts ...LogOption) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := LogEntry{
		ID:        s.nextID,
		Timestamp: time.Now(),
		Level:     level,
		Source:    source,
		Message:   message,
	}

	for _, opt := range opts {
		opt(&entry)
	}

	s.entries = append(s.entries, entry)
	s.nextID++

	// Trim old entries if exceeding max size
	if len(s.entries) > s.maxSize {
		trim := len(s.entries) - s.maxSize
		s.entries = s.entries[trim:]
	}
}

// LogOption is a function that modifies a log entry.
type LogOption func(*LogEntry)

// WithTaskID adds a task ID to the log entry.
func WithTaskID(taskID string) LogOption {
	return func(e *LogEntry) {
		e.TaskID = taskID
	}
}

// WithDetail adds detail info to the log entry.
func WithDetail(detail string) LogOption {
	return func(e *LogEntry) {
		e.Detail = detail
	}
}

// Info writes an INFO level log.
func (s *LogService) Info(source, message string, opts ...LogOption) {
	s.Write(LogLevelInfo, source, message, opts...)
}

// Warn writes a WARN level log.
func (s *LogService) Warn(source, message string, opts ...LogOption) {
	s.Write(LogLevelWarn, source, message, opts...)
}

// Error writes an ERROR level log.
func (s *LogService) Error(source, message string, opts ...LogOption) {
	s.Write(LogLevelError, source, message, opts...)
}

// Debug writes a DEBUG level log.
func (s *LogService) Debug(source, message string, opts ...LogOption) {
	s.Write(LogLevelDebug, source, message, opts...)
}

// Query retrieves logs with filtering and pagination.
func (s *LogService) Query(q LogQuery) (*LogQueryResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Size <= 0 || q.Size > 100 {
		q.Size = 20
	}

	// Parse time range
	var startTime, endTime time.Time
	if q.Start != "" {
		t, err := time.Parse(time.RFC3339, q.Start)
		if err != nil {
			return nil, fmt.Errorf("invalid start time: %w", err)
		}
		startTime = t
	}
	if q.End != "" {
		t, err := time.Parse(time.RFC3339, q.End)
		if err != nil {
			return nil, fmt.Errorf("invalid end time: %w", err)
		}
		endTime = t
	}

	// Filter
	var filtered []LogEntry
	for _, entry := range s.entries {
		if q.Level != "" && entry.Level != q.Level {
			continue
		}
		if q.Source != "" && entry.Source != q.Source {
			continue
		}
		if q.TaskID != "" && entry.TaskID != q.TaskID {
			continue
		}
		if q.Keyword != "" && !containsSubstring(toLower(entry.Message), toLower(q.Keyword)) {
			continue
		}
		if !startTime.IsZero() && entry.Timestamp.Before(startTime) {
			continue
		}
		if !endTime.IsZero() && entry.Timestamp.After(endTime) {
			continue
		}
		filtered = append(filtered, entry)
	}

	// Sort by time descending (newest first)
	sorted := make([]LogEntry, len(filtered))
	copy(sorted, filtered)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Timestamp.After(sorted[i].Timestamp) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	total := int64(len(sorted))

	// Paginate
	start := (q.Page - 1) * q.Size
	if start >= len(sorted) {
		return &LogQueryResult{Items: []LogEntry{}, Total: total, Page: q.Page, Size: q.Size}, nil
	}

	end := start + q.Size
	if end > len(sorted) {
		end = len(sorted)
	}

	result := &LogQueryResult{
		Items: sorted[start:end],
		Total: total,
		Page:  q.Page,
		Size:  q.Size,
	}

	return result, nil
}