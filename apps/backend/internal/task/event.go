package task

import (
	"log"
	"sync"
	"time"
)

// TaskEventType defines the type of a task lifecycle event.
type TaskEventType string

const (
	EventTaskCreated   TaskEventType = "task.created"
	EventTaskStarted   TaskEventType = "task.started"
	EventTaskProgress  TaskEventType = "task.progress"
	EventTaskCompleted TaskEventType = "task.completed"
	EventTaskFailed    TaskEventType = "task.failed"
	EventTaskCancelled TaskEventType = "task.cancelled"
	EventTaskLog       TaskEventType = "task.log"
)

// TaskEvent represents a lifecycle event emitted by the task system.
// These events can be consumed by WebSocket pushers, loggers, and external systems.
type TaskEvent struct {
	TaskID    string        `json:"taskId"`
	Type      TaskEventType `json:"type"`
	Status    Status        `json:"status"`
	Progress  float64       `json:"progress"`
	Timestamp time.Time     `json:"timestamp"`
	Data      interface{}   `json:"data,omitempty"`
}

// TaskEventListener is a callback function that receives task events.
type TaskEventListener func(event *TaskEvent)

// EventBus manages task event dispatch and listener registration.
// It is thread-safe and supports multiple listeners.
type EventBus struct {
	mu        sync.RWMutex
	listeners map[string][]TaskEventListener // keyed by event type, "" for all events
}

// NewEventBus creates a new event bus.
func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[string][]TaskEventListener),
	}
}

// Subscribe registers a listener for a specific event type.
// If eventType is "", the listener receives all events.
func (eb *EventBus) Subscribe(eventType TaskEventType, listener TaskEventListener) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	key := string(eventType)
	eb.listeners[key] = append(eb.listeners[key], listener)
}

// SubscribeAll registers a listener for all event types.
func (eb *EventBus) SubscribeAll(listener TaskEventListener) {
	eb.Subscribe("", listener)
}

// Emit dispatches an event to all registered listeners.
// It is safe to call from multiple goroutines.
func (eb *EventBus) Emit(event *TaskEvent) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	// Notify listeners registered for this specific event type
	if listeners, ok := eb.listeners[string(event.Type)]; ok {
		for _, listener := range listeners {
			eb.invokeListener(listener, event)
		}
	}

	// Notify catch-all listeners
	if listeners, ok := eb.listeners[""]; ok {
		for _, listener := range listeners {
			eb.invokeListener(listener, event)
		}
	}
}

func (eb *EventBus) invokeListener(listener TaskEventListener, event *TaskEvent) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[event-bus] listener panic recovered: %v", r)
		}
	}()
	listener(event)
}

// EmitTaskCreated emits a task.created event.
func (eb *EventBus) EmitTaskCreated(task *Task) {
	eb.Emit(&TaskEvent{
		TaskID:    task.ID,
		Type:      EventTaskCreated,
		Status:    task.Status,
		Progress:  task.Progress,
		Timestamp: time.Now(),
		Data:      task,
	})
}

// EmitTaskStarted emits a task.started event.
func (eb *EventBus) EmitTaskStarted(task *Task) {
	eb.Emit(&TaskEvent{
		TaskID:    task.ID,
		Type:      EventTaskStarted,
		Status:    task.Status,
		Progress:  task.Progress,
		Timestamp: time.Now(),
		Data:      task,
	})
}

// EmitTaskProgress emits a task.progress event.
func (eb *EventBus) EmitTaskProgress(task *Task) {
	eb.Emit(&TaskEvent{
		TaskID:    task.ID,
		Type:      EventTaskProgress,
		Status:    task.Status,
		Progress:  task.Progress,
		Timestamp: time.Now(),
		Data:      task,
	})
}

// EmitTaskCompleted emits a task.completed event.
func (eb *EventBus) EmitTaskCompleted(task *Task) {
	eb.Emit(&TaskEvent{
		TaskID:    task.ID,
		Type:      EventTaskCompleted,
		Status:    task.Status,
		Progress:  task.Progress,
		Timestamp: time.Now(),
		Data:      task.Result,
	})
}

// EmitTaskFailed emits a task.failed event.
func (eb *EventBus) EmitTaskFailed(task *Task) {
	eb.Emit(&TaskEvent{
		TaskID:    task.ID,
		Type:      EventTaskFailed,
		Status:    task.Status,
		Progress:  task.Progress,
		Timestamp: time.Now(),
		Data:      task.Error,
	})
}

// EmitTaskCancelled emits a task.cancelled event.
func (eb *EventBus) EmitTaskCancelled(task *Task) {
	eb.Emit(&TaskEvent{
		TaskID:    task.ID,
		Type:      EventTaskCancelled,
		Status:    task.Status,
		Progress:  task.Progress,
		Timestamp: time.Now(),
	})
}

// EmitTaskLog emits a task.log event with log data.
func (eb *EventBus) EmitTaskLog(taskID string, level, message, source string) {
	eb.Emit(&TaskEvent{
		TaskID: taskID,
		Type:   EventTaskLog,
		Status: StatusRunning,
		Data: map[string]interface{}{
			"level":   level,
			"message": message,
			"source":  source,
		},
		Timestamp: time.Now(),
	})
}