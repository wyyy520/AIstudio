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
	ProjectID   string      `json:"project_id"`
	WorkflowID  string      `json:"workflow_id"`
	Type        string      `json:"type"`
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

// GetStatus returns the status response for a task.
func (s *TaskService) GetStatus(ctx context.Context, id string) (*task.TaskStatusResponse, error) {
	return s.manager.GetTaskStatus(ctx, id)
}

// Create submits a new task to the task scheduler.
// It creates the task in waiting state and immediately starts it.
func (s *TaskService) Create(ctx context.Context, req CreateTaskRequest) (string, error) {
	priority := task.PriorityNormal
	if req.Priority >= 0 && req.Priority <= 3 {
		priority = task.Priority(req.Priority)
	}

	taskType := task.TaskTypeWorkflow
	if req.Type != "" {
		switch req.Type {
		case "workflow":
			taskType = task.TaskTypeWorkflow
		case "agent":
			taskType = task.TaskTypeAgent
		case "plugin":
			taskType = task.TaskTypePlugin
		}
	}

	taskID, err := s.manager.CreateTask(ctx, req.ProjectID, req.WorkflowID, taskType, req.Name, req.Handler, priority, req.Payload)
	if err != nil {
		return "", fmt.Errorf("create task: %w", err)
	}

	// Set description
	if req.Description != "" {
		t, _ := s.manager.GetTask(ctx, taskID)
		if t != nil {
			t.Description = req.Description
		}
	}

	// Start the task immediately
	if err := s.manager.StartTask(ctx, taskID); err != nil {
		return taskID, fmt.Errorf("start task: %w", err)
	}

	return taskID, nil
}

// Cancel cancels a running/waiting task.
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

// EventBus returns the task manager's event bus for WebSocket streaming.
func (s *TaskService) EventBus() *task.EventBus {
	return s.manager.EventBus()
}