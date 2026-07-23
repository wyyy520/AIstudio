package executors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aistudio/backend/internal/plugin"
)

type PythonExecutor struct{}

func NewPythonExecutor() *PythonExecutor {
	return &PythonExecutor{}
}

func (e *PythonExecutor) Language() string {
	return "python"
}

func (e *PythonExecutor) Execute(ctx context.Context, p *plugin.Plugin, input map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	entryPoint := "main.py"
	if p.Manifest != nil && p.Manifest.EntryPoint != "" {
		entryPoint = p.Manifest.EntryPoint
	}

	scriptPath := filepath.Join(p.SourceDir, entryPoint)

	timeout := 60 * time.Second
	if t, ok := config["timeout"].(float64); ok && t > 0 {
		timeout = time.Duration(t) * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python", scriptPath)
	cmd.Dir = p.SourceDir
	cmd.Env = filterPluginEnv(os.Environ())

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
			return nil, fmt.Errorf("python execution failed: %w, stderr: %s", err, stderr.String())
		}
		return nil, fmt.Errorf("python execution failed: %w", err)
	}

	var result map[string]interface{}
	if len(outData) > 0 {
		if err := json.Unmarshal(outData, &result); err != nil {
			return nil, fmt.Errorf("decode output: %w, stderr: %s", err, stderr.String())
		}
	}

	return result, nil
}

// filterPluginEnv filters environment variables for plugin execution.
// It removes potentially dangerous variables and returns a clean environment.
func filterPluginEnv(env []string) []string {
	var filtered []string
	blocklist := map[string]bool{
		"AISTUDIO_API_KEY": true,
		"OPENAI_API_KEY":   true,
		"ANTHROPIC_API_KEY": true,
	}
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) > 0 && !blocklist[parts[0]] {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
