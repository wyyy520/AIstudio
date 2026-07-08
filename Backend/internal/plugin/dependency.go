package plugin

import (
	"fmt"
	"log"
	"strings"
)

// Dependency defines a requirement for a plugin to function.
type Dependency struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	VersionMin     string `json:"version_min,omitempty"`
	Optional       bool   `json:"optional"`
	Description    string `json:"description,omitempty"`
}

// DependencyStatus represents the state of a dependency check.
type DependencyStatus string

const (
	DependencyStatusOK          DependencyStatus = "ok"
	DependencyStatusMissing     DependencyStatus = "missing"
	DependencyStatusVersionErr  DependencyStatus = "version_error"
	DependencyStatusNotRequired DependencyStatus = "not_required"
)

// DependencyCheckResult holds the result of checking a single dependency.
type DependencyCheckResult struct {
	Dependency Dependency       `json:"dependency"`
	Status     DependencyStatus `json:"status"`
	Message    string           `json:"message,omitempty"`
}

// DependencyManager handles dependency validation and resolution.
type DependencyManager struct {
	installedDeps map[string]string // name -> version of installed deps
}

// NewDependencyManager creates a new DependencyManager.
func NewDependencyManager() *DependencyManager {
	return &DependencyManager{
		installedDeps: make(map[string]string),
	}
}

// RegisterInstalled registers a dependency as already installed.
func (dm *DependencyManager) RegisterInstalled(name, version string) {
	dm.installedDeps[name] = version
}

// CheckDependencies validates all dependencies for a plugin.
// Returns a list of results and whether all required deps are satisfied.
func (dm *DependencyManager) CheckDependencies(deps []Dependency) ([]DependencyCheckResult, bool) {
	results := make([]DependencyCheckResult, 0, len(deps))
	allOK := true

	for _, dep := range deps {
		result := dm.checkSingle(dep)
		results = append(results, result)

		if result.Status != DependencyStatusOK && !dep.Optional {
			allOK = false
			log.Printf("[dependency] %s: %s - %s", dep.Name, result.Status, result.Message)
		}
	}

	return results, allOK
}

// checkSingle validates a single dependency.
func (dm *DependencyManager) checkSingle(dep Dependency) DependencyCheckResult {
	installedVersion, installed := dm.installedDeps[dep.Name]

	if !installed {
		if dep.Optional {
			return DependencyCheckResult{
				Dependency: dep,
				Status:     DependencyStatusNotRequired,
				Message:    fmt.Sprintf("optional dependency %s is not installed", dep.Name),
			}
		}
		return DependencyCheckResult{
			Dependency: dep,
			Status:     DependencyStatusMissing,
			Message:    fmt.Sprintf("dependency %s is not installed", dep.Name),
		}
	}

	// Version check
	if dep.VersionMin != "" {
		if !versionSatisfiesMin(installedVersion, dep.VersionMin) {
			return DependencyCheckResult{
				Dependency: dep,
				Status:     DependencyStatusVersionErr,
				Message:    fmt.Sprintf("dependency %s version %s < required %s", dep.Name, installedVersion, dep.VersionMin),
			}
		}
	}

	if dep.Version != "" {
		if installedVersion != dep.Version {
			return DependencyCheckResult{
				Dependency: dep,
				Status:     DependencyStatusVersionErr,
				Message:    fmt.Sprintf("dependency %s version %s != required %s", dep.Name, installedVersion, dep.Version),
			}
		}
	}

	return DependencyCheckResult{
		Dependency: dep,
		Status:     DependencyStatusOK,
		Message:    fmt.Sprintf("dependency %s (%s) is satisfied", dep.Name, installedVersion),
	}
}

// versionSatisfiesMin checks if installedVersion >= minVersion.
// Uses simple dot-separated version comparison.
func versionSatisfiesMin(installedVersion, minVersion string) bool {
	installedParts := splitVersion(installedVersion)
	minParts := splitVersion(minVersion)

	maxLen := len(installedParts)
	if len(minParts) > maxLen {
		maxLen = len(minParts)
	}

	for i := 0; i < maxLen; i++ {
		iv := 0
		mv := 0
		if i < len(installedParts) {
			iv = installedParts[i]
		}
		if i < len(minParts) {
			mv = minParts[i]
		}
		if iv > mv {
			return true
		}
		if iv < mv {
			return false
		}
	}
	return true
}

// splitVersion splits a version string into integer parts.
func splitVersion(version string) []int {
	parts := strings.Split(version, ".")
	result := make([]int, 0, len(parts))
	for _, p := range parts {
		v := 0
		fmt.Sscanf(p, "%d", &v)
		result = append(result, v)
	}
	return result
}

// InstallDependency simulates installing a dependency.
// In production, this would call pip install, apt-get, etc.
func (dm *DependencyManager) InstallDependency(dep Dependency) error {
	log.Printf("[dependency] installing %s@%s...", dep.Name, dep.Version)
	// Mock installation
	dm.RegisterInstalled(dep.Name, coalesce(dep.Version, dep.VersionMin, "latest"))
	log.Printf("[dependency] installed %s", dep.Name)
	return nil
}

func coalesce(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}