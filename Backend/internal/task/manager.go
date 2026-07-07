package task

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Manager is the central task management component.
// It coordinates the queue, worker pool, scheduler, and task storage.
type Manager struct {
	queue    *TaskQueue
	pool     *WorkerPool
	sched    *Scheduler
	tasks    map[string]*Task // in-memory task store
	handlers map[string]TaskHandler
	mu       sync.RWMutex
}

// NewManager creates a new task manager with the given number of workers.
func NewManager(numWorkers int) *Manager {
	queue := NewTaskQueue()
	pool := NewWorkerPool(numWorkers, queue)

	m := &Manager{
		queue:    queue,
		pool:     pool,
		tasks:    make(map[string]*Task),
		handlers: make(map[string]TaskHandler),
	}

	m.sched = NewScheduler(m)
	return m
}

// RegisterHandler registers a task handler for the given task type name.
func (m *Manager) RegisterHandler(name string, handler TaskHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.handlers[name] = handler
	m.pool.RegisterHandler(name, handler)
	log.Printf("[task-manager] registered handler: %s", name)
}

// Submit adds a new task to the queue and returns its ID.
func (m *Manager) Submit(ctx context.Context, name, description, handler string, priority Priority, payload interface{}) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate handler exists
	if _, ok := m.handlers[handler]; !ok {
		return "", fmt.Errorf("no handler registered for: %s", handler)
	}

	task := &Task{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Priority:    priority,
		Status:      StatusPending,
		Handler:     handler,
		Payload:     payload,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	m.tasks[task.ID] = task
	m.queue.Enqueue(task)
	log.Printf("[task-manager] submitted task: %s (handler: %s, priority: %s)", task.ID, handler, priority)
	return task.ID, nil
}

// GetTask returns a task by ID.
func (m *Manager) GetTask(ctx context.Context, taskID string) (*Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	return task, nil
}

// Cancel cancels a task if it is still pending or running.
func (m *Manager) Cancel(ctx context.Context, taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if err := ValidateTransition(task.Status, StatusCancelled); err != nil {
		return err
	}

	task.Status = StatusCancelled
	task.UpdatedAt = time.Now()
	completedAt := time.Now()
	task.CompletedAt = &completedAt

	// Remove from queue if still pending
	if task.Status == StatusPending {
		m.queue.Remove(taskID)
	}

	log.Printf("[task-manager] cancelled task: %s", taskID)
	return nil
}

// ListTasks returns all tasks, optionally filtered by status.
func (m *Manager) ListTasks(ctx context.Context, status ...Status) ([]*Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Task, 0)
	for _, task := range m.tasks {
		if len(status) == 0 {
			result = append(result, task)
		} else {
			for _, s := range status {
				if task.Status == s {
					result = append(result, task)
					break
				}
			}
		}
	}
	return result, nil
}

// Start initializes the worker pool and scheduler.
func (m *Manager) Start() {
	m.pool.Start()
	m.sched.Start()
	log.Println("[task-manager] started")
}

// Stop gracefully shuts down the task manager.
func (m *Manager) Stop() {
	log.Println("[task-manager] stopping...")
	m.sched.Stop()
	m.pool.Stop()
	log.Println("[task-manager] stopped")
}

// Queue returns the underlying task queue.
func (m *Manager) Queue() *TaskQueue {
	return m.queue
}

// UpdateStatus updates the status of a task if the transition is valid.
func (m *Manager) UpdateStatus(ctx context.Context, taskID string, newStatus Status) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if err := ValidateTransition(task.Status, newStatus); err != nil {
		return err
	}

	task.Status = newStatus
	task.UpdatedAt = time.Now()
	if newStatus == StatusSuccess || newStatus == StatusFailed || newStatus == StatusCancelled {
		completedAt := time.Now()
		task.CompletedAt = &completedAt
	}

	log.Printf("[task-manager] updated task %s status: %s -> %s", taskID, task.Status, newStatus)
	return nil
}

// DeleteTask removes a task from the store.
func (m *Manager) DeleteTask(ctx context.Context, taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.tasks[taskID]; !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Remove from queue if still pending
	m.queue.Remove(taskID)

	delete(m.tasks, taskID)
	log.Printf("[task-manager] deleted task: %s", taskID)
	return nil
}

// TaskCount returns the number of tasks in the given status.
func (m *Manager) TaskCount(status Status) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, t := range m.tasks {
		if t.Status == status {
			count++
		}
	}
	return count
}