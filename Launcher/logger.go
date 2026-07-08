package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger

	logFile *os.File
)

func Init(logDir string) error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, fmt.Sprintf("launcher_%s.log", timestamp))

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	logFile = f

	multiWriter := os.Stdout

	infoLogger = log.New(multiWriter, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	warnLogger = log.New(multiWriter, "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	debugLogger = log.New(multiWriter, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)

	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	return nil
}

func Info(msg string, keysAndValues ...interface{}) {
	if infoLogger != nil {
		infoLogger.Output(2, formatMessage(msg, keysAndValues...))
	} else {
		log.Output(2, formatMessage(msg, keysAndValues...))
	}
}

func Warn(msg string, keysAndValues ...interface{}) {
	if warnLogger != nil {
		warnLogger.Output(2, formatMessage(msg, keysAndValues...))
	} else {
		log.Output(2, formatMessage(msg, keysAndValues...))
	}
}

func Error(msg string, keysAndValues ...interface{}) {
	if errorLogger != nil {
		errorLogger.Output(2, formatMessage(msg, keysAndValues...))
	} else {
		log.Output(2, formatMessage(msg, keysAndValues...))
	}
}

func Debug(msg string, keysAndValues ...interface{}) {
	if debugLogger != nil {
		debugLogger.Output(2, formatMessage(msg, keysAndValues...))
	}
}

func formatMessage(msg string, keysAndValues ...string) string {
	if len(keysAndValues) == 0 {
		return msg
	}

	result := msg
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			result += fmt.Sprintf(" %s=%s", keysAndValues[i], keysAndValues[i+1])
		}
	}
	return result
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}