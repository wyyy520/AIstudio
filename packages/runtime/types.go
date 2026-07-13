package runtime

import (
	"time"

	env "github.com/aistudio/packages/environment"
)

type Requirement = env.Requirement
type EnvironmentReport = env.EnvironmentReport

type BundleSpec struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Python      string   `json:"python"`
	Packages    []string `json:"packages"`
	Commands    []string `json:"commands"`
	GPUOptional bool     `json:"gpu_optional"`
	Description string   `json:"description,omitempty"`
}

type Bundle struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	PythonPath  string            `json:"pythonPath,omitempty"`
	Packages    []string          `json:"packages,omitempty"`
	Commands    []string          `json:"commands,omitempty"`
	EnvVars     map[string]string `json:"envVars,omitempty"`
	Path        string            `json:"path"`
	InstalledAt time.Time         `json:"installedAt"`
	SizeMB      int64             `json:"sizeMb"`
	GPUEnabled  bool              `json:"gpuEnabled"`
	Shared      bool              `json:"shared"`
}

func (b *Bundle) IsInstalled() bool {
	return b.Path != "" && !b.InstalledAt.IsZero()
}

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	RootPath    string `json:"rootPath"`
	EntryPoint  string `json:"entryPoint"`
	ProjectType string `json:"projectType"`
}

type RunConfig struct {
	ProjectDir  string            `json:"projectDir"`
	EntryPoint  string            `json:"entryPoint"`
	Args        []string          `json:"args"`
	Env         map[string]string `json:"env,omitempty"`
	Timeout     time.Duration     `json:"timeout"`
	LogCallback func(LogEntry)    `json:"-"`
}

type RunResult struct {
	RunID       string        `json:"runId"`
	Status      RunStatusEnum `json:"status"`
	ExitCode    int           `json:"exitCode"`
	Stdout      string        `json:"stdout,omitempty"`
	Stderr      string        `json:"stderr,omitempty"`
	Duration    time.Duration `json:"duration"`
	StartedAt   time.Time     `json:"startedAt"`
	CompletedAt time.Time     `json:"completedAt,omitempty"`
	Error       string        `json:"error,omitempty"`
}

type RunStatus struct {
	RunID     string        `json:"runId"`
	ProjectID string        `json:"projectId"`
	Status    RunStatusEnum `json:"status"`
	Progress  float64       `json:"progress,omitempty"`
	StartedAt time.Time     `json:"startedAt"`
	Duration  time.Duration `json:"duration"`
	LogCount  int           `json:"logCount"`
	Error     string        `json:"error,omitempty"`
}

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

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Source    string    `json:"source"`
	Message   string    `json:"message"`
	TaskID    string    `json:"taskId,omitempty"`
	RunID     string    `json:"runId,omitempty"`
	Raw       string    `json:"raw,omitempty"`
}

type ProgressCallback func(message string, progress float64)

type ExecutorKind string

const (
	ExecutorLocal  ExecutorKind = "local"
	ExecutorDocker ExecutorKind = "docker"
	ExecutorSSH    ExecutorKind = "ssh"
)

type DockerConfig struct {
	Image        string
	Tag          string
	Volumes      []string
	Env          map[string]string
	Network      string
	WorkingDir   string
	RemoveOnExit bool
	GPUs         bool
}

type SSHConfig struct {
	Host       string
	Port       int
	KeyPath    string
	KnownHosts string
}