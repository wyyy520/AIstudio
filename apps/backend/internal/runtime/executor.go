package runtime

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// Executor Kind
// ============================================================================

// ExecutorKind identifies the type of executor.
type ExecutorKind string

const (
	ExecutorLocal  ExecutorKind = "local"
	ExecutorDocker ExecutorKind = "docker"
	ExecutorSSH    ExecutorKind = "ssh"
)

// ============================================================================
// Command Executor Interface
// ============================================================================

// CommandExecutor executes commands with real-time log streaming.
// This is the low-level primitive 鈥?it knows nothing about projects or workflows.
type CommandExecutor interface {
	// Execute runs a command and returns the result.
	Execute(ctx context.Context, config RunConfig) *RunResult

	// Stop terminates a running execution.
	Stop(runID string) error

	// Status returns the status of a running execution.
	Status(runID string) (*RunStatus, bool)

	// ListRunning returns all currently running executions.
	ListRunning() []*RunStatus

	// SetLogger sets the log callback for all executions.
	SetLogger(logger func(LogEntry))
}

// ============================================================================
// Base Executor
// ============================================================================

// baseExecutor holds common executor state.
type baseExecutor struct {
	mu       sync.RWMutex
	running  map[string]*exec.Cmd
	runInfos map[string]*runInfo
	logger   func(LogEntry)
}

type runInfo struct {
	runID     string
	startedAt time.Time
	status    RunStatusEnum
}

func newBaseExecutor() *baseExecutor {
	return &baseExecutor{
		running:  make(map[string]*exec.Cmd),
		runInfos: make(map[string]*runInfo),
		logger: func(entry LogEntry) {
			log.Printf("[runtime] %s: %s", entry.Level, entry.Message)
		},
	}
}

// ============================================================================
// Local Executor 鈥?runs commands as subprocesses on the local machine
// ============================================================================

// LocalExecutor runs commands in a subprocess with real-time log streaming.
type LocalExecutor struct {
	*baseExecutor
}

// NewLocalExecutor creates a new LocalExecutor.
func NewLocalExecutor() *LocalExecutor {
	return &LocalExecutor{baseExecutor: newBaseExecutor()}
}

// SetLogger sets the log callback for all executions.
func (e *LocalExecutor) SetLogger(logger func(LogEntry)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.logger = logger
}

// Execute runs a command locally and returns the result.
func (e *LocalExecutor) Execute(ctx context.Context, config RunConfig) *RunResult {
	runID := uuid.New().String()
	startedAt := time.Now()

	e.mu.Lock()
	e.runInfos[runID] = &runInfo{
		runID:     runID,
		startedAt: startedAt,
		status:    RunStatusRunning,
	}
	e.mu.Unlock()

	// Prepare command
	cmd := exec.CommandContext(ctx, config.EntryPoint, config.Args...)
	cmd.Dir = config.ProjectDir

	// Set environment variables
	if config.Env != nil {
		cmd.Env = os.Environ()
		for k, v := range config.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return e.fail(runID, startedAt, fmt.Sprintf("stdout pipe error: %v", err))
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return e.fail(runID, startedAt, fmt.Sprintf("stderr pipe error: %v", err))
	}

	// Store the command for potential stop
	e.mu.Lock()
	e.running[runID] = cmd
	e.mu.Unlock()

	// Log execution start
	e.emitLog(LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Source:    "system",
		Message:   fmt.Sprintf("Executing: %s %s", config.EntryPoint, strings.Join(config.Args, " ")),
		RunID:     runID,
	})

	// Start the command
	if err := cmd.Start(); err != nil {
		e.cleanup(runID)
		return e.fail(runID, startedAt, fmt.Sprintf("start error: %v", err))
	}

	// Stream logs in real-time
	var wg sync.WaitGroup
	wg.Add(2)

	go e.streamLogs(runID, stdout, "stdout", &wg)
	go e.streamLogs(runID, stderr, "stderr", &wg)

	wg.Wait()

	// Wait for command to complete
	err = cmd.Wait()

	e.cleanup(runID)

	completedAt := time.Now()
	duration := completedAt.Sub(startedAt)

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		e.mu.Lock()
		e.runInfos[runID] = &runInfo{
			runID:     runID,
			startedAt: startedAt,
			status:    RunStatusTimeout,
		}
		e.mu.Unlock()

		e.emitLog(LogEntry{
			Timestamp: time.Now(),
			Level:     "ERROR",
			Source:    "system",
			Message:   "execution timed out",
			RunID:     runID,
		})

		return &RunResult{
			RunID:       runID,
			Status:      RunStatusTimeout,
			ExitCode:    -1,
			Duration:    duration,
			StartedAt:   startedAt,
			CompletedAt: completedAt,
			Error:       "execution timed out",
		}
	}

	exitCode := 0
	runStatus := RunStatusCompleted
	errMsg := ""
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		errMsg = err.Error()
		runStatus = RunStatusFailed
	}

	e.mu.Lock()
	e.runInfos[runID] = &runInfo{
		runID:     runID,
		startedAt: startedAt,
		status:    runStatus,
	}
	e.mu.Unlock()

	e.emitLog(LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Source:    "system",
		Message:   fmt.Sprintf("Execution completed: exitCode=%d, duration=%v", exitCode, duration),
		RunID:     runID,
	})

	return &RunResult{
		RunID:       runID,
		Status:      runStatus,
		ExitCode:    exitCode,
		Duration:    duration,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Error:       errMsg,
	}
}

