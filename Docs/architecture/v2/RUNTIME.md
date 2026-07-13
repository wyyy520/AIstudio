# Runtime Architecture

## Overview

The Runtime is the **unified execution engine** for AIStudio V2. It is a pure execution layer with zero knowledge of workflows, business logic, or algorithm logic. It only executes standard commands and manages runtime bundles.

```
                    ┌─────────────────────────────────────┐
                    │            Runtime                   │
                    │                                      │
                    │  ┌────────────────────────────────┐  │
                    │  │        Bundle Manager          │  │
                    │  │  Install │ Cache │ Share       │  │
                    │  └────────────────────────────────┘  │
                    │                                      │
                    │  ┌────────────────────────────────┐  │
                    │  │         Executor               │  │
                    │  │  ┌──────┐ ┌───────┐ ┌──────┐  │  │
                    │  │  │Local │ │Docker │ │ SSH  │  │  │
                    │  │  └──────┘ └───────┘ └──────┘  │  │
                    │  └────────────────────────────────┘  │
                    │                                      │
                    │  ┌────────────────────────────────┐  │
                    │  │    Environment Detector        │  │
                    │  │  Python │ CUDA │ Commands      │  │
                    │  └────────────────────────────────┘  │
                    └─────────────────────────────────────┘
```

## Runtime Interface

```go
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
```

## Bundle Manager

The BundleManager manages **runtime bundles** — shared, cached environments that can be installed once and reused across projects.

### Bundle Spec (bundle.json)

Bundles are declared in `packages/bundles/*/bundle.json`:

```json
{
  "name": "yolo",
  "version": "1.0.0",
  "python": ">=3.9",
  "packages": ["torch>=2.0", "ultralytics", "opencv-python"],
  "commands": ["python3"],
  "gpu_optional": false,
  "description": "YOLOv8 training and inference bundle"
}
```

### Available Bundles

| Bundle | Location | Description |
|--------|----------|-------------|
| `yolo` | `packages/bundles/yolo/bundle.json` | YOLOv8 training/inference |
| `transformer` | `packages/bundles/transformer/bundle.json` | Transformer models |
| `ros` | `packages/bundles/ros/bundle.json` | ROS2 dependencies |
| `stm32` | `packages/bundles/stm32/bundle.json` | STM32 toolchain |
| `matlab` | `packages/bundles/matlab/bundle.json` | MATLAB runtime |

### Bundle Manager Interface

```go
type BundleManager interface {
    List() []*Bundle
    Get(name string) (*Bundle, bool)
    Install(ctx context.Context, req *Requirement, progress ProgressCallback) (*Bundle, error)
    InstallFromSpec(ctx context.Context, spec *BundleSpec, progress ProgressCallback) (*Bundle, error)
    Uninstall(name string) error
    CachePath() string
    SharedBundles() []*Bundle
    Clean(ctx context.Context) error
    DetectInstalled(ctx context.Context, spec *BundleSpec) (*Bundle, bool)
}
```

### Bundle Structure

```go
type Bundle struct {
    Name        string            `json:"name"`
    Version     string            `json:"version"`
    PythonPath  string            `json:"pythonPath,omitempty"`
    Packages    []string          `json:"packages,omitempty"`
    Commands    []string          `json:"commands,omitempty"`
    EnvVars     map[string]string `json:"envVars,omitempty"`
    Path        string            `json:"path"`        // Installation path in cache
    InstalledAt time.Time         `json:"installedAt"`
    SizeMB      int64             `json:"sizeMb"`
    GPUEnabled  bool              `json:"gpuEnabled"`
    Shared      bool              `json:"shared"`      // Cross-project shared
}
```

### Bundle Cache

Bundles are cached at `~/.aistudio/bundles/<name>-<version>/`:

```
~/.aistudio/bundles/
├── yolo-1.0.0/
│   ├── venv/              ← Python virtual environment
│   │   ├── Scripts/       ← or bin/ on Linux/macOS
│   │   │   └── python.exe
│   │   └── Lib/           ← site-packages with installed packages
│   └── bundle.json        ← Bundle spec copy
└── transformer-1.0.0/
    └── ...
```

### Cross-Project Sharing

Bundles in the central cache are automatically **shared** across all projects. A bundle marked as `Shared: true` can be used by any project without reinstallation. The `SharedBundles()` method returns all shared bundles.

