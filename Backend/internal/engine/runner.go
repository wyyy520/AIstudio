package engine

import (
	"context"
	"fmt"
	"log"

	"github.com/aistudio/backend/internal/task"
)

// TaskHandler bridges the Go task system with the Python AI Engine.
// It implements task.TaskHandler so that Python-based AI tasks
// (YOLO training, inference, etc.) can be submitted through the
// task scheduler.
type TaskHandler struct {
	runner  *PythonRunner
	taskMgr *task.Manager
}

// NewTaskHandler creates a new Engine TaskHandler.
func NewTaskHandler(runner *PythonRunner) *TaskHandler {
	return &TaskHandler{runner: runner}
}

// SetTaskManager sets the task manager for lifecycle management.
func (h *TaskHandler) SetTaskManager(mgr *task.Manager) {
	h.taskMgr = mgr
	h.runner.SetTaskManager(mgr)
}

// Execute runs a Python engine task.
// The task.Payload should contain:
//
//	{
//	  "plugin": "yolo",
//	  "action": "train",
//	  "params": { ... }
//	}
func (h *TaskHandler) Execute(ctx context.Context, t *task.Task) (interface{}, error) {
	payload, ok := t.Payload.(map[string]interface{})
	if !ok {
		err := fmt.Errorf("invalid engine payload: expected map[string]interface{}, got %T", t.Payload)
		if h.taskMgr != nil {
			_ = h.taskMgr.FailTask(ctx, t.ID, err.Error())
		}
		return nil, err
	}

	plugin, _ := payload["plugin"].(string)
	action, _ := payload["action"].(string)
	params, _ := payload["params"].(map[string]interface{})

	if plugin == "" {
		err := fmt.Errorf("missing 'plugin' field in payload")
		if h.taskMgr != nil {
			_ = h.taskMgr.FailTask(ctx, t.ID, err.Error())
		}
		return nil, err
	}
	if action == "" {
		err := fmt.Errorf("missing 'action' field in payload")
		if h.taskMgr != nil {
			_ = h.taskMgr.FailTask(ctx, t.ID, err.Error())
		}
		return nil, err
	}

	if params == nil {
		params = make(map[string]interface{})
	}

	log.Printf("[engine-handler] executing task: id=%s plugin=%s action=%s",
		t.ID, plugin, action)

	// Update progress to indicate execution has started
	if h.taskMgr != nil {
		_ = h.taskMgr.UpdateProgress(ctx, t.ID, 0.05)
	}

	// Run the Python engine
	input := TaskInput{
		TaskID: t.ID,
		Plugin: plugin,
		Action: action,
		Params: params,
	}

	result, err := h.runner.Run(ctx, input)
	if err != nil {
		errMsg := fmt.Sprintf("engine execution failed: %v", err)
		if h.taskMgr != nil {
			_ = h.taskMgr.FailTask(ctx, t.ID, errMsg)
		}
		return nil, fmt.Errorf(errMsg)
	}

	// Build the result map for the task system
	output := map[string]interface{}{
		"status":     result.Status,
		"model_path": result.ModelPath,
	}

	if result.Metrics != nil {
		output["metrics"] = result.Metrics
	}
	if result.Error != "" {
		output["error"] = result.Error
	}

	log.Printf("[engine-handler] task %s completed: status=%s model=%s",
		t.ID, result.Status, result.ModelPath)

	return output, nil
}

// RunPluginAction is a convenience method to run a plugin action directly
// without going through the task system. Useful for sync operations.
func (h *TaskHandler) RunPluginAction(ctx context.Context, taskID, plugin, action string, params map[string]interface{}) (map[string]interface{}, error) {
	input := TaskInput{
		TaskID: taskID,
		Plugin: plugin,
		Action: action,
		Params: params,
	}
	result, err := h.runner.Run(ctx, input)
	if err != nil {
		return nil, err
	}

	output := map[string]interface{}{
		"status":     result.Status,
		"model_path": result.ModelPath,
	}
	if result.Metrics != nil {
		output["metrics"] = result.Metrics
	}
	if result.Error != "" {
		output["error"] = result.Error
	}
	return output, nil
}

// CheckEnvironment runs the Python environment detection.
func (h *TaskHandler) CheckEnvironment(ctx context.Context) (map[string]interface{}, error) {
	return h.runner.CheckEnvironment(ctx)
}