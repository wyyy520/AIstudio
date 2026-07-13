package environment

import "time"

// ---- Component Detection Results ----

// PythonInfo holds the result of Python detection.
type PythonInfo struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
	Path      string `json:"path"`
	Pip       bool   `json:"pip"`
}

// GPUInfo holds information about a single GPU.
type GPUInfo struct {
	Name   string `json:"name"`
	Memory string `json:"memory"`
}

// CUDAInfo holds the result of CUDA/GPU detection.
type CUDAInfo struct {
	Available bool      `json:"available"`
	Version   string    `json:"version"`
	GPUs      []GPUInfo `json:"gpus"`
}

// DependencyInfo holds the result of a single dependency check.
type DependencyInfo struct {
	Name      string `json:"name"`
	Required  string `json:"required"`
	Version   string `json:"version"`
	Status    string `json:"status"` // "installed", "missing", "version_mismatch"
}

// ---- Environment Status ----

// EnvironmentStatus is the top-level environment status.
type EnvironmentStatus struct {
	Health       string           `json:"health"`       // "healthy", "warning", "critical"
	Python       PythonInfo       `json:"python"`
	CUDA         CUDAInfo         `json:"cuda"`
	Dependencies []DependencyInfo `json:"dependencies"`
	CheckedAt    time.Time        `json:"checked_at"`
}

// ---- Check Results ----

// Severity indicates issue severity.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// Issue represents a single environment problem.
type Issue struct {
	Code        string   `json:"code"`
	Severity    Severity `json:"severity"`
	Component   string   `json:"component"`   // "python", "cuda", "pip", "torch", etc.
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Suggestion  string   `json:"suggestion"`
	AutoFixable bool     `json:"auto_fixable"`
}

// CheckResult is the result of a full environment check.
type CheckResult struct {
	Passed      bool      `json:"passed"`
	Health      string    `json:"health"`
	Issues      []Issue   `json:"issues"`
	Status      EnvironmentStatus `json:"status"`
	CheckedAt   time.Time `json:"checked_at"`
}

// ---- Repair ----

// RepairStep represents a single repair action.
type RepairStep struct {
	Step        int    `json:"step"`
	Action      string `json:"action"`       // e.g. "install_package", "reinstall_package", "create_venv"
	Description string `json:"description"`
	Command     string `json:"command"`      // human-readable command description
	Package     string `json:"package,omitempty"`
	Status      string `json:"status"`       // "pending", "running", "completed", "failed"
	Error       string `json:"error,omitempty"`
}

// RepairPlan contains the analysis and steps to fix the environment.
type RepairPlan struct {
	IssuesFound   int          `json:"issues_found"`
	IssuesFixed   int          `json:"issues_fixed"`
	AutoFixable   int          `json:"auto_fixable"`
	NeedsManual   int          `json:"needs_manual"`
	Steps         []RepairStep `json:"steps"`
	Analysis      string       `json:"analysis"`
}

// RepairResult is the result of executing a repair plan.
type RepairResult struct {
	Success      bool         `json:"success"`
	Plan         RepairPlan   `json:"plan"`
	StepsExecuted int         `json:"steps_executed"`
	StepsFailed  int          `json:"steps_failed"`
	Logs         []string     `json:"logs"`
	CompletedAt  time.Time    `json:"completed_at"`
}

// ---- Install ----

// InstallResult holds the result of a dependency installation.
type InstallResult struct {
	Success    bool     `json:"success"`
	Package    string   `json:"package"`
	Version    string   `json:"version,omitempty"`
	Output     string   `json:"output,omitempty"`
	Error      string   `json:"error,omitempty"`
	DurationMs int64    `json:"duration_ms"`
}

// ---- Environment Log Entry ----

// LogEntry represents a single environment operation log.
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`   // "info", "warn", "error"
	Operation string    `json:"operation"` // "check", "repair", "install", "detect"
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
}

// LogCollector collects environment operation logs.
type LogCollector struct {
	Entries []LogEntry `json:"entries"`
}

func (lc *LogCollector) Add(level, operation, message, details string) {
	lc.Entries = append(lc.Entries, LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Operation: operation,
		Message:   message,
		Details:   details,
	})
}

func (lc *LogCollector) Info(op, msg string) {
	lc.Add("info", op, msg, "")
}

func (lc *LogCollector) Warn(op, msg string) {
	lc.Add("warn", op, msg, "")
}

func (lc *LogCollector) Error(op, msg, details string) {
	lc.Add("error", op, msg, details)
}

// Health determines overall health from issues.
func HealthFromIssues(issues []Issue) string {
	hasError := false
	hasWarning := false
	for _, iss := range issues {
		if iss.Severity == SeverityCritical || iss.Severity == SeverityError {
			hasError = true
		}
		if iss.Severity == SeverityWarning {
			hasWarning = true
		}
	}
	if hasError {
		return "critical"
	}
	if hasWarning {
		return "warning"
	}
	return "healthy"
}