## Executor

The Executor runs commands with real-time log streaming. Three executor types:

### 1. Local Executor

Runs commands as subprocesses on the local machine:

```go
executor := runtime.NewLocalExecutor()
result := executor.Execute(ctx, RunConfig{
    ProjectDir: "/path/to/project",
    EntryPoint: "python",
    Args:       []string{"train.py", "--epochs", "100"},
    Env:        map[string]string{"CUDA_VISIBLE_DEVICES": "0"},
    Timeout:    3600 * time.Second,
    LogCallback: func(entry LogEntry) {
        fmt.Println(entry.Message)
    },
})
```

### 2. Docker Executor

Runs commands inside a Docker container:

```go
executor := runtime.NewDockerExecutor(DockerConfig{
    Image:        "python:3.11",
    Volumes:      []string{"/host/project:/workspace"},
    WorkingDir:   "/workspace",
    GPUs:         true,
    RemoveOnExit: true,
})
```

### 3. SSH Executor

Runs commands on a remote machine via SSH:

```go
executor := runtime.NewSSHExecutor(SSHConfig{
    Host:    "user@192.168.1.100",
    Port:    22,
    KeyPath: "~/.ssh/id_rsa",
})
```

### Executor Factory

```go
// Create executor by kind
executor := runtime.NewExecutor(runtime.ExecutorDocker, dockerConfig)

// Or directly
local := runtime.NewLocalExecutor()
docker := runtime.NewDockerExecutor(cfg)
ssh := runtime.NewSSHExecutor(cfg)
```

## Environment Detection

The `Runtime.Detect()` method produces an `EnvironmentReport`:

```go
type EnvironmentReport struct {
    Ready         bool              `json:"ready"`
    PythonVersion string            `json:"pythonVersion,omitempty"`
    GPUAvailable  bool              `json:"gpuAvailable"`
    GPUNames      []string          `json:"gpuNames,omitempty"`
    Packages      map[string]string `json:"packages,omitempty"`
    Commands      map[string]bool   `json:"commands,omitempty"`
    Issues        []*Issue          `json:"issues,omitempty"`
    Warnings      []string          `json:"warnings,omitempty"`
}
```

Detection checks:
- **OS**: runtime.GOOS / GOARCH
- **Python**: `python3 --version` / `pip list`
- **GPU**: `nvidia-smi` output
- **Commands**: docker, git, make, curl, wget, matlab
- **Dependencies**: Required pip packages from requirement

## Process Lifecycle

```
Execute(ctx, project, config)
    │
    ├── RunStatusPending
    │
    ├── RunStatusPreparing
    │   ├── Check environment
    │   └── Install bundle if needed
    │
    ├── RunStatusRunning
    │   ├── Spawn process
    │   ├── Stream stdout/stderr (real-time via LogCallback)
    │   └── Log level detection (INFO/WARN/ERROR)
    │
    ├── RunStatusCompleted    (exit code 0)
    ├── RunStatusFailed       (non-zero exit)
    ├── RunStatusStopped      (user requested Stop)
    └── RunStatusTimeout      (context deadline exceeded)
```

## Log Streaming

Real-time log capture uses `LogEntry`:

```go
type LogEntry struct {
    Timestamp time.Time `json:"timestamp"`
    Level     string    `json:"level"`   // DEBUG, INFO, WARN, ERROR
    Source    string    `json:"source"`  // stdout, stderr, system
    Message   string    `json:"message"`
    TaskID    string    `json:"taskId,omitempty"`
    RunID     string    `json:"runId,omitempty"`
    Raw       string    `json:"raw,omitempty"`
}
```

The `LogCallback` function receives entries in real-time for display in the UI.

## Events

The Runtime publishes events via the EventBus:

| Topic | When |
|-------|------|
| `runtime.started` | Execution started |
| `runtime.preparing` | Bundle installation |
| `runtime.running` | Process started |
| `runtime.completed` | Process completed |
| `runtime.failed` | Process failed |
| `runtime.stopped` | User stopped execution |
| `runtime.log` | Log entry produced |
| `runtime.progress` | Progress update |

## API Routes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/runtime/execute` | Execute a project |
| POST | `/api/v1/runtime/stop` | Stop a running execution |
| GET | `/api/v1/runtime/status/:runId` | Get execution status |
| GET | `/api/v1/runtime/running` | List running executions |
