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
// It coordinates the queue, worker pool, scheduler, event bus, and task storage.
type Manager struct {
	queue    *TaskQueue
	pool     *WorkerPool
	sched    *Scheduler
	events   *EventBus
	repo     *TaskRepository
	tasks    map[string]*Task
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
		events:   NewEventBus(),
		tasks:    make(map[string]*Task),
		handlers: make(map[string]TaskHandler),
	}

	m.sched = NewScheduler(m)
	m.pool.SetManager(m)
	return m
}

// SetRepository sets the database repository for task persistence.
func (m *Manager) SetRepository(repo *TaskRepository) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.repo = repo
}

// EventBus returns the event bus for subscribing to task lifecycle events.
func (m *Manager) EventBus() *EventBus {
	return m.events
}

// RegisterHandler registers a task handler for the given task type name.
func (m *Manager) RegisterHandler(name string, handler TaskHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.handlers[name] = handler
	m.pool.RegisterHandler(name, handler)
	log.Printf("[task-manager] registered handler: %s", name)
}

// CreateTask creates a new task and places it in the waiting state.
// It does NOT enqueue the task; StartTask must be called to begin execution.
func (m *Manager) CreateTask(ctx context.Context, projectID, workflowID string, taskType TaskType, name, handler string, priority Priority, payload interface{}) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.handlers[handler]; !ok {
		return "", fmt.Errorf("no handler registered for: %s", handler)
	}

	task := &Task{
		ID:          uuid.New().String(),
		ProjectID:   projectID,
		WorkflowID:  workflowID,
		Type:        taskType,
		Name:        name,
		Status:      StatusWaiting,
		Progress:    0,
		Priority:    priority,
		Handler:     handler,
		Payload:     payload,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	m.tasks[task.ID] = task

	// Persist to database
	if m.repo != nil {
		if err := m.repo.Save(task); err != nil {
			log.Printf("[task-manager] failed to persist task %s: %v", task.ID, err)
		}
	}

	// Emit event
	m.events.EmitTaskCreated(task)

	log.Printf("[task-manager] created task: %s (type: %s, handler: %s)", task.ID, taskType, handler)
	return task.ID, nil
}

// StartTask moves a task from waiting to running and enqueues it for execution.
func (m *Manager) StartTask(ctx context.Context, taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if err := ValidateTransition(task.Status, StatusRunning); err != nil {
		return fmt.Errorf("cannot start task %s: %w", taskID, err)
	}

	task.Status = StatusRunning
	now := time.Now()
	task.StartTime = &now
	task.UpdatedAt = now

	// Persist status change
	if m.repo != nil {
		if err := m.repo.Update(task); err != nil {
			log.Printf("[task-manager] failed to update task %s in db: %v", taskID, err)
		}
	}

	// Enqueue for execution
	m.queue.Enqueue(task)

	// Emit event
	m.events.EmitTaskStarted(task)

	log.Printf("[task-manager] started task: %s", taskID)
	return nil
}

// UpdateProgress updates the progress of a running task (0.0 ~ 1.0).
func (m *Manager) UpdateProgress(ctx context.Context, taskID string, progress float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if task.Status != StatusRunning {
		return fmt.Errorf("task %s is not running (status: %s)", taskID, task.Status)
	}

	if progress < 0 {
		progress = 0
	}
	if progress > 1.0 {
		progress = 1.0
	}

	task.Progress = progress
	task.UpdatedAt = time.Now()

	// Persist progress update
	if m.repo != nil {
		if err := m.repo.Update(task); err != nil {
			log.Printf("[task-manager] failed to update progress for task %s: %v", taskID, err)
		}
	}

	// Emit event
	m.events.EmitTaskProgress(task)

	log.Printf("[task-manager] task %s progress: %.0f%%", taskID, progress*100)
	return nil
}

