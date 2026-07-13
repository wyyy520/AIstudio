package environment

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// CreateVirtualEnv creates a Python virtual environment at the given path.
func (m *Manager) CreateVirtualEnv(venvPath string) error {
	m.log.Info("install", "Creating virtual environment: "+venvPath)

	cmd := exec.Command(getPythonCommand(), "-m", "venv", venvPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		m.log.Error("install", "Failed to create virtual environment", string(output))
		return fmt.Errorf("create venv: %w\n%s", err, string(output))
	}

	m.log.Info("install", "Virtual environment created at "+venvPath)
	return nil
}

// getVenvPip returns the pip path inside a virtual environment.
func getVenvPip(venvPath string) string {
	if venvPath == "" {
		return getPipCommand()
	}
	if _, err := os.Stat(filepath.Join(venvPath, "Scripts", "pip.exe")); err == nil {
		return filepath.Join(venvPath, "Scripts", "pip.exe")
	}
	return filepath.Join(venvPath, "bin", "pip")
}

// getVenvPython returns the python path inside a virtual environment.
func getVenvPython(venvPath string) string {
	if venvPath == "" {
		return getPythonCommand()
	}
	if _, err := os.Stat(filepath.Join(venvPath, "Scripts", "python.exe")); err == nil {
		return filepath.Join(venvPath, "Scripts", "python.exe")
	}
	return filepath.Join(venvPath, "bin", "python")
}

// InstallPackage installs a single package, optionally inside a venv.
func (m *Manager) InstallPackage(name, versionSpec, venvPath string) *InstallResult {
	start := time.Now()

	pkg := name
	if versionSpec != "" {
		pkg = name + versionSpec
	}

	m.log.Info("install", "Installing package: "+pkg)

	pipCmd := getVenvPip(venvPath)
	cmd := exec.Command(pipCmd, "install", pkg)
	output, err := cmd.CombinedOutput()

	duration := time.Since(start).Milliseconds()

	if err != nil {
		m.log.Error("install", "Failed to install "+pkg, string(output))
		return &InstallResult{
			Success:    false,
			Package:    name,
			Output:     string(output),
			Error:      err.Error(),
			DurationMs: duration,
		}
	}

	// Verify installation
	version, installed := checkPackageInstalled(name)
	if installed {
		m.log.Info("install", name+" installed successfully: version="+version)
	} else {
		m.log.Warn("install", name+" installed but version check failed")
	}

	return &InstallResult{
		Success:    true,
		Package:    name,
		Version:    version,
		Output:     string(output),
		DurationMs: duration,
	}
}

// InstallRequirements installs all packages from a requirements.txt file.
func (m *Manager) InstallRequirements(reqPath, venvPath string) ([]InstallResult, error) {
	m.log.Info("install", "Installing from requirements: "+reqPath)

	// Parse the requirements file
	required, err := parseRequirementsFile(reqPath)
	if err != nil {
		return nil, fmt.Errorf("parse requirements: %w", err)
	}
	if len(required) == 0 {
		m.log.Warn("install", "No packages found in requirements file")
		return nil, nil
	}

	// First try to install all at once (faster)
	pipCmd := getVenvPip(venvPath)
	cmd := exec.Command(pipCmd, "install", "-r", reqPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		m.log.Warn("install", "Batch install had errors, falling back to individual installs: "+string(output))
		// Fall back to individual installs
	}

	// Verify each package
	var results []InstallResult
	for _, req := range required {
		version, installed := checkPackageInstalled(req.Name)
		if installed {
			results = append(results, InstallResult{
				Success: true,
				Package: req.Name,
				Version: version,
			})
			m.log.Info("install", req.Name+"="+version+" (verified)")
		} else {
			// Try individual install
			result := m.InstallPackage(req.Name, req.Required, venvPath)
			results = append(results, *result)
		}
	}

	successCount := 0
	failCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		} else {
			failCount++
		}
	}
	m.log.Info("install", fmt.Sprintf("Installation complete: %d succeeded, %d failed",
		successCount, failCount))

	return results, nil
}

// VerifyInstallation verifies that all required packages are installed.
func (m *Manager) VerifyInstallation(reqPath string) ([]DependencyInfo, bool) {
	m.log.Info("install", "Verifying installation...")

	required, err := parseRequirementsFile(reqPath)
	if err != nil {
		m.log.Error("install", "Failed to parse requirements for verification", err.Error())
		return nil, false
	}

	var deps []DependencyInfo
	allInstalled := true

	for _, req := range required {
		version, installed := checkPackageInstalled(req.Name)
		status := "missing"
		if installed {
			status = "installed"
		} else {
			allInstalled = false
		}
		deps = append(deps, DependencyInfo{
			Name:     req.Name,
			Required: req.Required,
			Version:  version,
			Status:   status,
		})
	}

	if allInstalled {
		m.log.Info("install", "All packages verified: installed")
	} else {
		m.log.Warn("install", "Some packages are still missing")
	}

	return deps, allInstalled
}

// GetRequirementsPath returns the path to the requirements.txt file.
func GetRequirementsPath() string {
	return findRequirementsFile()
}

// EnsureRequirementsPath returns the absolute path to the Engine requirements.txt.
func EnsureRequirementsPath() string {
	// Try common locations
	candidates := []string{
		"../Engine/requirements.txt",
		"../../Engine/requirements.txt",
		"Engine/requirements.txt",
		"requirements.txt",
	}

	for _, rel := range candidates {
		abs, err := filepath.Abs(rel)
		if err == nil {
			if _, err := os.Stat(abs); err == nil {
				return abs
			}
		}
	}

	// Fallback: return the first resolved path
	for _, rel := range candidates {
		abs, err := filepath.Abs(rel)
		if err == nil {
			return abs
		}
	}
	return "requirements.txt"
}

// DepsToInstallList returns a list of package names that need to be installed.
func DepsToInstallList(deps []DependencyInfo) []string {
	var list []string
	for _, dep := range deps {
		if dep.Status == "missing" {
			if dep.Required != "" {
				list = append(list, dep.Name+dep.Required)
			} else {
				list = append(list, dep.Name)
			}
		}
	}
	return list
}

// JoinPackages joins package names for display.
func JoinPackages(pkgs []string) string {
	return strings.Join(pkgs, ", ")
}