// Stop terminates a running execution.
func (e *LocalExecutor) Stop(runID string) error {
	e.mu.Lock()
	cmd, ok := e.running[runID]
	e.mu.Unlock()

	if !ok {
		return fmt.Errorf("run not found: %s", runID)
	}

	if err := cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill process: %w", err)
	}

	e.mu.Lock()
	if info, exists := e.runInfos[runID]; exists {
		info.status = RunStatusStopped
	}
	e.mu.Unlock()

	return nil
}

// Status returns the status of a running execution.
func (e *LocalExecutor) Status(runID string) (*RunStatus, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	info, ok := e.runInfos[runID]
	if !ok {
		return nil, false
	}

	return &RunStatus{
		RunID:     info.runID,
		Status:    info.status,
		StartedAt: info.startedAt,
		Duration:  time.Since(info.startedAt),
	}, true
}

// ListRunning returns all running executions.
func (e *LocalExecutor) ListRunning() []*RunStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []*RunStatus
	for _, info := range e.runInfos {
		result = append(result, &RunStatus{
			RunID:     info.runID,
			Status:    info.status,
			StartedAt: info.startedAt,
			Duration:  time.Since(info.startedAt),
		})
	}
	return result
}

// ============================================================================
// Docker Executor 鈥?runs commands inside a Docker container
// ============================================================================

// DockerConfig configures the Docker executor.
type DockerConfig struct {
	Image        string            // Docker image (e.g. "python:3.11")
	Tag          string            // Image tag
	Volumes      []string          // Volume mounts (e.g. ["/host:/container"])
	Env          map[string]string // Environment variables
	Network      string            // Network mode
	WorkingDir   string            // Working directory inside container
	RemoveOnExit bool              // Auto-remove container after execution
	GPUs         bool              // Enable GPU access (--gpus all)
}

// DockerExecutor runs commands inside a Docker container.
type DockerExecutor struct {
	*baseExecutor
	config DockerConfig
}

// NewDockerExecutor creates a new DockerExecutor.
func NewDockerExecutor(cfg DockerConfig) *DockerExecutor {
	return &DockerExecutor{
		baseExecutor: newBaseExecutor(),
		config:       cfg,
	}
}

