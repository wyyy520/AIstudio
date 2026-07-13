package service

import (
	"testing"

	"github.com/aistudio/backend/internal/logcenter"
)

func newTestLogService() *LogService {
	lc := logcenter.New(1000)
	return NewLogService(lc)
}

func TestNewLogService(t *testing.T) {
	t.Run("creates service with log center", func(t *testing.T) {
		svc := newTestLogService()
		if svc == nil {
			t.Fatal("expected non-nil service")
		}
		if svc.center == nil {
			t.Error("expected non-nil log center")
		}
	})
}

func TestLogOptions(t *testing.T) {
	t.Run("WithTaskID", func(t *testing.T) {
		entry := logcenter.Entry{}
		WithTaskID("task-123")(&entry)
		if entry.TaskID != "task-123" {
			t.Errorf("expected task-123, got %s", entry.TaskID)
		}
	})

	t.Run("WithDetail", func(t *testing.T) {
		entry := logcenter.Entry{}
		WithDetail("detailed info")(&entry)
		if entry.Detail != "detailed info" {
			t.Errorf("expected 'detailed info', got '%s'", entry.Detail)
		}
	})

	t.Run("multiple options", func(t *testing.T) {
		entry := logcenter.Entry{}
		WithTaskID("task-456")(&entry)
		WithDetail("extra")(&entry)
		if entry.TaskID != "task-456" || entry.Detail != "extra" {
			t.Error("multiple options failed")
		}
	})
}

func TestWriteAndQuery(t *testing.T) {
	svc := newTestLogService()

	t.Run("write and query basic", func(t *testing.T) {
		svc.Info("test-source", "test info message")
		svc.Warn("test-source", "test warn message")
		svc.Error("test-source", "test error message")
		svc.Debug("test-source", "test debug message")

		result, err := svc.Query(LogQuery{Page: 1, Size: 100})
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if result.Total < 4 {
			t.Errorf("expected at least 4 entries, got %d", result.Total)
		}
	})

	t.Run("query with level filter", func(t *testing.T) {
		result, err := svc.Query(LogQuery{
			Level: LogLevelError,
			Page:  1,
			Size:  10,
		})
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		for _, entry := range result.Items {
			if entry.Level != LogLevelError {
				t.Errorf("expected ERROR level, got %s", entry.Level)
			}
		}
	})

	t.Run("query with source filter", func(t *testing.T) {
		result, err := svc.Query(LogQuery{
			Source: "test-source",
			Page:   1,
			Size:   10,
		})
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		for _, entry := range result.Items {
			if entry.Source != "test-source" {
				t.Errorf("expected 'test-source', got '%s'", entry.Source)
			}
		}
	})

	t.Run("query with task filter", func(t *testing.T) {
		svc.Info("task-test", "task message", WithTaskID("task-999"))
		result, err := svc.Query(LogQuery{
			TaskID: "task-999",
			Page:   1,
			Size:   10,
		})
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if result.Total < 1 {
			t.Errorf("expected at least 1 entry, got %d", result.Total)
		}
		for _, entry := range result.Items {
			if entry.TaskID != "task-999" {
				t.Errorf("expected task-999, got %s", entry.TaskID)
			}
		}
	})

	t.Run("pagination", func(t *testing.T) {
		// Clear and write test entries
		svc.center.Clear()
		for i := 0; i < 25; i++ {
			svc.Info("pagination-test", "message")
		}

		// Page 1
		page1, err := svc.Query(LogQuery{Page: 1, Size: 10})
		if err != nil {
			t.Fatalf("page 1 query failed: %v", err)
		}
		if len(page1.Items) != 10 {
			t.Errorf("expected 10 items on page 1, got %d", len(page1.Items))
		}

		// Page 3
		page3, err := svc.Query(LogQuery{Page: 3, Size: 10})
		if err != nil {
			t.Fatalf("page 3 query failed: %v", err)
		}
		if len(page3.Items) != 5 {
			t.Errorf("expected 5 items on page 3, got %d", len(page3.Items))
		}
	})

	t.Run("writef", func(t *testing.T) {
		svc.center.Clear()
		svc.Writef(LogLevelInfo, "writef-test", "formatted %s %d", "test", 42)
		result, err := svc.Query(LogQuery{Page: 1, Size: 10})
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		found := false
		for _, entry := range result.Items {
			if entry.Message == "formatted test 42" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected formatted message not found")
		}
	})
}

func TestMaxEntries(t *testing.T) {
	lc := logcenter.New(50)
	svc := NewLogService(lc)

	for i := 0; i < 100; i++ {
		svc.Info("overflow-test", "message")
	}

	result, err := svc.Query(LogQuery{Page: 1, Size: 200})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if result.Total > 50 {
		t.Errorf("expected at most 50 entries, got %d", result.Total)
	}
}

func TestGetLogCenter(t *testing.T) {
	svc := newTestLogService()
	lc := svc.GetLogCenter()
	if lc == nil {
		t.Error("expected non-nil log center from GetLogCenter")
	}
}

func TestConcurrentWrites(t *testing.T) {
	svc := newTestLogService()
	done := make(chan bool)

	write := func(n int) {
		for i := 0; i < 100; i++ {
			svc.Info("concurrent", "message")
		}
		done <- true
	}

	go write(1)
	go write(2)
	go write(3)

	for i := 0; i < 3; i++ {
		<-done
	}

	result, err := svc.Query(LogQuery{Page: 1, Size: 500})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if result.Total < 300 {
		t.Errorf("expected at least 300 entries, got %d", result.Total)
	}
}

func TestLogLevels(t *testing.T) {
	svc := newTestLogService()

	svc.Debug("lvl", "debug msg")
	svc.Info("lvl", "info msg")
	svc.Warn("lvl", "warn msg")
	svc.Error("lvl", "error msg")

	for _, level := range []LogLevel{LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError} {
		result, err := svc.Query(LogQuery{Level: level, Page: 1, Size: 10})
		if err != nil {
			t.Fatalf("query failed for level %s: %v", level, err)
		}
		if result.Total < 1 {
			t.Errorf("expected at least 1 entry for level %s, got %d", level, result.Total)
		}
	}

	// Verify all entries exist
	result, err := svc.Query(LogQuery{Page: 1, Size: 10})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if result.Total < 4 {
		t.Errorf("expected at least 4 entries, got %d", result.Total)
	}
}