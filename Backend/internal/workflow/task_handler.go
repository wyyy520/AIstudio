package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aistudio/backend/internal/task"
)

// EnvironmentChecker is the interface for pre-execution environment validation.
// Implemented by the environment package to avoid circular imports.
type EnvironmentChecker interface {
	Check() interface{}
	GetStatus() interface{}
}

// TaskHandler bridges the task system with the workflow engine.
// It implements task.TaskHandler so that workflow runs can be
// submitted through the task scheduler.
type TaskHandler struct {
	engine     *Engine
	envChecker EnvironmentChecker
	taskMgr    *task.Manager
}

// NewTaskHandler creates a new workflow TaskHandler.
func NewTaskHandler(engine *Engine) *TaskHandler {
	return &TaskHandler{engine: engine}
}

// SetEnvironmentChecker sets the environment checker for pre-execution validation.
func (h *TaskHandler) SetEnvironmentChecker(checker EnvironmentChecker) {
	h.envChecker = checker
}

// SetTaskManager sets the task manager reference for lifecycle management.
// When set, the handler will auto-manage task progress and completion.
func (h *TaskHandler) SetTaskManager(mgr *task.Manager) {
	h.taskMgr = mgr
}

// Execute runs a workflow as a task.
// The task.Payload should contain the workflow definition JSON.
// Before execution, it validates the environment is ready.
func (h *TaskHandler) Execute(ctx context.Context, t *task.Task) (interface{}, error) {
	payload, ok := t.Payload.(map[string]interface{})
	if !ok {
		err := fmt.Errorf("invalid workflow payload: expected map[string]interface{}, got %T", t.Payload)
		if h.taskMgr != nil {
			_ = h.taskMgr.FailTask(ctx, t.ID, err.Error())
		}
		return nil, err
	}

	// ---- Pre-execution Environment Check ----
	if h.envChecker != nil {
		log.Printf("[workflow-task-handler] running environment check before workflow: %s", t.ID)
		if h.taskMgr != nil {
			_ = h.taskMgr.UpdateProgress(ctx, t.ID, 0.05)
		}

		checkResult := h.envChecker.Check()
		// If the check result has issues, log them but continue
		// (we don't block execution, just warn)
		if checkResult != nil {
			log.Printf("[workflow-task-handler] environment check completed for task: %s", t.ID)
		}
	}

	workflowJSON, err := json.Marshal(payload)
	if err != nil {
		err = fmt.Errorf("failed to marshal workflow payload: %w", err)
		if h.taskMgr != nil {
			_ = h.taskMgr.FailTask(ctx, t.ID, err.Error())
		}
		return nil, err
	}

	// Update progress to indicate execution has started
	if h.taskMgr != nil {
		_ = h.taskMgr.UpdateProgress(ctx, t.ID, 0.1)
	}

	log.Printf("[workflow-task-handler] executing workflow for task: %s", t.ID)

	result, err := h.engine.Run(ctx, workflowJSON)
	if err != nil {
		errMsg := fmt.Sprintf("workflow execution failed: %v", err)
		if h.taskMgr != nil {
			_ = h.taskMgr.FailTask(ctx, t.ID, errMsg)
		}
		return nil, fmt.Errorf(errMsg)
	}

	// Update progress based on engine result
	if h.taskMgr != nil {
		_ = h.taskMgr.UpdateProgress(ctx, t.ID, result.Progress)
	}

	log.Printf("[workflow-task-handler] workflow completed for task: %s (status: %s)", t.ID, result.Status)
	return result, nil
}

// RunWorkflowWithTask creates a task, starts it, and runs the workflow.
// This is the main entry point for workflow execution with full task lifecycle.
func (h *TaskHandler) RunWorkflowWithTask(ctx context.Context, projectID, workflowID string, workflowJSON []byte) (string, *ExecutionResult, error) {
	if h.taskMgr == nil {
		return "", nil, fmt.Errorf("task manager not set on workflow task handler")
	}

	// Parse the workflow to create a proper task name
	var wf map[string]interface{}
	var taskName string
	if err := json.Unmarshal(workflowJSON, &wf); err == nil {
		if name, ok := wf["name"].(string); ok {
			taskName = name
		}
	}
	if taskName == "" {
		taskName = "Workflow Run"
	}

	// Create the task
	taskID, err := h.taskMgr.CreateTask(ctx, projectID, workflowID, task.TaskTypeWorkflow, taskName, "workflow", task.PriorityNormal, wf)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Start the task
	if err := h.taskMgr.StartTask(ctx, taskID); err != nil {
		return taskID, nil, fmt.Errorf("failed to start task: %w", err)
	}

	return taskID, nil, nil
}