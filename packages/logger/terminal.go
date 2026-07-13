package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type TerminalWriter struct {
	mu     sync.Mutex
	writer io.Writer
}

type ColorCode string

const (
	ColorReset  ColorCode = "\033[0m"
	ColorRed    ColorCode = "\033[31m"
	ColorGreen  ColorCode = "\033[32m"
	ColorYellow ColorCode = "\033[33m"
	ColorBlue   ColorCode = "\033[34m"
	ColorPurple ColorCode = "\033[35m"
	ColorCyan   ColorCode = "\033[36m"
	ColorGray   ColorCode = "\033[37m"
)

var levelColors = map[LogLevel]ColorCode{
	LogLevelDebug: ColorGray,
	LogLevelInfo:  ColorGreen,
	LogLevelWarn:  ColorYellow,
	LogLevelError: ColorRed,
	LogLevelFatal: ColorPurple,
}

var levelLabels = map[LogLevel]string{
	LogLevelDebug: "DBG",
	LogLevelInfo:  "INF",
	LogLevelWarn:  "WRN",
	LogLevelError: "ERR",
	LogLevelFatal: "FTL",
}

func NewTerminalWriter() *TerminalWriter {
	return &TerminalWriter{
		writer: os.Stdout,
	}
}

func NewTerminalWriterWith(w io.Writer) *TerminalWriter {
	return &TerminalWriter{
		writer: w,
	}
}

func (tw *TerminalWriter) Write(entry LogEntry) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	color := levelColors[entry.Level]
	label := levelLabels[entry.Level]
	timestamp := entry.Timestamp.Format("15:04:05.000")

	line := fmt.Sprintf("%s%s %s [%s]%s %s\n",
		color, timestamp, label, entry.Source, ColorReset, entry.Message)

	if entry.Raw != "" {
		line = fmt.Sprintf("%s%s %s [%s]%s %s\n  raw: %s\n",
			color, timestamp, label, entry.Source, ColorReset, entry.Message, entry.Raw)
	}

	fmt.Fprint(tw.writer, line)
}

func (tw *TerminalWriter) WriteStructured(entry LogEntry) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	color := levelColors[entry.Level]
	label := levelLabels[entry.Level]
	timestamp := entry.Timestamp.Format(time.RFC3339)

	fmt.Fprintf(tw.writer, "%s%s [%s] [%s]%s %s\n",
		color, timestamp, label, entry.Source, ColorReset, entry.Message)

	if len(entry.Metadata) > 0 {
		for k, v := range entry.Metadata {
			fmt.Fprintf(tw.writer, "  %s%s%s: %v\n", ColorCyan, k, ColorReset, v)
		}
	}
}

func (tw *TerminalWriter) Clear() {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	fmt.Fprint(tw.writer, "\033[2J\033[H")
}

func SupportsColor() bool {
	term := os.Getenv("TERM")
	if term == "" {
		return false
	}
	return true
}

func FormatLevel(level LogLevel) string {
	label := levelLabels[level]
	if !SupportsColor() {
		return label
	}
	color := levelColors[level]
	return string(color) + label + string(ColorReset)
}

func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000")
}
