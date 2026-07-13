package runtime

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// ============================================================================
// System Environment Detector
// ============================================================================

// NewEnvironmentReport creates an EnvironmentReport for the current system.
// It checks Python, GPU/CUDA, OS info, and required commands.
func NewEnvironmentReport(ctx context.Context) *EnvironmentReport {
	report := &EnvironmentReport{
		Packages: make(map[string]string),
		Commands: make(map[string]bool),
	}

	// Detect OS
	detectOS(report)

	// Detect Python
	detectPython(ctx, report)

	// Detect GPU/CUDA
	detectGPU(ctx, report)

	// Check common commands
	detectCommands(ctx, report)

	return report
}

// CheckRequirements checks if the current environment meets the requirements.
// Returns a report with issues if any requirements are not met.
func CheckRequirements(ctx context.Context, req *Requirement) *EnvironmentReport {
	report := NewEnvironmentReport(ctx)

	if req == nil {
		return report
	}

	// Check Python version
	if req.Python != "" && report.PythonVersion != "" {
		if !versionSatisfies(report.PythonVersion, req.Python) {
			report.Issues = append(report.Issues, &Issue{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("Python %s required, found %s", req.Python, report.PythonVersion),
				FixCommand: "Install Python " + req.Python,
			})
		}
	}

	// Check GPU requirement
	if req.GPU && !report.GPUAvailable {
		report.Issues = append(report.Issues, &Issue{
			Severity: SeverityError,
			Message:  "GPU required but not available",
		})
	}

	// Check required commands
	for _, cmd := range req.Commands {
		if !report.Commands[cmd] {
			report.Issues = append(report.Issues, &Issue{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("Required command not found: %s", cmd),
				FixCommand: fmt.Sprintf("Install %s using your package manager", cmd),
			})
		}
	}

	// Check required packages
	for _, pkg := range req.Packages {
		pkgName := extractPackageName(pkg)
		if _, found := report.Packages[pkgName]; !found {
			report.Issues = append(report.Issues, &Issue{
				Severity:   SeverityError,
				Message:    fmt.Sprintf("Required package not found: %s", pkgName),
				FixCommand: fmt.Sprintf("pip install %s", pkg),
			})
		}
	}

	report.Ready = len(report.Issues) == 0
	return report
}

// ============================================================================
// Internal Detection Functions
// ============================================================================

func detectOS(report *EnvironmentReport) {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	report.Warnings = append(report.Warnings, fmt.Sprintf("OS: %s/%s", osName, arch))
}

func detectPython(ctx context.Context, report *EnvironmentReport) {
	pythonCmd := "python3"
	if runtime.GOOS == "windows" {
		pythonCmd = "python"
	}

	// Check python exists
	cmd := exec.CommandContext(ctx, pythonCmd, "--version")
	output, err := cmd.Output()
	if err != nil {
		report.Commands[pythonCmd] = false
		return
	}
	report.Commands[pythonCmd] = true

	version := strings.TrimSpace(string(output))
	parts := strings.Split(version, " ")
	if len(parts) >= 2 {
		report.PythonVersion = strings.TrimPrefix(parts[1], "v")
	}

	// Check pip
	pipCmd := "pip3"
	if runtime.GOOS == "windows" {
		pipCmd = "pip"
	}
	if err := exec.CommandContext(ctx, pipCmd, "--version").Run(); err == nil {
		report.Commands[pipCmd] = true
		// List installed packages
		listPackages(ctx, pipCmd, report)
	} else {
		report.Commands[pipCmd] = false
	}
}

func detectGPU(ctx context.Context, report *EnvironmentReport) {
	// Check nvidia-smi
	if err := exec.CommandContext(ctx, "nvidia-smi").Run(); err != nil {
		report.GPUAvailable = false
		return
	}

	report.GPUAvailable = true
	report.Commands["nvidia-smi"] = true

	// Get GPU names
	cmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=name", "--format=csv,noheader")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if name := strings.TrimSpace(line); name != "" {
				report.GPUNames = append(report.GPUNames, name)
			}
		}
	}
}

func detectCommands(ctx context.Context, report *EnvironmentReport) {
	commonCommands := []string{"docker", "git", "make", "curl", "wget", "matlab"}
	for _, cmdName := range commonCommands {
		checkCmd := "which"
		if runtime.GOOS == "windows" {
			checkCmd = "where"
		}
		err := exec.CommandContext(ctx, checkCmd, cmdName).Run()
		report.Commands[cmdName] = err == nil
	}
}

func listPackages(ctx context.Context, pipCmd string, report *EnvironmentReport) {
	cmd := exec.CommandContext(ctx, pipCmd, "list", "--format=columns")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Package") || strings.HasPrefix(line, "-") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			report.Packages[fields[0]] = fields[1]
		}
	}
}

// ============================================================================
// Version Helpers
// ============================================================================

// versionSatisfies checks if version satisfies a constraint like ">=3.9".
// This is a simplified parser; production would use a semver library.
func versionSatisfies(version, constraint string) bool {
	version = strings.TrimSpace(version)
	constraint = strings.TrimSpace(constraint)

	if constraint == "" {
		return true
	}

	// Parse operator and required version
	op := ""
	required := constraint
	for _, possibleOp := range []string{">=", "<=", "==", "!=", ">", "<", "~="} {
		if strings.HasPrefix(constraint, possibleOp) {
			op = possibleOp
			required = strings.TrimPrefix(constraint, op)
			break
		}
	}

	if op == "" {
		return version == constraint
	}

	cmp := compareVersions(version, required)
	switch op {
	case ">=":
		return cmp >= 0
	case "<=":
		return cmp <= 0
	case "==":
		return cmp == 0
	case "!=":
		return cmp != 0
	case ">":
		return cmp > 0
	case "<":
		return cmp < 0
	case "~=":
		// Compatible release: major.minor must match
		vParts := strings.Split(version, ".")
		rParts := strings.Split(required, ".")
		if len(vParts) >= 2 && len(rParts) >= 2 {
			return vParts[0] == rParts[0] && vParts[1] >= rParts[1]
		}
		return cmp >= 0
	}
	return true
}

func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var aInt, bInt int
		if i < len(aParts) {
			fmt.Sscanf(aParts[i], "%d", &aInt)
		}
		if i < len(bParts) {
			fmt.Sscanf(bParts[i], "%d", &bInt)
		}
		if aInt < bInt {
			return -1
		}
		if aInt > bInt {
			return 1
		}
	}
	return 0
}

func extractPackageName(pkgSpec string) string {
	// Strip version specifiers: torch>=2.0 -> torch
	for _, sep := range []string{">=", "<=", "==", "!=", ">", "<", "~="} {
		if idx := strings.Index(pkgSpec, sep); idx >= 0 {
			return pkgSpec[:idx]
		}
	}
	return pkgSpec
}