// Execute runs a command inside a Docker container.
func (e *DockerExecutor) Execute(ctx context.Context, config RunConfig) *RunResult {
	runID := uuid.New().String()
	startedAt := time.Now()

	e.mu.Lock()
	e.runInfos[runID] = &runInfo{
		runID:     runID,
		startedAt: startedAt,
		status:    RunStatusRunning,
	}
	e.mu.Unlock()

	// Build docker run args
	args := []string{"run"}
	args = append(args, "--name", "aistudio-"+runID)

	// Working directory
	if e.config.WorkingDir != "" {
		args = append(args, "-w", e.config.WorkingDir)
	}

	// Volume mounts
	for _, vol := range e.config.Volumes {
		args = append(args, "-v", vol)
	}

	// Network
	if e.config.Network != "" {
		args = append(args, "--network", e.config.Network)
	}

	// GPU access
	if e.config.GPUs {
		args = append(args, "--gpus", "all")
	}

	// Environment variables
	env := config.Env
	if env == nil {
		env = e.config.Env
	} else {
		for k, v := range e.config.Env {
			env[k] = v
		}
	}
	for k, v := range env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Remove on exit
	if e.config.RemoveOnExit {
		args = append(args, "--rm")
	}

	// Image
	image := e.config.Image
	if e.config.Tag != "" {
		image = image + ":" + e.config.Tag
	}
	args = append(args, image)

	// Entry point and args
	args = append(args, config.EntryPoint)
	args = append(args, config.Args...)

	dockerCmd := exec.CommandContext(ctx, "docker", args...)

	// Stream output
	stdout, err := dockerCmd.StdoutPipe()
	if err != nil {
		return e.fail(runID, startedAt, fmt.Sprintf("docker stdout pipe error: %v", err))
	}
	stderr, err := dockerCmd.StderrPipe()
	if err != nil {
		return e.fail(runID, startedAt, fmt.Sprintf("docker stderr pipe error: %v", err))
	}

	e.mu.Lock()
	e.running[runID] = dockerCmd
	e.mu.Unlock()

	e.emitLog(LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Source:    "system",
		Message:   fmt.Sprintf("Docker execute: %s", strings.Join(args, " ")),
		RunID:     runID,
	})

	if err := dockerCmd.Start(); err != nil {
		e.cleanup(runID)
		return e.fail(runID, startedAt, fmt.Sprintf("docker start error: %v", err))
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go e.streamLogs(runID, stdout, "stdout", &wg)
	go e.streamLogs(runID, stderr, "stderr", &wg)
	wg.Wait()

	err = dockerCmd.Wait()
	e.cleanup(runID)

	completedAt := time.Now()
	duration := completedAt.Sub(startedAt)

	exitCode := 0
	runStatus := RunStatusCompleted
	errMsg := ""
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		errMsg = err.Error()
		runStatus = RunStatusFailed
	}

	e.mu.Lock()
	e.runInfos[runID] = &runInfo{
		runID:     runID,
		startedAt: startedAt,
		status:    runStatus,
	}
	e.mu.Unlock()

	return &RunResult{
		RunID:       runID,
		Status:      runStatus,
		ExitCode:    exitCode,
		Duration:    duration,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Error:       errMsg,
	}
}

// Stop stops a Docker container.
func (e *DockerExecutor) Stop(runID string) error {
	// Stop the docker container
	stopCmd := exec.Command("docker", "stop", "aistudio-"+runID)
	if err := stopCmd.Run(); err != nil {
		return fmt.Errorf("docker stop failed: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	if info, exists := e.runInfos[runID]; exists {
		info.status = RunStatusStopped
	}
	return nil
}

// Status returns the status of a Docker execution.
func (e *DockerExecutor) Status(runID string) (*RunStatus, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	info, ok := e.runInfos[runID]
	if !ok {
		return nil, false
	}

	return &RunStatus{
		RunID:     info.runID,
		Status:    info.status,
		StartedAt: info.startedAt,
		Duration:  time.Since(info.startedAt),
	}, true
}

// ListRunning returns all running Docker executions.
func (e *DockerExecutor) ListRunning() []*RunStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []*RunStatus
	for _, info := range e.runInfos {
		result = append(result, &RunStatus{
			RunID:     info.runID,
			Status:    info.status,
			StartedAt: info.startedAt,
			Duration:  time.Since(info.startedAt),
		})
	}
	return result
}

// SetLogger sets the log callback for all Docker executions.
func (e *DockerExecutor) SetLogger(logger func(LogEntry)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.logger = logger
}

// ============================================================================
// SSH Executor 鈥?runs commands on a remote machine via SSH
// ============================================================================

// SSHConfig configures the SSH executor.
type SSHConfig struct {
	Host       string // Remote host (user@hostname)
	Port       int    // SSH port (default 22)
	KeyPath    string // Path to SSH private key
	KnownHosts string // Path to known_hosts file
}

// SSHExecutor runs commands on a remote machine via SSH.
type SSHExecutor struct {
	*baseExecutor
	config SSHConfig
}

