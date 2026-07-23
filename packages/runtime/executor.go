// Package runtime provides the unified execution engine for AIStudio.
//
// The Runtime layer is a pure execution engine with ZERO knowledge of:
//   - Workflows (DSL structure, node types, business logic)
//   - AI algorithms (training, inference, data processing)
//
// Runtime ONLY executes standard OS commands (python, matlab, docker, ssh)
// and manages process lifecycles.
//
// This file implements three Executors:
//
//  1. LocalExecutor — runs commands as local OS processes
//     Steps:
//     a. Create exec.Cmd with context (supports cancellation/timeout)
//     b. Pipe stdout/stderr for real-time log streaming
//     c. Start → stream logs concurrently → Wait → collect exit code
//     d. Handle timeout via context.DeadlineExceeded
//
//  2. DockerExecutor — runs commands inside Docker containers
//     Steps:
//     a. Build docker run args (volumes, env, GPU, network, --rm)
//     b. Execute "docker run <image> <entrypoint> <args>"
//     c. Same log streaming and exit code collection as LocalExecutor
//
//  3. SSHExecutor — runs commands on remote machines via SSH
//     Steps:
//     a. Build "ssh" args (host, port, key, StrictHostKeyChecking)
//     b. Execute remote command via SSH
//     c. Same log streaming and exit code collection
//
// All three executors share a baseExecutor with:
//   - Run tracking: running map (cmd → runID), runInfos map (runID → status)
//   - Log streaming: Scanner-based line-by-line stdout/stderr capture
//   - Log level detection: heuristic keyword matching on output lines
//   - Panic-safe log emission via sync.RWMutex-protected logger callback
//
// EngStudio.md §9 — Runtime Execution Layer
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
// CommandExecutor Interface — the contract all executors must fulfill
// ============================================================================

// CommandExecutor is the single interface for process execution.
// Implementations: LocalExecutor, DockerExecutor, SSHExecutor.
type CommandExecutor interface {
	Execute(ctx context.Context, config RunConfig) *RunResult
	Stop(runID string) error
	Status(runID string) (*RunStatus, bool)
	ListRunning() []*RunStatus
	SetLogger(logger func(LogEntry))
}

// ============================================================================
// baseExecutor — shared infrastructure for all executor implementations
// ============================================================================

// baseExecutor manages run tracking, log streaming, and cleanup.
// All concrete executors embed this type.
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

// newBaseExecutor creates the shared base with default settings.
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
// LocalExecutor — native OS process execution
// ============================================================================

// LocalExecutor runs commands as local OS processes via os/exec.
// It's the default executor for desktop use.
type LocalExecutor struct {
	*baseExecutor
}

// NewLocalExecutor creates a new local process executor.
func NewLocalExecutor() *LocalExecutor {
	return &LocalExecutor{baseExecutor: newBaseExecutor()}
}

func (e *baseExecutor) SetLogger(logger func(LogEntry)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.logger = logger
}

// ============================================================================
// Execute — LocalExecutor implementation
// ============================================================================

