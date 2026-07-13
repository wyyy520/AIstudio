package logger

import (
	"fmt"
	"sync"
	"time"

	"github.com/aistudio/packages/event"
	"github.com/google/uuid"
)

type Logger interface {
	Info(source string, message string)
	Warn(source string, message string)
	Error(source string, message string, err error)
	Stream(source string, message string)
	QueryLogs(filter Filter) ([]LogEntry, error)
	Flush() error
}

type Config struct {
	Level        LogLevel
	FileEnabled  bool
	FilePath     string
	MaxFileSize  int64
	MaxFiles     int
	TermEnabled  bool
	EventEnabled bool
	EventBus     *event.EventBus
	BufferSize   int
}

func DefaultConfig() Config {
	return Config{
		Level:        LogLevelDebug,
		FileEnabled:  true,
		FilePath:     "logs/studio.log",
		MaxFileSize:  10 * 1024 * 1024,
		MaxFiles:     5,
		TermEnabled:  true,
		EventEnabled: false,
		BufferSize:   100,
	}
}

type loggerImpl struct {
	mu       sync.Mutex
	cfg      Config
	store    *FileStore
	terminal *TerminalWriter
	buffer   []LogEntry
	closed   bool
}

func New(cfg Config) Logger {
	l := &loggerImpl{
		cfg:    cfg,
		buffer: make([]LogEntry, 0, cfg.BufferSize),
	}

	if cfg.FileEnabled && cfg.FilePath != "" {
		l.store = NewFileStore(cfg.FilePath, cfg.MaxFileSize, cfg.MaxFiles)
	}

	if cfg.TermEnabled {
		l.terminal = NewTerminalWriter()
	}

	return l
}

func (l *loggerImpl) Info(source string, message string) {
	if !l.shouldLog(LogLevelInfo) {
		return
	}
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LogLevelInfo,
		Source:    source,
		Message:   message,
	}
	l.write(entry)
}

func (l *loggerImpl) Warn(source string, message string) {
	if !l.shouldLog(LogLevelWarn) {
		return
	}
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LogLevelWarn,
		Source:    source,
		Message:   message,
	}
	l.write(entry)
}

func (l *loggerImpl) Error(source string, message string, err error) {
	if !l.shouldLog(LogLevelError) {
		return
	}
	metadata := map[string]any{}
	raw := message
	if err != nil {
		metadata["error"] = err.Error()
		raw = fmt.Sprintf("%s: %v", message, err)
	}
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LogLevelError,
		Source:    source,
		Message:   message,
		Raw:       raw,
		Metadata:  metadata,
	}
	l.write(entry)
}

func (l *loggerImpl) Stream(source string, message string) {
	if !l.shouldLog(LogLevelInfo) {
		return
	}
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LogLevelInfo,
		Source:    source,
		Message:   message,
	}
	l.write(entry)

	if l.cfg.EventEnabled && l.cfg.EventBus != nil {
		l.cfg.EventBus.Publish(event.TopicLogEntry, event.LogEventData{
			Level:   string(LogLevelInfo),
			Message: message,
			Source:  source,
		})
	}
}

func (l *loggerImpl) QueryLogs(filter Filter) ([]LogEntry, error) {
	if l.store != nil {
		return l.store.Query(filter)
	}
	return nil, nil
}

func (l *loggerImpl) Flush() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.buffer) == 0 {
		return nil
	}

	if l.store != nil {
		for _, entry := range l.buffer {
			if err := l.store.Write(entry); err != nil {
				return err
			}
		}
	}

	l.buffer = l.buffer[:0]
	return nil
}

func (l *loggerImpl) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
		LogLevelFatal: 4,
	}
	return levels[level] >= levels[l.cfg.Level]
}

func (l *loggerImpl) write(entry LogEntry) {
	entry.Timestamp = time.Now()
	if entry.Metadata == nil {
		entry.Metadata = make(map[string]any)
	}

	if l.terminal != nil {
		l.terminal.Write(entry)
	}

	l.mu.Lock()
	l.buffer = append(l.buffer, entry)
	shouldFlush := len(l.buffer) >= l.cfg.BufferSize
	l.mu.Unlock()

	if shouldFlush {
		l.Flush()
	}
}

type BufferedLogger struct {
	logger Logger
	bus    *event.EventBus
	done   chan struct{}
}

func NewBufferedLogger(logger Logger, bus *event.EventBus) *BufferedLogger {
	bl := &BufferedLogger{
		logger: logger,
		bus:    bus,
		done:   make(chan struct{}),
	}
	return bl
}

func (bl *BufferedLogger) Log(entry LogEntry) {
	switch entry.Level {
	case LogLevelWarn:
		bl.logger.Warn(entry.Source, entry.Message)
	case LogLevelError, LogLevelFatal:
		var err error
		if e, ok := entry.Metadata["error"]; ok {
			if s, ok := e.(string); ok {
				err = fmt.Errorf("%s", s)
			}
		}
		bl.logger.Error(entry.Source, entry.Message, err)
	default:
		bl.logger.Info(entry.Source, entry.Message)
	}
}

func (bl *BufferedLogger) FlushAndClose() error {
	close(bl.done)
	return bl.logger.Flush()
}

func GenerateEntryID() string {
	return uuid.New().String()
}