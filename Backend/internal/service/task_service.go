package service

import (
	"context"
	"fmt"

	"github.com/aistudio/backend/internal/task"
	"gorm.io/gorm"
)

// TaskService handles task business logic.
type TaskService struct {
	db      *gorm.DB
	manager *task.Manager
}

// NewTaskService creates a new TaskService.
func NewTaskService(db *gorm.DB, manager *task.Manager) *TaskService {
	return &TaskService{db: db, manager: manager}
}

// CreateTaskRequest represents the input for creating a task.
type CreateTaskRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Handler     string      `json:"handler"`
	Priority    int         `json:"priority"`
	Payload     interface{} `json:"payload"`
}

// List returns all tasks from the task manager.
func (s *TaskService) List(ctx context.Context) ([]*task.Task, error) {
	return s.manager.ListTasks(ctx)
}

// Get returns a single task by ID.
func (s *TaskService) Get(ctx context.Context, id string) (*task.Task, error) {
	return s.manager.GetTask(ctx, id)
}

// Create submits a new task to the task scheduler.
func (s *TaskService) Create(ctx context.Context, req CreateTaskRequest) (string, error) {
	priority := task.PriorityNormal
	if req.Priority >= 0 && req.Priority <= 3 {
		priority = task.Priority(req.Priority)
	}

	taskID, err := s.manager.Submit(ctx, req.Name, req.Description, req.Handler, priority, req.Payload)
	if err != nil {
		return "", fmt.Errorf("submit task: %w", err)
	}
	return taskID, nil
}

// Cancel cancels a running/pending task.
func (s *TaskService) Cancel(ctx context.Context, id string) error {
	return s.manager.Cancel(ctx, id)
}

// UpdateStatus updates the status of a task.
func (s *TaskService) UpdateStatus(ctx context.Context, id, status string) error {
	t, err := s.manager.GetTask(ctx, id)
	if err != nil {
		return fmt.Errorf("task not found: %s", id)
	}

	newStatus := task.Status(status)
	if err := task.ValidateTransition(t.Status, newStatus); err != nil {
		return err
	}

	if err := s.manager.UpdateStatus(ctx, id, newStatus); err != nil {
		return fmt.Errorf("update task status: %w", err)
	}
	return nil
}

// Delete removes a task.
func (s *TaskService) Delete(ctx context.Context, id string) error {
	return s.manager.DeleteTask(ctx, id)
}