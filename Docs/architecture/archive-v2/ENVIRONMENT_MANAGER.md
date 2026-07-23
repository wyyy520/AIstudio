# Environment Manager

## Overview

The Environment Manager handles **runtime environment detection, installation, and repair** for AIStudio. It ensures that the host system has all required tools, packages, and resources needed to compile and execute generated projects.

```
┌─────────────────────────────────────────────────────────────┐
│                 Environment Manager                         │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ Python       │  │ CUDA/GPU     │  │ Dependency       │  │
│  │ Detector     │  │ Detector     │  │ Detector         │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                  Status Cache                         │  │
│  │  (TTL: 30s, auto-invalidates on install/repair)      │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │
│  │ Installer    │  │ Repair       │  │ Log Collector    │  │
│  │ (pip, system)│  │ (auto-fix)   │  │ (install logs)   │  │
│  └──────────────┘  └──────────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Manager Interface

```go
type Manager struct {
    pythonDetector func() PythonInfo
    cudaDetector   func() CUDAInfo
    depDetector    func() []DependencyInfo
    installer      func(string) error

    cachedStatus  *EnvironmentStatus
    cacheExpiry   time.Time
    cacheTTL      time.Duration  // default: 30s
}
```

### Core API

```go
// GetStatus returns cached environment status (refreshes if expired)
func (m *Manager) GetStatus() interface{}

// InstallDependency installs a single dependency
func (m *Manager) InstallDependency(name string) error

// GetLogs returns installation/operation logs
func (m *Manager) GetLogs() []LogEntry

// InvalidateCache forces a fresh check on next GetStatus
func (m *Manager) InvalidateCache()
```

## Environment Status

```go
type EnvironmentStatus struct {
    Python       PythonInfo     `json:"python"`
    CUDA         CUDAInfo      `json:"cuda"`
    Dependencies []DependencyInfo `json:"dependencies"`
    Health       HealthStatus  `json:"health"`   // healthy, warning, critical
    CheckedAt    time.Time     `json:"checkedAt"`
}
```

## Detectors

### Python Detector

```go
type PythonInfo struct {
    Available   bool              `json:"available"`
    Version     string            `json:"version,omitempty"`
    Path        string            `json:"path,omitempty"`
    Packages    map[string]string `json:"packages,omitempty"` // package → version
}
```

Detection method:
- Run `python3 --version` (or `python --version` on Windows)
- Run `pip3 list --format=columns` (or `pip list`) for installed packages

### CUDA/GPU Detector

```go
type CUDAInfo struct {
    Available   bool     `json:"available"`
    Version     string   `json:"version,omitempty"`
    GPUNames    []string `json:"gpuNames,omitempty"`  // From nvidia-smi
    MemoryFree  int64    `json:"memoryFree,omitempty"` // In MB
    MemoryTotal int64    `json:"memoryTotal,omitempty"`
}
```

Detection method:
- Run `nvidia-smi` to check GPU availability
- Parse GPU names with `nvidia-smi --query-gpu=name --format=csv,noheader`
- CUDA version from `nvcc --version`

### Dependency Detector

```go
type DependencyInfo struct {
    Name    string `json:"name"`
    Version string `json:"version,omitempty"`
    Status  string `json:"status"` // "installed", "missing", "outdated"
    Command string `json:"command,omitempty"`
}
```

Checks for:
- System commands: `python3`, `pip3`, `docker`, `git`, `make`, `curl`, `wget`, `matlab`
- Required pip packages (from requirements/bundle specs)

## Health Status

```go
type HealthStatus string