// Execute runs a command as a local OS process.
//
// Steps:
//  1. Generate a unique runID
//  2. Build exec.Cmd from EntryPoint + Args, set WorkingDir + Env
//  3. Create stdout/stderr pipes for real-time log capture
//  4. Start the process (non-blocking)
//  5. Stream stdout and stderr concurrently via goroutines
//  6. Wait for process completion
//  7. Collect exit code and determine final status (Completed/Failed/Timeout)
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

	cmd := exec.CommandContext(ctx, config.EntryPoint, config.Args...)
	cmd.Dir = config.ProjectDir

	if config.Env != nil {
		cmd.Env = os.Environ()
		for k, v := range config.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	setProcessGroup(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return e.fail(runID, startedAt, fmt.Sprintf("stdout pipe error: %v", err))
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return e.fail(runID, startedAt, fmt.Sprintf("stderr pipe error: %v", err))
	}

	e.mu.Lock()
	e.running[runID] = cmd
	e.mu.Unlock()

	e.emitLog(LogEntry{
		Timestamp: time.Now(),
		Level:     "INFO",
		Source:    "system",
		Message:   fmt.Sprintf("Executing: %s %s", config.EntryPoint, strings.Join(config.Args, " ")),
		RunID:     runID,
	})

	if err := cmd.Start(); err != nil {
		e.cleanup(runID)
		return e.fail(runID, startedAt, fmt.Sprintf("start error: %v", err))
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go e.streamLogs(runID, stdout, "stdout", &wg)
	go e.streamLogs(runID, stderr, "stderr", &wg)

	wg.Wait()

	err = cmd.Wait()

	e.cleanup(runID)

	completedAt := time.Now()
	duration := completedAt.Sub(startedAt)

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

// ============================================================================
// Stop — LocalExecutor process group termination
// ============================================================================

// Stop terminates a running process by killing its process group.
// This ensures all child processes are also terminated.
func (e *LocalExecutor) Stop(runID string) error {
	e.mu.Lock()
	cmd, ok := e.running[runID]
	e.mu.Unlock()

	if !ok {
		return fmt.Errorf("run not found: %s", runID)
	}

	if cmd.Process == nil {
		return fmt.Errorf("process not started for run: %s", runID)
	}

	if err := killProcessGroup(cmd); err != nil {
		return fmt.Errorf("failed to kill process group: %w", err)
	}

	e.mu.Lock()
	if info, exists := e.runInfos[runID]; exists {
		info.status = RunStatusStopped
	}
	e.mu.Unlock()

	return nil
}

// ============================================================================
// Status & List — query execution state
// ============================================================================

// Status returns the current status of a running execution.
func (e *baseExecutor) Status(runID string) (*RunStatus, bool) {
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

// ListRunning returns all currently tracked executions.
func (e *baseExecutor) ListRunning() []*RunStatus {
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
// DockerExecutor — containerized execution
// ============================================================================

// DockerExecutor runs commands inside Docker containers.
type DockerExecutor struct {
	*baseExecutor
	config DockerConfig
}

// NewDockerExecutor creates a new Docker executor.
func NewDockerExecutor(cfg DockerConfig) *DockerExecutor {
	return &DockerExecutor{
		baseExecutor: newBaseExecutor(),
		config:       cfg,
	}
}

// Execute — DockerExecutor implementation
//
// Steps:
//  1. Generate runID and build "docker run" arguments from config
//  2. Attach volumes, env vars, GPU support, network, --rm flag
//  3. Execute "docker run <image> <entrypoint> <args>"
//  4. Stream logs, wait, collect exit code (same pattern as LocalExecutor)
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

	args := []string{"run"}
	args = append(args, "--name", "aistudio-"+runID)

	if e.config.WorkingDir != "" {
		args = append(args, "-w", e.config.WorkingDir)
	}

	for _, vol := range e.config.Volumes {
		args = append(args, "-v", vol)
	}

	if e.config.Network != "" {
		args = append(args, "--network", e.config.Network)
	}

	if e.config.GPUs {
		args = append(args, "--gpus", "all")
	}

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

	if e.config.RemoveOnExit {
		args = append(args, "--rm")
	}

	image := e.config.Image
	if e.config.Tag != "" {
		image = image + ":" + e.config.Tag
	}
	args = append(args, image)

	args = append(args, config.EntryPoint)
	args = append(args, config.Args...)

	dockerCmd := exec.CommandContext(ctx, "docker", args...)

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

// Stop — DockerExecutor sends "docker stop <container>".
func (e *DockerExecutor) Stop(runID string) error {
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

// ============================================================================
// SSHExecutor — remote execution via SSH
// ============================================================================

// SSHExecutor runs commands on remote machines via SSH.
type SSHExecutor struct {
	*baseExecutor
	config SSHConfig
}

// NewSSHExecutor creates a new SSH executor.
func NewSSHExecutor(cfg SSHConfig) *SSHExecutor {
	return &SSHExecutor{
		baseExecutor: newBaseExecutor(),
		config:       cfg,
	}
}

// Execute — SSHExecutor implementation
//
// Steps:
//  1. Build SSH args: host, port, keypath, StrictHostKeyChecking
//  2. Build remote command: "cd <projectDir> && <entryPoint> <args>"
//  3. Inject environment variables as prefix (KEY=VALUE cmd)
//  4. Execute "ssh <args> <remoteCommand>"
//  5. Stream logs, wait, collect exit code (same pattern)
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

	sshArgs := []string{"-o", "StrictHostKeyChecking=accept-new"}

	if e.config.Port > 0 {
		sshArgs = append(sshArgs, "-p", fmt.Sprintf("%d", e.config.Port))
	}
	if e.config.KeyPath != "" {
		sshArgs = append(sshArgs, "-i", e.config.KeyPath)
	}

	sshArgs = append(sshArgs, e.config.Host)

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

	if config.Env != nil {
		envPrefix := ""
		for k, v := range config.Env {
			envPrefix += fmt.Sprintf("%s=%s ", k, v)
		}
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

// Stop — SSHExecutor kills the local SSH process.
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

// ============================================================================
// Executor Factory — create the right executor by kind
// ============================================================================

// NewExecutor creates a CommandExecutor based on the given ExecutorKind.
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
// Log Streaming — real-time stdout/stderr capture
// ============================================================================

// streamLogs reads lines from a reader and emits them via the logger callback.
// It runs in a goroutine and signals completion via WaitGroup.
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

// ============================================================================
// Internal Helpers — fail, cleanup, emitLog, detectLevel
// ============================================================================

// fail records a failure and returns a RunResult with error information.
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

// cleanup removes a completed run's command from the running map.
func (e *baseExecutor) cleanup(runID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.running, runID)
}

// emitLog thread-safely calls the configured logger callback.
func (e *baseExecutor) emitLog(entry LogEntry) {
	e.mu.RLock()
	logger := e.logger
	e.mu.RUnlock()
	if logger != nil {
		logger(entry)
	}
}

// detectLevel uses heuristic keyword matching to guess the log level from
// a line of output text. stderr lines are always ERROR.
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
	if strings.ContainsAny(arg, " \t\n\"'") {
		return fmt.Sprintf("'%s'", strings.ReplaceAll(arg, "'", "'\\''"))
	}
	return arg
}

// ============================================================================
// Compile-time interface compliance check
// ============================================================================

var _ CommandExecutor = (*LocalExecutor)(nil)
var _ CommandExecutor = (*DockerExecutor)(nil)
var _ CommandExecutor = (*SSHExecutor)(nil)
