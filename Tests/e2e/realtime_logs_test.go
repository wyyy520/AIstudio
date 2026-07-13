package e2e

import (
	"testing"
	"time"

	"github.com/aistudio/backend/internal/logcenter"
	"github.com/aistudio/backend/internal/task"
)

func TestLogCenter(t *testing.T) {
	lc := logcenter.New(100)

	lc.Info("test-source", "test message 1")
	lc.Warn("test-source", "test warning")
	lc.Error("test-source", "test error")

	entries := lc.Query(logcenter.Filter{})
	if len(entries) != 3 {
		t.Fatalf("expected 3 log entries, got %d", len(entries))
	}

	infoEntries := lc.Query(logcenter.Filter{Levels: []logcenter.Level{logcenter.LevelInfo}})
	if len(infoEntries) != 1 {
		t.Errorf("expected 1 info entry, got %d", len(infoEntries))
	}
}

func TestLogCenterWithTaskID(t *testing.T) {
	lc := logcenter.New(100)

	entry := logcenter.Entry{
		TaskID:  "task-123",
		Level:   logcenter.LevelInfo,
		Source:  "test",
		Message: "task log entry",
	}
	lc.LogEntry(entry)

	filtered := lc.GetByTaskID("task-123")
	if len(filtered) != 1 {
		t.Fatalf("expected 1 entry for task-123, got %d", len(filtered))
	}
	if filtered[0].Message != "task log entry" {
		t.Errorf("expected message 'task log entry', got '%s'", filtered[0].Message)
	}

	notFound := lc.GetByTaskID("nonexistent")
	if len(notFound) != 0 {
		t.Errorf("expected 0 entries for nonexistent task, got %d", len(notFound))
	}
}

func TestLogCenterSubscribe(t *testing.T) {
	lc := logcenter.New(100)

	received := make(chan logcenter.Entry, 1)
	unsub := lc.Subscribe(func(entry logcenter.Entry) {
		received <- entry
	})
	defer unsub()

	lc.Info("test", "realtime message")

	select {
	case entry := <-received:
		if entry.Message != "realtime message" {
			t.Errorf("expected 'realtime message', got '%s'", entry.Message)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for log subscription callback")
	}
}

func TestTaskEventBus(t *testing.T) {
	bus := task.NewEventBus()

	received := make(chan *task.TaskEvent, 1)
	bus.Subscribe(task.EventTaskProgress, func(event *task.TaskEvent) {
		received <- event
	})

	taskObj := &task.Task{
		ID:       "task-event-test",
		Status:   task.StatusRunning,
		Progress: 0.5,
	}
	bus.EmitTaskProgress(taskObj)

	select {
	case event := <-received:
		if event.TaskID != "task-event-test" {
			t.Errorf("expected task ID 'task-event-test', got '%s'", event.TaskID)
		}
		if event.Type != task.EventTaskProgress {
			t.Errorf("expected event type %s, got %s", task.EventTaskProgress, event.Type)
		}
		if event.Progress != 0.5 {
			t.Errorf("expected progress 0.5, got %f", event.Progress)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for task event")
	}
}
