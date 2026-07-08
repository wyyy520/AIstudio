package environment

import (
	"log"
	"sync"
	"time"
)

// Manager is the central environment management component.
// It coordinates detection, installation, repair, and status reporting.
type Manager struct {
	pythonDetector func() PythonInfo
	cudaDetector   func() CUDAInfo
	depDetector    func() []DependencyInfo
	installer      func(string) error

	log *LogCollector
	mu  sync.RWMutex

	// Cached status for quick access
	cachedStatus  *EnvironmentStatus
	cacheExpiry   time.Time
	cacheTTL      time.Duration
}

// NewManager creates a new environment manager.
func NewManager() *Manager {
	return &Manager{
		pythonDetector: DetectPython,
		cudaDetector:   DetectCUDA,
		depDetector:    DetectDependencies,
		installer:      InstallDependency,
		log:            &LogCollector{},
		cacheTTL:       30 * time.Second,
	}
}

// ---- Core API ----

// GetStatus returns the current environment status (cached if fresh).
func (m *Manager) GetStatus() interface{} {
	m.mu.RLock()
	if m.cachedStatus != nil && time.Now().Before(m.cacheExpiry) {
		status := *m.cachedStatus
		m.mu.RUnlock()
		return status
	}
	m.mu.RUnlock()

	// Run full detection
	status := m.detectStatus()

	m.mu.Lock()
	m.cachedStatus = &status
	m.cacheExpiry = time.Now().Add(m.cacheTTL)
	m.mu.Unlock()

	return status
}

// InstallDependency installs a single dependency by name.
func (m *Manager) InstallDependency(name string) error {
	m.log.Info("install", "Installing dependency: "+name)
	err := m.installer(name)
	if err != nil {
		m.log.Error("install", "Failed to install "+name, err.Error())
		return err
	}
	m.log.Info("install", name+" installed successfully")
	// Invalidate cache
	m.invalidateCache()
	return nil
}

// GetLogs returns all collected environment operation logs.
func (m *Manager) GetLogs() []LogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.log.Entries
}

// ClearLogs clears the log collector.
func (m *Manager) ClearLogs() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.log = &LogCollector{}
}

// InvalidateCache forces a fresh status check on next GetStatus.
func (m *Manager) InvalidateCache() {
	m.invalidateCache()
}

// Check returns the detailed check result as interface{} for the workflow.EnvironmentChecker interface.
func (m *Manager) Check() interface{} {
	return m.CheckEnvironment()
}

// Status returns the environment status as interface{} for the workflow.EnvironmentChecker interface.
func (m *Manager) Status() interface{} {
	return m.GetStatus()
}

// ---- Internal Helpers ----

func (m *Manager) detectStatus() EnvironmentStatus {
	status := EnvironmentStatus{
		CheckedAt: time.Now(),
	}

	status.Python = m.pythonDetector()
	status.CUDA = m.cudaDetector()
	status.Dependencies = m.depDetector()

	// Determine health
	issues := m.collectIssues(status)
	status.Health = HealthFromIssues(issues)

	return status
}

func (m *Manager) collectIssues(status EnvironmentStatus) []Issue {
	var issues []Issue

	if !status.Python.Available {
		issues = append(issues, Issue{
			Code: "PYTHON_NOT_FOUND", Severity: SeverityCritical,
			Component: "python", Title: "Python not found",
			AutoFixable: false,
		})
	}
	if !status.CUDA.Available {
		issues = append(issues, Issue{
			Code: "CUDA_NOT_FOUND", Severity: SeverityWarning,
			Component: "cuda", Title: "CUDA/GPU not detected",
			AutoFixable: false,
		})
	}
	for _, dep := range status.Dependencies {
		if dep.Status == "missing" {
			issues = append(issues, Issue{
				Code: "DEP_MISSING_" + dep.Name, Severity: SeverityError,
				Component: dep.Name, Title: "Missing: " + dep.Name,
				AutoFixable: true,
			})
		}
	}

	return issues
}

func (m *Manager) invalidateCache() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cachedStatus = nil
}

// Logf logs a message to both the Go logger and the log collector.
func (m *Manager) Logf(level, op, format string, args ...interface{}) {
	msg := format
	if len(args) > 0 {
		msg = format
		for _, a := range args {
			msg = msg + " " + stringify(a)
		}
	}
	m.mu.Lock()
	m.log.Add(level, op, msg, "")
	m.mu.Unlock()
	log.Printf("[env][%s][%s] %s", level, op, msg)
}

func stringify(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case error:
		return val.Error()
	case int:
		return itoa(val)
	default:
		return ""
	}
}