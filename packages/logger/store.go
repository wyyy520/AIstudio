package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type FileStore struct {
	mu          sync.Mutex
	dir         string
	baseName    string
	maxSize     int64
	maxFiles    int
	currentSize int64
	currentFile *os.File
	entries     []LogEntry
}

func NewFileStore(path string, maxSize int64, maxFiles int) *FileStore {
	absPath, _ := filepath.Abs(path)
	dir := filepath.Dir(absPath)
	baseName := filepath.Base(absPath)
	os.MkdirAll(dir, 0755)

	fs := &FileStore{
		dir:      dir,
		baseName: baseName,
		maxSize:  maxSize,
		maxFiles: maxFiles,
		entries:  make([]LogEntry, 0),
	}

	fs.rotateIfNeeded()
	return fs
}

func (fs *FileStore) Write(entry LogEntry) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.entries = append(fs.entries, entry)

	if err := fs.flushLocked(); err != nil {
		return err
	}

	return nil
}

func (fs *FileStore) WriteBatch(entries []LogEntry) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.entries = append(fs.entries, entries...)
	return fs.flushLocked()
}

func (fs *FileStore) Query(filter Filter) ([]LogEntry, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	var result []LogEntry
	for _, entry := range fs.entries {
		if matchFilter(entry, filter) {
			result = append(result, entry)
		}
	}

	if filter.Limit > 0 {
		start := filter.Offset
		if start >= len(result) {
			return nil, nil
		}
		end := start + filter.Limit
		if end > len(result) {
			end = len(result)
		}
		return result[start:end], nil
	}

	return result, nil
}

func (fs *FileStore) ReadAll() ([]LogEntry, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	entries := make([]LogEntry, len(fs.entries))
	copy(entries, fs.entries)
	return entries, nil
}

func (fs *FileStore) Clear() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.entries = fs.entries[:0]
	if fs.currentFile != nil {
		fs.currentFile.Close()
		fs.currentFile = nil
	}
	return os.RemoveAll(fs.dir)
}

func (fs *FileStore) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.currentFile != nil {
		fs.currentFile.Close()
		fs.currentFile = nil
	}
	return nil
}

func (fs *FileStore) currentPath() string {
	return filepath.Join(fs.dir, fs.baseName)
}

func (fs *FileStore) rotatedPath(index int) string {
	ext := filepath.Ext(fs.baseName)
	name := strings.TrimSuffix(fs.baseName, ext)
	if index == 0 {
		return filepath.Join(fs.dir, fs.baseName)
	}
	return filepath.Join(fs.dir, fmt.Sprintf("%s.%d%s", name, index, ext))
}

func (fs *FileStore) rotateIfNeeded() {
	path := fs.currentPath()
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if info.Size() >= fs.maxSize {
		fs.rotate()
	}
}

func (fs *FileStore) rotate() {
	for i := fs.maxFiles - 1; i >= 0; i-- {
		oldPath := fs.rotatedPath(i)
		newPath := fs.rotatedPath(i + 1)
		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath)
		}
	}
	if fs.currentFile != nil {
		fs.currentFile.Close()
		fs.currentFile = nil
	}
}

func (fs *FileStore) flushLocked() error {
	if len(fs.entries) == 0 {
		return nil
	}

	fs.rotateIfNeeded()

	path := fs.currentPath()
	if fs.currentFile == nil {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		fs.currentFile = f
		info, _ := os.Stat(path)
		if info != nil {
			fs.currentSize = info.Size()
		}
	}

	encoder := json.NewEncoder(fs.currentFile)
	for _, entry := range fs.entries {
		if err := encoder.Encode(entry); err != nil {
			return fmt.Errorf("failed to encode log entry: %w", err)
		}
	}
	fs.currentFile.Sync()

	info, _ := os.Stat(path)
	if info != nil {
		fs.currentSize = info.Size()
	}

	fs.entries = fs.entries[:0]
	return nil
}

func matchFilter(entry LogEntry, filter Filter) bool {
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

	if filter.RunID != "" && entry.RunID != filter.RunID {
		return false
	}

	if filter.Search != "" {
		if !strings.Contains(entry.Message, filter.Search) && !strings.Contains(entry.Raw, filter.Search) {
			return false
		}
	}

	if filter.Since != nil && entry.Timestamp.Before(*filter.Since) {
		return false
	}

	if filter.Until != nil && entry.Timestamp.After(*filter.Until) {
		return false
	}

	return true
}

func LoadLogFile(path string) ([]LogEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []LogEntry
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.Before(entries[j].Timestamp)
	})

	return entries, nil
}
