package executors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

type ProcessOptions struct {
	Path      string
	Args      []string
	Timeout   time.Duration
	Env       []string
	MaxMemory int64
}

type ProcessExecutor struct {
	path string
	opts ProcessOptions
}

func NewProcessExecutor(path string, opts ProcessOptions) *ProcessExecutor {
	return &ProcessExecutor{path: path, opts: opts}
}

func (e *ProcessExecutor) Execute(ctx context.Context, input map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	timeout := e.opts.Timeout
	if t, ok := config["timeout"].(float64); ok && t > 0 {
		timeout = time.Duration(t) * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.path, e.opts.Args...)
	cmd.Env = append(os.Environ(), e.opts.Env...)

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

	go func() {
		defer stdin.Close()
		json.NewEncoder(stdin).Encode(input)
	}()

	outData, err := io.ReadAll(stdout)
	if err != nil {
		cmd.Process.Kill()
		cmd.Wait()
		return nil, fmt.Errorf("read stdout: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("process exited with error: %w, stderr: %s", err, stderr.String())
		}
		return nil, fmt.Errorf("process exited with error: %w", err)
	}

	var result map[string]interface{}
	if len(outData) > 0 {
		if err := json.Unmarshal(outData, &result); err != nil {
			return nil, fmt.Errorf("decode output: %w, stderr: %s", err, stderr.String())
		}
	}

	return result, nil
}
