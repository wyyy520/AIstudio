package environment

import (
	"context"
	"sync"
	"time"
)

type Manager struct {
	log        *LogCollector
	mu         sync.RWMutex
	cachedStatus *EnvironmentStatus
	cacheExpiry  time.Time
	cacheTTL     time.Duration
}

func NewManager() *Manager {
	return &Manager{
		log:      &LogCollector{},
		cacheTTL: 30 * time.Second,
	}
}

func (m *Manager) GetStatus(ctx context.Context) *EnvironmentStatus {
	m.mu.RLock()
	if m.cachedStatus != nil && time.Now().Before(m.cacheExpiry) {
		status := *m.cachedStatus
		m.mu.RUnlock()
		return &status
	}
	m.mu.RUnlock()

	status := m.detectStatus(ctx)

	m.mu.Lock()
	m.cachedStatus = status
	m.cacheExpiry = time.Now().Add(m.cacheTTL)
	m.mu.Unlock()

	return status
}

func (m *Manager) detectStatus(ctx context.Context) *EnvironmentStatus {
	status := &EnvironmentStatus{
		CheckedAt: time.Now(),
	}

	status.Python = DetectPython(ctx)
	status.CUDA = DetectCUDA(ctx)
	status.Docker = DetectDocker(ctx)
	status.Git = DetectGit(ctx)
	status.Go = DetectGo(ctx)
	status.OS = DetectOS()

	issues := m.collectIssues(*status)
	status.Health = HealthFromIssues(issues)

	return status
}

func (m *Manager) collectIssues(status EnvironmentStatus) []Issue {
	var issues []Issue

	if !status.Python.Available {
		issues = append(issues, Issue{
			Code:        "PYTHON_NOT_FOUND",
			Severity:    SeverityCritical,
			Component:   "python",
			Title:       "Python not found",
			Description: "Python is required to run AI models",
			Suggestion:  "Install Python 3.8 or later from https://python.org",
			AutoFixable: false,
		})
	}

	if !status.CUDA.Available {
		issues = append(issues, Issue{
			Code:        "CUDA_NOT_FOUND",
			Severity:    SeverityWarning,
			Component:   "cuda",
			Title:       "CUDA/GPU not detected",
			Description: "No NVIDIA GPU or CUDA driver found. CPU will be used.",
			Suggestion:  "Install NVIDIA driver and CUDA toolkit for GPU acceleration",
			AutoFixable: false,
		})
	}

	return issues
}

func (m *Manager) InvalidateCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cachedStatus = nil
}

func (m *Manager) GetLogs() []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.log.Entries
}

func (m *Manager) ClearLogs() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.log = &LogCollector{}
}

func (m *Manager) CheckEnvironment(ctx context.Context) *CheckResult {
	m.log.Info("check", "Starting environment check...")

	status := &EnvironmentStatus{CheckedAt: time.Now()}
	var issues []Issue

	m.log.Info("check", "Checking Python...")
	pythonInfo := DetectPython(ctx)
	status.Python = pythonInfo

	if !pythonInfo.Available {
		issues = append(issues, Issue{
			Code:        "PYTHON_NOT_FOUND",
			Severity:    SeverityCritical,
			Component:   "python",
			Title:       "Python is not installed",
			Description: "Python is required to run AI models. Please install Python 3.8 or later.",
			Suggestion:  "Download from https://python.org or use your package manager",
			AutoFixable: false,
		})
		m.log.Error("check", "Python not found", "")
	} else {
		m.log.Info("check", "Python found: version="+pythonInfo.Version+", path="+pythonInfo.Path)
	}

	m.log.Info("check", "Checking CUDA/GPU...")
	cudaInfo := DetectCUDA(ctx)
	status.CUDA = cudaInfo

	if !cudaInfo.Available {
		issues = append(issues, Issue{
			Code:        "CUDA_NOT_FOUND",
			Severity:    SeverityWarning,
			Component:   "cuda",
			Title:       "CUDA/GPU not detected",
			Description: "AI training will use CPU, which will be slower",
			Suggestion:  "Install NVIDIA driver and CUDA toolkit",
			AutoFixable: false,
		})
		m.log.Warn("check", "CUDA/GPU not available, will use CPU")
	} else {
		m.log.Info("check", "CUDA found: version="+cudaInfo.Version)
	}

	health := HealthFromIssues(issues)
	status.Health = health

	passed := len(issues) == 0 || (health == "healthy" || health == "warning")

	m.log.Info("check", "Environment check complete: health="+health+
		", issues="+itoa(len(issues)))

	return &CheckResult{
		Passed:    passed,
		Health:    health,
		Issues:    issues,
		Status:    *status,
		CheckedAt: time.Now(),
	}
}

func (m *Manager) QuickCheck(ctx context.Context) bool {
	python := DetectPython(ctx)
	return python.Available && python.Pip
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		digits = string(rune('0'+n%10)) + digits
		n /= 10
	}
	if neg {
		digits = "-" + digits
	}
	return digits
}