// FinishTask marks a task as successfully completed.
func (m *Manager) FinishTask(ctx context.Context, taskID string, result interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if err := ValidateTransition(task.Status, StatusSuccess); err != nil {
		return fmt.Errorf("cannot finish task %s: %w", taskID, err)
	}

	task.Status = StatusSuccess
	task.Progress = 1.0
	task.Result = result
	now := time.Now()
	task.EndTime = &now
	task.UpdatedAt = now

	// Persist
	if m.repo != nil {
		if err := m.repo.Update(task); err != nil {
			log.Printf("[task-manager] failed to persist completed task %s: %v", taskID, err)
		}
	}

	// Emit event
	m.events.EmitTaskCompleted(task)

	log.Printf("[task-manager] task %s completed successfully", taskID)
	return nil
}

// FailTask marks a task as failed with an error message.
func (m *Manager) FailTask(ctx context.Context, taskID string, errMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if err := ValidateTransition(task.Status, StatusFailed); err != nil {
		return fmt.Errorf("cannot fail task %s: %w", taskID, err)
	}

	task.Status = StatusFailed
	task.Error = errMsg
	now := time.Now()
	task.EndTime = &now
	task.UpdatedAt = now

	// Persist
	if m.repo != nil {
		if err := m.repo.Update(task); err != nil {
			log.Printf("[task-manager] failed to persist failed task %s: %v", taskID, err)
		}
	}

	// Emit event
	m.events.EmitTaskFailed(task)

	log.Printf("[task-manager] task %s failed: %s", taskID, errMsg)
	return nil
}

// Submit adds a new task, places it in waiting, and starts it immediately.
// This is a convenience method combining CreateTask + StartTask.
func (m *Manager) Submit(ctx context.Context, name, description, handler string, priority Priority, payload interface{}) (string, error) {
	taskID, err := m.CreateTask(ctx, "", "", TaskTypeWorkflow, name, handler, priority, payload)
	if err != nil {
		return "", err
	}

	// Set description after creation
	m.mu.Lock()
	if task, ok := m.tasks[taskID]; ok {
		task.Description = description
	}
	m.mu.Unlock()

	if err := m.StartTask(ctx, taskID); err != nil {
		return taskID, err
	}

	return taskID, nil
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

// GetTaskStatus returns the status response for a task.
func (m *Manager) GetTaskStatus(ctx context.Context, taskID string) (*TaskStatusResponse, error) {
	task, err := m.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return &TaskStatusResponse{
		TaskID:   task.ID,
		Status:   task.Status,
		Progress: task.Progress,
		Error:    task.Error,
	}, nil
}

// Cancel cancels a task if it is still waiting or running.
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
	now := time.Now()
	task.EndTime = &now
	task.UpdatedAt = now

	// Remove from queue if still waiting
	m.queue.Remove(taskID)

	// Persist
	if m.repo != nil {
		if err := m.repo.Update(task); err != nil {
			log.Printf("[task-manager] failed to persist cancelled task %s: %v", taskID, err)
		}
	}

	// Emit event
	m.events.EmitTaskCancelled(task)

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

	if newStatus == StatusRunning && task.StartTime == nil {
		now := time.Now()
		task.StartTime = &now
	}

	if newStatus == StatusSuccess || newStatus == StatusFailed || newStatus == StatusCancelled {
		now := time.Now()
		task.EndTime = &now
	}

	// Persist
	if m.repo != nil {
		if err := m.repo.Update(task); err != nil {
			log.Printf("[task-manager] failed to persist status update for task %s: %v", taskID, err)
		}
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

	// Remove from queue if still waiting
	m.queue.Remove(taskID)

	delete(m.tasks, taskID)

	// Remove from database
	if m.repo != nil {
		if err := m.repo.Delete(taskID); err != nil {
			log.Printf("[task-manager] failed to delete task %s from db: %v", taskID, err)
		}
	}

	log.Printf("[task-manager] deleted task: %s", taskID)
	return nil
}