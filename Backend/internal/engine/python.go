package engine

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/aistudio/backend/internal/task"
)

// PythonRunner manages the lifecycle of a Python subprocess execution.
type PythonRunner struct {
	pythonPath string
	engineDir  string
	taskMgr    *task.Manager
	mu         sync.Mutex
}

// NewPythonRunner creates a new PythonRunner.
func NewPythonRunner(pythonPath, engineDir string) *PythonRunner {
	if pythonPath == "" {
		pythonPath = "python"
	}
	return &PythonRunner{
		pythonPath: pythonPath,
		engineDir:  engineDir,
	}
}

// SetTaskManager sets the task manager for progress updates.
func (r *PythonRunner) SetTaskManager(mgr *task.Manager) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.taskMgr = mgr
}

// TaskInput is the JSON structure passed to the Python runner via task.json.
type TaskInput struct {
	TaskID string                 `json:"task_id"`
	Plugin string                 `json:"plugin"`
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
}

// RunResult is the final result from Python execution.
type RunResult struct {
	Status    string                 `json:"status"`
	ModelPath string                 `json:"model_path"`
	Metrics   map[string]interface{} `json:"metrics"`
	Error     string                 `json:"error"`
}

// stdoutEvent is a single JSON line from the Python runner stdout.
type stdoutEvent struct {
	Type      string                 `json:"type"`
	Timestamp string                 `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Run executes a Python task by writing a task.json, launching the runner,
// and streaming stdout events back to the task manager.
func (r *PythonRunner) Run(ctx context.Context, input TaskInput) (*RunResult, error) {
	// Create a temp directory for this task
	taskDir, err := os.MkdirTemp("", "aistudio-task-"+input.TaskID)
	if err != nil {
		return nil, fmt.Errorf("create task dir: %w", err)
	}
	defer os.RemoveAll(taskDir)

	// Write task.json
	taskPath := filepath.Join(taskDir, "task.json")
	taskData, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal task: %w", err)
	}
	if err := os.WriteFile(taskPath, taskData, 0644); err != nil {
		return nil, fmt.Errorf("write task.json: %w", err)
	}

	log.Printf("[engine] task.json written: %s (plugin=%s, action=%s)",
		taskPath, input.Plugin, input.Action)

	// Build the Python command
	runnerScript := filepath.Join(r.engineDir, "runner.py")
	cmd := exec.CommandContext(ctx, r.pythonPath, runnerScript, "--task", taskPath)
	cmd.Dir = r.engineDir

	// Capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("create stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start python: %w", err)
	}

	log.Printf("[engine] python process started: pid=%d", cmd.Process.Pid)

	// Parse stdout events
	var finalResult *RunResult
	var resultMu sync.Mutex
	var parseErr error

	// Read stderr in background
	go func() {
		stderrReader := bufio.NewReader(stderr)
		for {
			line, err := stderrReader.ReadString('\n')
			if line != "" {
				log.Printf("[engine][stderr] %s", line)
			}
			if err != nil {
				break
			}
		}
	}()

	// Read stdout line by line
	stdoutReader := bufio.NewReader(stdout)
	for {
		line, readErr := stdoutReader.ReadString('\n')
		if line != "" {
			event := r.parseLine(line)
			if event != nil {
				switch event.Type {
				case "progress":
					r.handleProgress(ctx, input.TaskID, event)
				case "log":
					r.handleLog(ctx, input.TaskID, event)
				case "result":
					resultMu.Lock()
					finalResult = r.handleResult(event)
					resultMu.Unlock()
				case "error":
					resultMu.Lock()
					finalResult = &RunResult{
						Status: "failed",
						Error:  r.getStringField(event.Data, "message"),
					}
					resultMu.Unlock()
				}
			}
		}
		if readErr != nil {
			if readErr != io.EOF {
				parseErr = fmt.Errorf("stdout read error: %w", readErr)
			}
			break
		}
	}

	// Wait for process to complete
	waitErr := cmd.Wait()

	resultMu.Lock()
	defer resultMu.Unlock()

	if finalResult != nil {
		if finalResult.Error != "" {
			log.Printf("[engine] task %s failed: %s", input.TaskID, finalResult.Error)
			return finalResult, nil
		}
		log.Printf("[engine] task %s completed: status=%s, model=%s",
			input.TaskID, finalResult.Status, finalResult.ModelPath)
		return finalResult, nil
	}

	if parseErr != nil {
		return nil, parseErr
	}

	if waitErr != nil {
		return nil, fmt.Errorf("python process error: %w", waitErr)
	}

	return &RunResult{
		Status: "success",
	}, nil
}

// parseLine parses a single JSON line from stdout.
func (r *PythonRunner) parseLine(line string) *stdoutEvent {
	// Trim trailing newline/carriage return
	for len(line) > 0 && (line[len(line)-1] == '\n' || line[len(line)-1] == '\r') {
		line = line[:len(line)-1]
	}
	if line == "" {
		return nil
	}

	var event stdoutEvent
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		// Not a JSON line, treat as log
		log.Printf("[engine][stdout] %s", line)
		return nil
	}
	return &event
}

// handleProgress updates the task progress via the task manager.
func (r *PythonRunner) handleProgress(ctx context.Context, taskID string, event *stdoutEvent) {
	r.mu.Lock()
	mgr := r.taskMgr
	r.mu.Unlock()

	if mgr == nil {
		return
	}

	epoch := r.getFloatField(event.Data, "epoch")
	total := r.getFloatField(event.Data, "total_epochs")
	loss := r.getFloatField(event.Data, "loss")

	var progress float64
	if total > 0 {
		progress = epoch / total
	}

	_ = mgr.UpdateProgress(ctx, taskID, progress)

	// Emit a progress event with details
	if eventBus := mgr.EventBus(); eventBus != nil {
		eventBus.EmitTaskProgress(&task.Task{
			ID:       taskID,
			Progress: progress,
			Status:   task.StatusRunning,
		})
	}

	log.Printf("[engine][progress] task=%s epoch=%.0f/%.0f loss=%.4f progress=%.2f%%",
		taskID, epoch, total, loss, progress*100)
}

// handleLog forwards log events to the task manager.
func (r *PythonRunner) handleLog(ctx context.Context, taskID string, event *stdoutEvent) {
	level := r.getStringField(event.Data, "level")
	message := r.getStringField(event.Data, "message")
	log.Printf("[engine][%s] task=%s: %s", level, taskID, message)
}

// handleResult parses the final result event.
func (r *PythonRunner) handleResult(event *stdoutEvent) *RunResult {
	return &RunResult{
		Status:    r.getStringField(event.Data, "status"),
		ModelPath: r.getStringField(event.Data, "model_path"),
		Metrics:   r.getMapField(event.Data, "metrics"),
		Error:     r.getStringField(event.Data, "error"),
	}
}

// ---- helper methods ----

func (r *PythonRunner) getStringField(data map[string]interface{}, key string) string {
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (r *PythonRunner) getFloatField(data map[string]interface{}, key string) float64 {
	if v, ok := data[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case json.Number:
			f, _ := val.Float64()
			return f
		}
	}
	return 0
}

func (r *PythonRunner) getMapField(data map[string]interface{}, key string) map[string]interface{} {
	if v, ok := data[key]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

// CheckEnvironment runs the Python environment detection and returns the result.
func (r *PythonRunner) CheckEnvironment(ctx context.Context) (map[string]interface{}, error) {
	runnerScript := filepath.Join(r.engineDir, "runner.py")
	cmd := exec.CommandContext(ctx, r.pythonPath, runnerScript, "--env-check")
	cmd.Dir = r.engineDir

	// Capture stdout only
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("env check failed: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("parse env check result: %w", err)
	}

	return result, nil
}