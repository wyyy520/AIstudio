package environment

import (
	"time"
)

// CheckEnvironment runs a full environment check and returns a detailed result
// with issues, health status, and component status.
func (m *Manager) CheckEnvironment() *CheckResult {
	m.log.Info("check", "Starting environment check...")

	status := EnvironmentStatus{
		CheckedAt: time.Now(),
	}

	var issues []Issue

	// ---- 1. Check Python ----
	m.log.Info("check", "Checking Python...")
	pythonInfo := m.pythonDetector()
	status.Python = pythonInfo

	if !pythonInfo.Available {
		issues = append(issues, Issue{
			Code:        "PYTHON_NOT_FOUND",
			Severity:    SeverityCritical,
			Component:   "python",
			Title:       "Python is not installed",
			Description: "Python is required to run AI models. Please install Python 3.8 or later.",
			Suggestion:  "Download from https://python.org or use your package manager (apt install python3, brew install python3).",
			AutoFixable: false,
		})
		m.log.Error("check", "Python not found", "")
	} else {
		m.log.Info("check", "Python found: version="+pythonInfo.Version+", path="+pythonInfo.Path)

		// Check Python version meets minimum
		if pythonInfo.Version < "3.8" {
			issues = append(issues, Issue{
				Code:        "PYTHON_VERSION_TOO_OLD",
				Severity:    SeverityError,
				Component:   "python",
				Title:       "Python version is too old",
				Description: "Python " + pythonInfo.Version + " is below the minimum required version 3.8.",
				Suggestion:  "Upgrade Python to 3.8 or later.",
				AutoFixable: false,
			})
		}

		// Check pip
		if !pythonInfo.Pip {
			issues = append(issues, Issue{
				Code:        "PIP_NOT_FOUND",
				Severity:    SeverityError,
				Component:   "pip",
				Title:       "pip is not available",
				Description: "pip is required to install Python AI packages.",
				Suggestion:  "Install pip: python -m ensurepip --upgrade",
				AutoFixable: true,
			})
		} else {
			m.log.Info("check", "pip available")
		}
	}

	// ---- 2. Check CUDA / GPU ----
	m.log.Info("check", "Checking CUDA/GPU...")
	cudaInfo := m.cudaDetector()
	status.CUDA = cudaInfo

	if !cudaInfo.Available {
		issues = append(issues, Issue{
			Code:        "CUDA_NOT_FOUND",
			Severity:    SeverityWarning,
			Component:   "cuda",
			Title:       "CUDA/GPU not detected",
			Description: "No NVIDIA GPU or CUDA driver found. AI training will use CPU, which will be slower.",
			Suggestion:  "Install NVIDIA driver and CUDA toolkit for GPU acceleration.",
			AutoFixable: false,
		})
		m.log.Warn("check", "CUDA/GPU not available, will use CPU")
	} else {
		m.log.Info("check", "CUDA found: version="+cudaInfo.Version)
		for i, gpu := range cudaInfo.GPUs {
			m.log.Info("check", "GPU["+itoa(i)+"]: "+gpu.Name+" ("+gpu.Memory+")")
		}
	}

	// ---- 3. Check Dependencies ----
	m.log.Info("check", "Checking Python dependencies...")
	deps := m.depDetector()
	status.Dependencies = deps

	missingCount := 0
	for _, dep := range deps {
		if dep.Status == "missing" {
			missingCount++
			issues = append(issues, Issue{
				Code:        "DEP_MISSING_" + dep.Name,
				Severity:    SeverityError,
				Component:   dep.Name,
				Title:       "Package " + dep.Name + " is missing",
				Description: "The required package " + dep.Name + " is not installed.",
				Suggestion:  "Run: pip install " + dep.Name,
				AutoFixable: true,
			})
			m.log.Warn("check", "Missing: "+dep.Name)
		} else {
			m.log.Info("check", dep.Name+"="+dep.Version+" (installed)")
		}
	}

	if missingCount == 0 && len(deps) > 0 {
		m.log.Info("check", "All dependencies installed")
	}

	// ---- 4. Determine overall health ----
	health := HealthFromIssues(issues)
	status.Health = health

	passed := len(issues) == 0 || (health == "healthy" || health == "warning")

	m.log.Info("check", "Environment check complete: health="+health+
		", issues="+itoa(len(issues)))

	return &CheckResult{
		Passed:    passed,
		Health:    health,
		Issues:    issues,
		Status:    status,
		CheckedAt: time.Now(),
	}
}

// QuickCheck runs a fast check without scanning all dependencies.
// Returns true if the environment is minimally viable.
func (m *Manager) QuickCheck() bool {
	python := m.pythonDetector()
	return python.Available && python.Pip
}

// itoa is a simple int-to-string helper.
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