const (
    HealthHealthy  HealthStatus = "healthy"   // All checks pass
    HealthWarning  HealthStatus = "warning"   // Non-critical issues (no GPU)
    HealthCritical HealthStatus = "critical"  // Critical issues (no Python)
)
```

Health is derived from collected issues:

```go
func HealthFromIssues(issues []Issue) HealthStatus {
    for _, issue := range issues {
        if issue.Severity == SeverityCritical {
            return HealthCritical
        }
    }
    for _, issue := range issues {
        if issue.Severity == SeverityError {
            return HealthCritical
        }
    }
    if len(issues) > 0 {
        return HealthWarning
    }
    return HealthHealthy
}
```

## Issue Structure

```go
type Issue struct {
    Code        string `json:"code"`        // e.g., "PYTHON_NOT_FOUND"
    Severity    string `json:"severity"`    // critical, error, warning
    Component   string `json:"component"`   // python, cuda, torch
    Title       string `json:"title"`       // Human-readable title
    Description string `json:"description,omitempty"`
    AutoFixable bool   `json:"autoFixable"`
    FixCommand  string `json:"fixCommand,omitempty"`
}
```

### Common Issues

| Code | Severity | Component | Auto-Fixable |
|------|----------|-----------|--------------|
| `PYTHON_NOT_FOUND` | critical | python | no |
| `CUDA_NOT_FOUND` | warning | cuda | no |
| `DEP_MISSING_<name>` | error | package | yes |

## Installation

```go
// InstallDependency installs a single dependency (pip package or system)
func (m *Manager) InstallDependency(name string) error {
    m.log.Info("install", "Installing dependency: "+name)
    err := m.installer(name)  // Calls the configured installer function
    if err != nil {
        m.log.Error("install", "Failed", err.Error())
        return err
    }
    m.InvalidateCache()
    return nil
}
```

The installer function delegates to:
- `pip install <package>` for Python packages
- System-specific package manager for system dependencies

## Cache

Status is cached with a 30-second TTL to avoid repeated system calls:

```go
func (m *Manager) GetStatus() interface{} {
    // Return cached if fresh
    if m.cachedStatus != nil && time.Now().Before(m.cacheExpiry) {
        return *m.cachedStatus
    }
    // Run full detection
    status := m.detectStatus()
    m.cachedStatus = &status
    m.cacheExpiry = time.Now().Add(m.cacheTTL)
    return status
}
```

Cache is invalidated after:
- Dependency installation
- Explicit `InvalidateCache()` call

## Runtime Bundle Lifecycle

The Environment Manager works with the Runtime's BundleManager for **progressive installation**:

1. **Detect** — Check what bundles are already installed in cache
2. **Install** — Install missing bundles (virtual env + pip packages)
3. **Share** — Bundles in central cache are shared across all projects
4. **Offline** — If a bundle is already cached, no network is needed

### Progressive Installation Flow

```
Request to run project requiring "yolo" bundle
    │
    ▼
Environment Manager: Detect
    │
    ├── Bundle installed? ──► Yes ──► Ready to execute
    │
    └── No ──► Bundle Manager: Install
        │
        ├── Check system commands (python3, nvidia-smi)
        ├── Create virtual environment (~/.aistudio/bundles/yolo-1.0.0/venv)
        ├── Install pip packages (torch, ultralytics, etc.)
        └── Mark as shared
            │
            ▼
        Ready to execute
```

## Offline Installation Support

Bundles in the cache (`~/.aistudio/bundles/`) can be pre-installed or transferred from another machine:

```
~/.aistudio/bundles/
├── yolo-1.0.0/
│   ├── venv/          ← Complete virtual environment
│   └── bundle.json
└── transformer-1.0.0/
    └── ...
```

The `DetectInstalled` method checks if the cache directory and virtual environment exist without requiring network access:

```go
func (m *BundleManager) DetectInstalled(ctx context.Context, spec *BundleSpec) (*Bundle, bool) {
    bundleDir := filepath.Join(m.cacheDir, spec.Name+"-"+spec.Version)
    venvPython := filepath.Join(bundleDir, "venv", "bin", "python")
    if _, err := os.Stat(venvPython); os.IsNotExist(err) {
        return nil, false  // Not installed
    }
    return bundle, true  // Found in cache
}
```

## GPU Detection

GPU detection uses `nvidia-smi`:

```go
func detectGPU(ctx context.Context, report *EnvironmentReport) {
    if err := exec.CommandContext(ctx, "nvidia-smi").Run(); err != nil {
        report.GPUAvailable = false
        return
    }
    report.GPUAvailable = true
    // Get GPU names
    cmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=name", "--format=csv,noheader")
    output, _ := cmd.Output()
    report.GPUNames = parseGPUOutput(string(output))
}
```

## Log Collector

```go
type LogCollector struct {
    Entries []LogEntry
}

type LogEntry struct {
    Level   string    `json:"level"`   // info, error, warn
    Op      string    `json:"op"`      // install, detect, repair
    Message string    `json:"message"`
    Detail  string    `json:"detail,omitempty"`
    Time    time.Time `json:"time"`
}
```

## API Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/environment/status` | Get cached environment status |
| POST | `/api/environment/check` | Run a fresh environment check |
| GET | `/api/environment/repair-plan` | Get suggested repair plan |
| POST | `/api/environment/repair` | Execute auto-repair |
| POST | `/api/environment/install` | Install a dependency |
| GET | `/api/environment/logs` | Get environment operation logs |
| DELETE | `/api/environment/logs` | Clear logs |