// NewSSHExecutor creates a new SSHExecutor.
func NewSSHExecutor(cfg SSHConfig) *SSHExecutor {
	return &SSHExecutor{
		baseExecutor: newBaseExecutor(),
		config:       cfg,
	}
}

// Execute runs a command on a remote machine via SSH.
func (e *SSHExecutor) Execute(ctx context.Context, config RunConfig) *RunResult {
	runID := uuid.New().String()
	startedAt := time.Now()

	e.mu.Lock()
	e.runInfos[runID] = &runInfo{
		runID:     runID,
		startedAt: startedAt,
		status:    RunStatusRunning,
	}
	e.mu.Unlock()

	// Build SSH args
	sshArgs := []string{
		"-o", "StrictHostKeyChecking=yes",
	}

	if e.config.Port > 0 {
		sshArgs = append(sshArgs, "-p", fmt.Sprintf("%d", e.config.Port))
	}
	if e.config.KeyPath != "" {
		sshArgs = append(sshArgs, "-i", e.config.KeyPath)
	}

	sshArgs = append(sshArgs, e.config.Host)

	// Build remote command
	var cmdBuf bytes.Buffer
	if config.ProjectDir != "" {
		cmdBuf.WriteString(fmt.Sprintf("cd %s && ", config.ProjectDir))
	}
	cmdBuf.WriteString(config.EntryPoint)
	for _, arg := range config.Args {
		cmdBuf.WriteString(" ")
		cmdBuf.WriteString(escapeSSHArg(arg))
	}

	sshArgs = append(sshArgs, cmdBuf.String())

	sshCmd := exec.CommandContext(ctx, "ssh", sshArgs...)

	// Set environment for SSH
	if config.Env != nil {
		envPrefix := ""
		for k, v := range config.Env {
			envPrefix += fmt.Sprintf("%s=%s ", k, v)
		}
		// Prepend env vars to the command
		sshArgs[len(sshArgs)-1] = envPrefix + sshArgs[len(sshArgs)-1]
		sshCmd = exec.CommandContext(ctx, "ssh", sshArgs...)
	}

	stdout, err := sshCmd.StdoutPipe()
	if err != nil {
		return e.fail(runID, startedAt, fmt.Sprintf("ssh stdout pipe error: %v", err))
	}
	stderr, err := sshCmd.StderrPipe()
	if err != nil {
		return e.fail(runID, startedAt, fmt.Sprintf("ssh stderr pipe error: %v", err))
	}

	e.mu.Lock()
	e.running[runID] = sshCmd
	e.mu.Unlock()

	e.emitLog(LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Source:    "system",
		Message:   fmt.Sprintf("SSH execute on %s: %s", e.config.Host, config.EntryPoint),
		RunID:     runID,
	})

	if err := sshCmd.Start(); err != nil {
		e.cleanup(runID)
		return e.fail(runID, startedAt, fmt.Sprintf("ssh start error: %v", err))
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go e.streamLogs(runID, stdout, "stdout", &wg)
	go e.streamLogs(runID, stderr, "stderr", &wg)
	wg.Wait()

	err = sshCmd.Wait()
	e.cleanup(runID)

	completedAt := time.Now()
	duration := completedAt.Sub(startedAt)

	exitCode := 0
	runStatus := RunStatusCompleted
	errMsg := ""
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		errMsg = err.Error()
		runStatus = RunStatusFailed
	}

	e.mu.Lock()
	e.runInfos[runID] = &runInfo{
		runID:     runID,
		startedAt: startedAt,
		status:    runStatus,
	}
	e.mu.Unlock()

	return &RunResult{
		RunID:       runID,
		Status:      runStatus,
		ExitCode:    exitCode,
		Duration:    duration,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Error:       errMsg,
	}
}

// Stop terminates an SSH execution.
func (e *SSHExecutor) Stop(runID string) error {
	e.mu.Lock()
	cmd, ok := e.running[runID]
	e.mu.Unlock()

	if !ok {
		return fmt.Errorf("run not found: %s", runID)
	}
	if err := cmd.Process.Kill(); err != nil {
		return fmt.Errorf("failed to kill SSH process: %w", err)
	}

	e.mu.Lock()
	if info, exists := e.runInfos[runID]; exists {
		info.status = RunStatusStopped
	}
	e.mu.Unlock()
	return nil
}

