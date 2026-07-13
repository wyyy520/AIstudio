// Package runtime provides the unified execution engine for AIStudio.
//
// Runtime is a pure execution engine with ZERO knowledge of:
//   - Workflows
//   - Business logic
//   - Algorithm logic
//
// Runtime ONLY executes standard commands (python, matlab, docker, etc.)
// and manages runtime bundles (shared, cached environments).
//
// Responsibilities:
//   1. Environment detection (check if required tools exist)
//   2. Runtime Bundle management (install, cache, share across projects)
//   3. Project execution (call standard commands)
//   4. Process lifecycle management (start, stop, monitor)
//   5. Log streaming (real-time output capture)
//   6. Error handling and status reporting
package runtime

import (
	"context"
	"time"
)

// ============================================================================
// BundleSpec — declarative bundle definition
// ============================================================================

// BundleSpec declares a runtime bundle that can be installed and shared.
// These specs are loaded from bundle.json files (e.g. packages/bundles/*/bundle.json).
type BundleSpec struct {
	Name         string   `json:"name"`         // Bundle name (e.g. "yolo", "transformer")
	Version      string   `json:"version"`      // Bundle version (e.g. "1.0.0")
	Python       string   `json:"python"`       // Python version requirement (e.g. ">=3.9")
	Packages     []string `json:"packages"`     // Required pip packages (e.g. ["torch>=2.0", "ultralytics"])
	Commands     []string `json:"commands"`     // Required system commands (e.g. ["python3", "docker"])
	GPUOptional  bool     `json:"gpu_optional"` // Whether GPU is optional or required
	Description  string   `json:"description,omitempty"`
}

// ============================================================================
// Runtime Interface
// ============================================================================

// Runtime is the unified execution engine.
// It does not know about Workflow — it only executes projects.
type Runtime interface {
	// Detect checks if the environment meets the runtime requirements.
	Detect(ctx context.Context, req *Requirement) (*EnvironmentReport, error)

	// Prepare installs the runtime bundle and prepares the environment.
	Prepare(ctx context.Context, req *Requirement) error

	// Execute runs the project with the given configuration.
	Execute(ctx context.Context, project *Project, config *RunConfig) (*RunResult, error)

	// Stop terminates a running execution.
	Stop(ctx context.Context, runID string) error

	// Status returns the current status of a running execution.
	Status(ctx context.Context, runID string) (*RunStatus, error)

	// ListRunning returns all currently running executions.
	ListRunning() []*RunStatus
}

// ============================================================================
// Types
// ============================================================================

// Requirement declares what runtime environment is needed.
// This is used at prepare time to install/verify a bundle.
type Requirement struct {
	Name      string   `json:"name"`      // Bundle name
	Version   string   `json:"version"`   // Bundle version
	Python    string   `json:"python"`    // Python version requirement (e.g., ">=3.9")
	Packages  []string `json:"packages"`  // Required pip packages
	Commands  []string `json:"commands"`  // Required system commands
	GPU       bool     `json:"gpu"`       // GPU required
	MemoryMB  int      `json:"memoryMb"`  // Minimum memory
	DiskMB    int      `json:"diskMb"`    // Minimum disk space
}

// RequirementFromSpec creates a Requirement from a BundleSpec.
func RequirementFromSpec(spec *BundleSpec) *Requirement {
	return &Requirement{
		Name:     spec.Name,
		Version:  spec.Version,
		Python:   spec.Python,
		Packages: spec.Packages,
		Commands: spec.Commands,
		GPU:      !spec.GPUOptional,
	}
}

// EnvironmentReport contains the result of environment detection.
type EnvironmentReport struct {
	Ready         bool              `json:"ready"`
	PythonVersion string            `json:"pythonVersion,omitempty"`
	GPUAvailable  bool              `json:"gpuAvailable"`
	GPUNames      []string          `json:"gpuNames,omitempty"`
	Packages      map[string]string `json:"packages,omitempty"` // package → version
	Commands      map[string]bool   `json:"commands,omitempty"` // command → available
	Issues        []*Issue          `json:"issues,omitempty"`
	Warnings      []string          `json:"warnings,omitempty"`
}

// Issue represents an environment issue.
type Issue struct {
	Severity   Severity `json:"severity"`
	Message    string   `json:"message"`
	FixCommand string   `json:"fixCommand,omitempty"`
	AutoFix    bool     `json:"autoFix,omitempty"`
}

// Severity represents the severity of an issue.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Project represents a project to be executed.
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	RootPath    string `json:"rootPath"`    // Absolute path to project
	EntryPoint  string `json:"entryPoint"`  // Entry point file/command
	ProjectType string `json:"projectType"` // python, matlab, ros2, etc.
}

