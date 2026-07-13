// Package logcenter provides unified log collection, storage, and streaming.
//
// The Log Center is responsible for:
//   - Collecting logs from all sources (runtime, tasks, plugins, etc.)
//   - Persisting logs for historical review
//   - Streaming logs in real-time to the UI
//   - Preserving original raw logs for advanced users
//
// Log Center does NOT analyze logs — that's Diagnostic's job.
package logcenter

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Types
// ============================================================================

// Level represents a log level.
type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
	LevelFatal Level = "FATAL"
)

// Entry represents a single log entry.
type Entry struct {
	ID         string            `json:"id"`
	Timestamp  time.Time         `json:"timestamp"`
	Level      Level             `json:"level"`
	Source     string            `json:"source"`     // Module name
	Message    string            `json:"message"`
	Detail     string            `json:"detail,omitempty"` // Detailed message
	TaskID     string            `json:"taskId,omitempty"`
	WorkflowID string            `json:"workflowId,omitempty"`
	NodeID     string            `json:"nodeId,omitempty"` // Workflow node ID
	RunID      string            `json:"runId,omitempty"`
	Raw        string            `json:"raw,omitempty"`  // Original raw log (preserved for advanced users)
	Metadata   map[string]any    `json:"metadata,omitempty"`
}

// Filter defines log query filters.
type Filter struct {
	Levels     []Level  `json:"levels,omitempty"`
	Sources    []string `json:"sources,omitempty"`
	TaskID     string   `json:"taskId,omitempty"`
	WorkflowID string   `json:"workflowId,omitempty"`
	NodeID     string   `json:"nodeId,omitempty"`
	RunID      string   `json:"runId,omitempty"`
	Search     string   `json:"search,omitempty"`  // Full-text search
	Since      *time.Time `json:"since,omitempty"`
	Until      *time.Time `json:"until,omitempty"`
	Limit      int      `json:"limit,omitempty"`
	Offset     int      `json:"offset,omitempty"`
}

// ============================================================================
// Log Center
// ============================================================================

// LogCenter is the central log management system.
type LogCenter struct {
	mu         sync.RWMutex
	entries    []Entry
	maxEntries int
	listeners  []func(Entry)
}

// New creates a new LogCenter.
func New(maxEntries int) *LogCenter {
	if maxEntries <= 0 {
		maxEntries = 10000
	}
	return &LogCenter{
		entries:    make([]Entry, 0, maxEntries),
		maxEntries: maxEntries,
		listeners:  make([]func(Entry), 0),
	}
}

// Log logs a message with the given level and source.
func (lc *LogCenter) Log(level Level, source, message string) {
	lc.logEntry(Entry{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Level:     level,
		Source:    source,
		Message:   message,
	})
}

// Logf logs a formatted message.
func (lc *LogCenter) Logf(level Level, source, format string, args ...interface{}) {
	lc.Log(level, source, fmt.Sprintf(format, args...))
}

// LogEntry logs a structured entry.
func (lc *LogCenter) LogEntry(entry Entry) {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	lc.logEntry(entry)
}

// Debug logs a debug message.
func (lc *LogCenter) Debug(source, message string) {
	lc.Log(LevelDebug, source, message)
}

// Info logs an info message.
func (lc *LogCenter) Info(source, message string) {
	lc.Log(LevelInfo, source, message)
}

// Warn logs a warning message.
func (lc *LogCenter) Warn(source, message string) {
	lc.Log(LevelWarn, source, message)
}

// Error logs an error message.
func (lc *LogCenter) Error(source, message string) {
	lc.Log(LevelError, source, message)
}

// Query retrieves log entries matching the filter.
func (lc *LogCenter) Query(filter Filter) []Entry {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	var result []Entry
	for _, entry := range lc.entries {
		if matchFilter(entry, filter) {
			result = append(result, entry)
		}
	}

	// Apply pagination
	if filter.Limit > 0 {
		start := filter.Offset
		if start >= len(result) {
			return nil
		}
		end := start + filter.Limit
		if end > len(result) {
			end = len(result)
		}
		return result[start:end]
	}

	return result
}

// GetByTaskID returns all log entries for a task.
func (lc *LogCenter) GetByTaskID(taskID string) []Entry {
	return lc.Query(Filter{TaskID: taskID})
}

// GetByWorkflowID returns all log entries for a workflow.
func (lc *LogCenter) GetByWorkflowID(workflowID string) []Entry {
	return lc.Query(Filter{WorkflowID: workflowID})
}

// GetByRunID returns all log entries for a run.
func (lc *LogCenter) GetByRunID(runID string) []Entry {
	return lc.Query(Filter{RunID: runID})
}

// GetByNodeID returns all log entries for a workflow node.
func (lc *LogCenter) GetByNodeID(nodeID string) []Entry {
	return lc.Query(Filter{NodeID: nodeID})
}

// Subscribe registers a listener for real-time log streaming.
// Returns an unsubscribe function.
func (lc *LogCenter) Subscribe(listener func(Entry)) func() {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	lc.listeners = append(lc.listeners, listener)
	idx := len(lc.listeners) - 1

	return func() {
		lc.mu.Lock()
		defer lc.mu.Unlock()
		lc.listeners = append(lc.listeners[:idx], lc.listeners[idx+1:]...)
	}
}

// Clear removes all log entries.
func (lc *LogCenter) Clear() {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.entries = make([]Entry, 0, lc.maxEntries)
}

// ClearByTaskID removes all log entries for a task.
func (lc *LogCenter) ClearByTaskID(taskID string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	var remaining []Entry
	for _, entry := range lc.entries {
		if entry.TaskID != taskID {
			remaining = append(remaining, entry)
		}
	}
	lc.entries = remaining
}

// Count returns the total number of log entries.
func (lc *LogCenter) Count() int {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return len(lc.entries)
}

// ============================================================================
// Private
// ============================================================================

func (lc *LogCenter) logEntry(entry Entry) {
	lc.mu.Lock()
	lc.entries = append(lc.entries, entry)
	if len(lc.entries) > lc.maxEntries {
		lc.entries = lc.entries[len(lc.entries)-lc.maxEntries:]
	}
	listeners := make([]func(Entry), len(lc.listeners))
	copy(listeners, lc.listeners)
	lc.mu.Unlock()

	// Notify listeners
	for _, listener := range listeners {
		listener(entry)
	}
}

func matchFilter(entry Entry, filter Filter) bool {
	if len(filter.Levels) > 0 {
		matched := false
		for _, l := range filter.Levels {
			if entry.Level == l {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if len(filter.Sources) > 0 {
		matched := false
		for _, s := range filter.Sources {
			if entry.Source == s {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	if filter.TaskID != "" && entry.TaskID != filter.TaskID {
		return false
	}
	if filter.WorkflowID != "" && entry.WorkflowID != filter.WorkflowID {
		return false
	}
	if filter.NodeID != "" && entry.NodeID != filter.NodeID {
		return false
	}
	if filter.RunID != "" && entry.RunID != filter.RunID {
		return false
	}

	if filter.Search != "" && !strings.Contains(entry.Message, filter.Search) && !strings.Contains(entry.Raw, filter.Search) {
		return false
	}

	if filter.Since != nil && entry.Timestamp.Before(*filter.Since) {
		return false
	}
	if filter.Until != nil && entry.Timestamp.After(*filter.Until) {
		return false
	}

	return true
}