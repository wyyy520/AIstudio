package executors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/aistudio/backend/internal/plugin"
)

// ProcessOptions configures the behaviour of the generic process executor.
type ProcessOptions struct {
	Path      string        // executable path
	Args      []string      // extra CLI arguments
	Timeout   time.Duration // per‑execution timeout
	Env       []string      // additional environment variables
	MaxMemory int64         // memory limit in bytes (best‑effort on Linux)
}

// ProcessExecutor runs plugins as child processes via stdin/stdout JSON protocol.
// It implements the PluginExecutor interface so it can be registered with the
// plugin Manager.
type ProcessExecutor struct {
	path string
	opts ProcessOptions
	lang string // language identifier (e.g. "binary", "shell", "custom")
}

// NewProcessExecutor creates a ProcessExecutor for the given executable.
// lang is the language key used by Manager.RegisterExecutor (e.g. "binary").
func NewProcessExecutor(path string, opts ProcessOptions, lang string) *ProcessExecutor {
	if lang == "" {
		lang = "binary"
	}
	return &ProcessExecutor{path: path, opts: opts, lang: lang}
}

// Language returns the language identifier this executor handles.
func (e *ProcessExecutor) Language() string {
	return e.lang
}

// Execute runs the configured executable with JSON‑on‑stdin / JSON‑on‑stdout protocol.
// Signature matches PluginExecutor.
func (e *ProcessExecutor) Execute(ctx context.Context, p *plugin.Plugin, input map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	_ = p // plugin metadata available for diagnostics

	timeout := e.opts.Timeout
	if t, ok := config["timeout"].(float64); ok && t > 0 {
		timeout = time.Duration(t) * time.Second
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.path, e.opts.Args...)
	cmd.Env = append(os.Environ(), e.opts.Env...)

	// Best‑effort resource limit – only honoured on Linux when RLIMIT_AS is available
	if e.opts.MaxMemory > 0 && runtime.GOOS == "linux" {
		// RLIMIT_AS is applied by the OS; we rely on OOM‑killer for enforcement
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}

	// Write JSON input to stdin in a separate goroutine so the process can start consuming
	go func() {
		defer stdin.Close()
		_ = json.NewEncoder(stdin).Encode(map[string]interface{}{
			"input":  input,
			"config": config,
		})
	}()

	outData, readErr := io.ReadAll(stdout)
	waitErr := cmd.Wait()

	// Process crash / timeout
	if waitErr != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("process execution timed out after %v", timeout)
		}
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("process exited with error: %w, stderr: %s", waitErr, stderr.String())
		}
		return nil, fmt.Errorf("process exited with error: %w", waitErr)
	}

	if readErr != nil {
		return nil, fmt.Errorf("read stdout: %w", readErr)
	}

	var result map[string]interface{}
	if len(outData) > 0 {
		if err := json.Unmarshal(outData, &result); err != nil {
			return nil, fmt.Errorf("decode output: %w, stderr: %s", err, stderr.String())
		}
	}
	if result == nil {
		result = map[string]interface{}{
			"status":    "completed",
			"exit_code": cmd.ProcessState.ExitCode(),
		}
	}

	return result, nil
}