// RunConfig controls execution behavior.
type RunConfig struct {
	ProjectDir  string            `json:"projectDir"`  // Working directory
	EntryPoint  string            `json:"entryPoint"`  // Entry point relative to project dir
	Args        []string          `json:"args"`        // Command arguments
	Env         map[string]string `json:"env,omitempty"` // Environment variables
	Timeout     time.Duration     `json:"timeout"`     // Execution timeout
	LogCallback func(LogEntry)    `json:"-"`           // Real-time log callback
}

// RunResult contains the result of an execution.
type RunResult struct {
	RunID      string        `json:"runId"`
	Status     RunStatusEnum `json:"status"`
	ExitCode   int           `json:"exitCode"`
	Stdout     string        `json:"stdout,omitempty"`
	Stderr     string        `json:"stderr,omitempty"`
	Duration   time.Duration `json:"duration"`
	StartedAt  time.Time     `json:"startedAt"`
	CompletedAt time.Time    `json:"completedAt,omitempty"`
	Error      string        `json:"error,omitempty"`
}

// RunStatus represents the current status of a running execution.
type RunStatus struct {
	RunID      string        `json:"runId"`
	ProjectID  string        `json:"projectId"`
	Status     RunStatusEnum `json:"status"`
	Progress   float64       `json:"progress,omitempty"`
	StartedAt  time.Time     `json:"startedAt"`
	Duration   time.Duration `json:"duration"`
	LogCount   int           `json:"logCount"`
	Error      string        `json:"error,omitempty"`
}

// RunStatusEnum represents the status of a run.
type RunStatusEnum string

const (
	RunStatusPending   RunStatusEnum = "pending"
	RunStatusPreparing RunStatusEnum = "preparing"
	RunStatusRunning   RunStatusEnum = "running"
	RunStatusCompleted RunStatusEnum = "completed"
	RunStatusFailed    RunStatusEnum = "failed"
	RunStatusStopped   RunStatusEnum = "stopped"
	RunStatusTimeout   RunStatusEnum = "timeout"
)

// LogEntry represents a single log entry during execution.
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`   // DEBUG, INFO, WARN, ERROR
	Source    string    `json:"source"`  // stdout, stderr, system
	Message   string    `json:"message"`
	TaskID    string    `json:"taskId,omitempty"`
	RunID     string    `json:"runId,omitempty"`
	Raw       string    `json:"raw,omitempty"` // Original raw output
}

// ============================================================================
// Bundle Manager Interface
// ============================================================================

// BundleManager manages runtime bundles — shared, cached environments
// that can be installed once and reused across projects.
type BundleManager interface {
	// List returns all installed bundles.
	List() []*Bundle

	// Get returns a bundle by name.
	Get(name string) (*Bundle, bool)

	// Install installs a runtime bundle.
	// Progress is reported via the optional callback.
	Install(ctx context.Context, req *Requirement, progress ProgressCallback) (*Bundle, error)

	// InstallFromSpec installs a bundle from a BundleSpec.
	InstallFromSpec(ctx context.Context, spec *BundleSpec, progress ProgressCallback) (*Bundle, error)

	// Uninstall removes a runtime bundle.
	Uninstall(name string) error

	// CachePath returns the cache directory for bundles.
	CachePath() string

	// SharedBundles returns bundles that can be shared across projects.
	SharedBundles() []*Bundle

	// Clean removes all cached bundles.
	Clean(ctx context.Context) error

	// DetectInstalled checks which bundles are installed without installing.
	DetectInstalled(ctx context.Context, spec *BundleSpec) (*Bundle, bool)
}

// ProgressCallback is called during long-running operations (install, detect).
// message is a human-readable status, progress is 0.0–1.0.
type ProgressCallback func(message string, progress float64)

// Bundle is an immutable versioned runtime environment.
// Once created, its properties do not change.
type Bundle struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	PythonPath  string            `json:"pythonPath,omitempty"`
	Packages    []string          `json:"packages,omitempty"`
	Commands    []string          `json:"commands,omitempty"`
	EnvVars     map[string]string `json:"envVars,omitempty"`
	Path        string            `json:"path"` // Installation path in cache
	InstalledAt time.Time         `json:"installedAt"`
	SizeMB      int64             `json:"sizeMb"`
	GPUEnabled  bool              `json:"gpuEnabled"`
	Shared      bool              `json:"shared"` // true if cross-project shared
}

// IsInstalled returns true if the bundle has been installed.
func (b *Bundle) IsInstalled() bool {
	return b.Path != "" && !b.InstalledAt.IsZero()
}