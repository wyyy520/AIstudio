package workflow

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aistudio/backend/internal/task"
)

// TaskHandler bridges the task system with the workflow engine.
// It implements task.TaskHandler so that workflow runs can be
// submitted through the task scheduler.
type TaskHandler struct {
	engine *Engine
}

// NewTaskHandler creates a new workflow TaskHandler.
func NewTaskHandler(engine *Engine) *TaskHandler {
	return &TaskHandler{engine: engine}
}

// Execute runs a workflow as a task.
// The task.Payload should contain the workflow definition JSON.
func (h *TaskHandler) Execute(ctx context.Context, t *task.Task) (interface{}, error) {
	payload, ok := t.Payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid workflow payload: expected map[string]interface{}, got %T", t.Payload)
	}

	workflowJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal workflow payload: %w", err)
	}

	result, err := h.engine.Run(ctx, workflowJSON)
	if err != nil {
		return nil, fmt.Errorf("workflow execution failed: %w", err)
	}

	return result, nil
}