// Status returns the status of an SSH execution.
func (e *SSHExecutor) Status(runID string) (*RunStatus, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	info, ok := e.runInfos[runID]
	if !ok {
		return nil, false
	}

	return &RunStatus{
		RunID:     info.runID,
		Status:    info.status,
		StartedAt: info.startedAt,
		Duration:  time.Since(info.startedAt),
	}, true
}

// ListRunning returns all running SSH executions.
func (e *SSHExecutor) ListRunning() []*RunStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []*RunStatus
	for _, info := range e.runInfos {
		result = append(result, &RunStatus{
			RunID:     info.runID,
			Status:    info.status,
			StartedAt: info.startedAt,
			Duration:  time.Since(info.startedAt),
		})
	}
	return result
}

// SetLogger sets the log callback for all SSH executions.
func (e *SSHExecutor) SetLogger(logger func(LogEntry)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.logger = logger
}

// ============================================================================
// Executor Factory
// ============================================================================

// NewExecutor creates the appropriate executor based on the given kind.
func NewExecutor(kind ExecutorKind, opts ...interface{}) CommandExecutor {
	switch kind {
	case ExecutorDocker:
		var cfg DockerConfig
		for _, opt := range opts {
			if c, ok := opt.(DockerConfig); ok {
				cfg = c
			}
		}
		return NewDockerExecutor(cfg)
	case ExecutorSSH:
		var cfg SSHConfig
		for _, opt := range opts {
			if c, ok := opt.(SSHConfig); ok {
				cfg = c
			}
		}
		return NewSSHExecutor(cfg)
	default:
		return NewLocalExecutor()
	}
}

// ============================================================================
// Private Helpers (shared by all executors)
// ============================================================================

func (e *baseExecutor) streamLogs(runID string, reader io.ReadCloser, source string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		entry := LogEntry{
			Timestamp: time.Now(),
			Level:     detectLevel(line, source),
			Source:    source,
			Message:   line,
			RunID:     runID,
			Raw:       line,
		}

		e.mu.RLock()
		logger := e.logger
		e.mu.RUnlock()

		if logger != nil {
			logger(entry)
		}
	}
}

func (e *baseExecutor) fail(runID string, startedAt time.Time, errMsg string) *RunResult {
	e.mu.Lock()
	if info, exists := e.runInfos[runID]; exists {
		info.status = RunStatusFailed
	}
	e.mu.Unlock()

	e.emitLog(LogEntry{
		Timestamp: time.Now(),
		Level:     "ERROR",
		Source:    "system",
		Message:   errMsg,
		RunID:     runID,
	})

	return &RunResult{
		RunID:     runID,
		Status:    RunStatusFailed,
		ExitCode:  -1,
		Duration:  time.Since(startedAt),
		StartedAt: startedAt,
		Error:     errMsg,
	}
}

func (e *baseExecutor) cleanup(runID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.running, runID)
}

func (e *baseExecutor) emitLog(entry LogEntry) {
	e.mu.RLock()
	logger := e.logger
	e.mu.RUnlock()
	if logger != nil {
		logger(entry)
	}
}

func detectLevel(line string, source string) string {
	if source == "stderr" {
		return "ERROR"
	}
	if len(line) > 0 {
		switch {
		case contains(line, "ERROR") || contains(line, "error"):
			return "ERROR"
		case contains(line, "WARN") || contains(line, "warning"):
			return "WARN"
		case contains(line, "DEBUG") || contains(line, "debug"):
			return "DEBUG"
		case contains(line, "INFO") || contains(line, "info"):
			return "INFO"
		}
	}
	return "INFO"
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsSubstring(s, substr)
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func escapeSSHArg(arg string) string {
	// Simple SSH argument escaping
	if strings.ContainsAny(arg, " \t\n\"'") {
		return fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "'\\''"))
	}
	return arg
}

// Ensure interface compliance
var _ CommandExecutor = (*LocalExecutor)(nil)
var _ CommandExecutor = (*DockerExecutor)(nil)
var _ CommandExecutor = (*SSHExecutor)(nil)
