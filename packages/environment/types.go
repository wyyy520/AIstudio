package environment

import "time"

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

type PythonInfo struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
	Path      string `json:"path"`
	Pip       bool   `json:"pip"`
	VenvPath  string `json:"venvPath,omitempty"`
}

type GPUInfo struct {
	Name   string `json:"name"`
	Memory string `json:"memory"`
}

type CUDAInfo struct {
	Available bool      `json:"available"`
	Version   string    `json:"version"`
	GPUs      []GPUInfo `json:"gpus"`
}

type DockerInfo struct {
	Available    bool   `json:"available"`
	Version      string `json:"version"`
	Compose      bool   `json:"compose"`
	ComposeVer   string `json:"composeVersion,omitempty"`
}

type GitInfo struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
}

type GoInfo struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
}

type OSInfo struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

type DependencyInfo struct {
	Name     string `json:"name"`
	Required string `json:"required"`
	Version  string `json:"version"`
	Status   string `json:"status"`
}

type EnvironmentStatus struct {
	Health       string           `json:"health"`
	Python       PythonInfo       `json:"python"`
	CUDA         CUDAInfo         `json:"cuda"`
	Docker       DockerInfo       `json:"docker"`
	Git          GitInfo          `json:"git"`
	Go           GoInfo           `json:"go"`
	OS           OSInfo           `json:"os"`
	Dependencies []DependencyInfo `json:"dependencies"`
	CheckedAt    time.Time        `json:"checked_at"`
}

type Issue struct {
	Code        string   `json:"code"`
	Severity    Severity `json:"severity"`
	Component   string   `json:"component"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Suggestion  string   `json:"suggestion"`
	AutoFixable bool     `json:"auto_fixable"`
}

type CheckResult struct {
	Passed    bool              `json:"passed"`
	Health    string            `json:"health"`
	Issues    []Issue           `json:"issues"`
	Status    EnvironmentStatus `json:"status"`
	CheckedAt time.Time         `json:"checked_at"`
}

type InstallResult struct {
	Success    bool   `json:"success"`
	Package    string `json:"package"`
	Version    string `json:"version,omitempty"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Operation string    `json:"operation"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
}

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

type RepairStep struct {
	Step        int    `json:"step"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Command     string `json:"command"`
	Package     string `json:"package,omitempty"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
}

type RepairPlan struct {
	IssuesFound int          `json:"issues_found"`
	IssuesFixed int          `json:"issues_fixed"`
	AutoFixable int          `json:"auto_fixable"`
	NeedsManual int          `json:"needs_manual"`
	Steps       []RepairStep `json:"steps"`
	Analysis    string       `json:"analysis"`
}

type RepairResult struct {
	Success       bool       `json:"success"`
	Plan          RepairPlan `json:"plan"`
	StepsExecuted int        `json:"steps_executed"`
	StepsFailed   int        `json:"steps_failed"`
	Logs          []string   `json:"logs"`
	CompletedAt   time.Time  `json:"completed_at"`
}

type Requirement struct {
	Name     string   `json:"name"`
	Version  string   `json:"version"`
	Python   string   `json:"python"`
	Packages []string `json:"packages"`
	Commands []string `json:"commands"`
	GPU      bool     `json:"gpu"`
	MemoryMB int      `json:"memoryMb"`
	DiskMB   int      `json:"diskMb"`
}

type EnvironmentReport struct {
	Ready         bool              `json:"ready"`
	PythonVersion string            `json:"pythonVersion,omitempty"`
	GPUAvailable  bool              `json:"gpuAvailable"`
	GPUNames      []string          `json:"gpuNames,omitempty"`
	Packages      map[string]string `json:"packages,omitempty"`
	Commands      map[string]bool   `json:"commands,omitempty"`
	Issues        []*Issue          `json:"issues,omitempty"`
	Warnings      []string          `json:"warnings,omitempty"`
	PythonInfo    *PythonInfo       `json:"pythonInfo,omitempty"`
	CUDAInfo      *CUDAInfo         `json:"cudaInfo,omitempty"`
	DockerInfo    *DockerInfo       `json:"dockerInfo,omitempty"`
	GitInfo       *GitInfo          `json:"gitInfo,omitempty"`
	GoInfo        *GoInfo           `json:"goInfo,omitempty"`
	OSInfo        *OSInfo           `json:"osInfo,omitempty"`
}

func NewEnvironmentReport() *EnvironmentReport {
	return &EnvironmentReport{
		Packages: make(map[string]string),
		Commands: make(map[string]bool),
	}
}

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
