package logcenter

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// PersistentLogCenter wraps LogCenter with file-based persistence.
// Per silu.md 5.11, logs are saved to <project>/logs/ directory.
type PersistentLogCenter struct {
	*LogCenter
	mu         sync.Mutex
	logDir     string
	maxLogSize int64 // rotate when file exceeds this size (default 10MB)
}

// NewPersistent creates a PersistentLogCenter with file persistence.
func NewPersistent(maxEntries int, logDir string) *PersistentLogCenter {
	if logDir == "" {
		logDir = "logs"
	}
	os.MkdirAll(logDir, 0755)

	lc := &PersistentLogCenter{
		LogCenter:  New(maxEntries),
		logDir:     logDir,
		maxLogSize: 10 * 1024 * 1024, // 10MB
	}
	return lc
}

// LogEntry logs a structured entry and persists it to file.
func (lc *PersistentLogCenter) LogEntry(entry Entry) {
	lc.LogCenter.LogEntry(entry)
	lc.persistEntry(entry)
}

// Log logs a message with the given level and source, then persists.
func (lc *PersistentLogCenter) Log(level Level, source, message string) {
	entry := Entry{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Level:     level,
		Source:    source,
		Message:   message,
	}
	lc.LogCenter.LogEntry(entry)
	lc.persistEntry(entry)
}

// persistEntry writes a single log entry to the appropriate category file.
func (lc *PersistentLogCenter) persistEntry(entry Entry) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// Determine category based on source
	category := lc.sourceToCategory(entry.Source)
	logFile := filepath.Join(lc.logDir, category+".log")

	// Format: 2024-01-15T10:30:00Z [INFO] module: message
	line := fmt.Sprintf("%s [%s] %s: %s\n",
		entry.Timestamp.Format(time.RFC3339),
		entry.Level,
		entry.Source,
		entry.Message,
	)

	// Append to file (with rotation if needed)
	lc.appendWithRotation(logFile, line)
}

// sourceToCategory maps a module source to a log category.
// Per silu.md 5.10: system, runtime, compiler, generator, plugin, environment, ai
func (lc *PersistentLogCenter) sourceToCategory(source string) string {
	s := strings.ToLower(source)
	switch {
	case strings.Contains(s, "runtime") || strings.Contains(s, "run"):
		return "runtime"
	case strings.Contains(s, "compil"):
		return "compiler"
	case strings.Contains(s, "generat"):
		return "generator"
	case strings.Contains(s, "plugin"):
		return "plugin"
	case strings.Contains(s, "environ"):
		return "environment"
	case strings.Contains(s, "ai") || strings.Contains(s, "skill"):
		return "ai"
	case strings.Contains(s, "system") || strings.Contains(s, "server"):
		return "system"
	default:
		return "system"
	}
}

// appendWithRotation appends to a file, rotating if it exceeds maxLogSize.
func (lc *PersistentLogCenter) appendWithRotation(filePath, line string) {
	// Check if rotation is needed
	if info, err := os.Stat(filePath); err == nil && info.Size() > lc.maxLogSize {
		lc.rotateFile(filePath)
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(line)
}

// rotateFile renames the current log file with a timestamp suffix.
func (lc *PersistentLogCenter) rotateFile(filePath string) {
	timestamp := time.Now().Format("20060102-150405")
	rotatedPath := fmt.Sprintf("%s.%s", filePath, timestamp)
	os.Rename(filePath, rotatedPath)

	// Clean old rotated files (keep last 5)
	lc.cleanOldRotated(filePath, 5)
}

// cleanOldRotated removes old rotated log files, keeping only the most recent N.
func (lc *PersistentLogCenter) cleanOldRotated(basePath string, keep int) {
	dir := filepath.Dir(basePath)
	base := filepath.Base(basePath)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var rotated []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), base+".") {
			rotated = append(rotated, filepath.Join(dir, e.Name()))
		}
	}

	if len(rotated) <= keep {
		return
	}

	// Sort by name (which includes timestamp) and remove oldest
	sort.Strings(rotated)
	for i := 0; i < len(rotated)-keep; i++ {
		os.Remove(rotated[i])
	}
}
