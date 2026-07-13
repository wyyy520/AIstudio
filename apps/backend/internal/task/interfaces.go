package task

import "context"

// TaskHandler defines the interface for executing a task.
// Implementations should handle the actual task logic
// (e.g., running a workflow, training a model, calling an agent).
type TaskHandler interface {
	// Execute runs the task with the given payload.
	// Returns the result or an error.
	Execute(ctx context.Context, task *Task) (interface{}, error)
}

// HandlerFactory creates a TaskHandler by name.
// This allows the task system to look up handlers dynamically
// for different task types (workflow, agent, plugin, etc.).
type HandlerFactory interface {
	// CreateHandler returns a handler for the given name.
	CreateHandler(name string) (TaskHandler, error)
	// RegisterHandler registers a handler factory function.
	RegisterHandler(name string, factory func() TaskHandler)
}