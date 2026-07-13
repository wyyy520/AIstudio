package environment

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var defaultRequirementsFiles = []string{
	"requirements.txt",
	"../requirements.txt",
	"../../requirements.txt",
	"../Engine/requirements.txt",
}

func findRequirementsFile() string {
	for _, relPath := range defaultRequirementsFiles {
		absPath, err := filepath.Abs(relPath)
		if err == nil {
			if _, err := os.Stat(absPath); err == nil {
				return absPath
			}
		}
	}
	return ""
}

// parseRequirementsFile reads a requirements.txt and returns package names with version specs.
func parseRequirementsFile(path string) ([]DependencyInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var deps []DependencyInfo
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse: torch>=2.1.0  -> name=torch, required=>=2.1.0
		name, required := parsePackageLine(line)
		if name != "" {
			deps = append(deps, DependencyInfo{
				Name:     name,
				Required: required,
				Status:   "missing",
			})
		}
	}
	return deps, scanner.Err()
}

// parsePackageLine splits a requirements line into name and version spec.
func parsePackageLine(line string) (name, required string) {
	name = line
	required = ""

	// Handle version specifiers
	for _, sep := range []string{">=", "<=", "==", "!=", ">", "<", "~="} {
		if idx := strings.Index(name, sep); idx >= 0 {
			required = name[idx:]
			name = name[:idx]
			break
		}
	}

	// Handle extras: "package[extra]>=1.0"
	if idx := strings.Index(name, "["); idx >= 0 {
		name = name[:idx]
	}

	name = strings.TrimSpace(name)
	required = strings.TrimSpace(required)
	return
}

// checkPackageInstalled returns the installed version and whether it's installed.
func checkPackageInstalled(name string) (version string, installed bool) {
	cmd := exec.Command(getPipCommand(), "show", name)
	output, err := cmd.Output()
	if err != nil {
		return "", false
	}
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Version:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Version:")), true
		}
	}
	return "", false
}

// DetectDependencies scans requirements.txt and checks each package.
func DetectDependencies() []DependencyInfo {
	reqFile := findRequirementsFile()
	if reqFile == "" {
		return nil
	}

	required, err := parseRequirementsFile(reqFile)
	if err != nil || len(required) == 0 {
		return nil
	}

	var deps []DependencyInfo
	for _, req := range required {
		version, installed := checkPackageInstalled(req.Name)
		status := "missing"
		if installed {
			status = "installed"
		}
		deps = append(deps, DependencyInfo{
			Name:     req.Name,
			Required: req.Required,
			Version:  version,
			Status:   status,
		})
	}
	return deps
}

// InstallDependency installs a single package via pip.
func InstallDependency(name string) error {
	cmd := exec.Command(getPipCommand(), "install", name)
	return cmd.Run()
}

// InstallDependencyWithVersion installs a package with a specific version spec.
func InstallDependencyWithVersion(name, versionSpec string) error {
	if versionSpec != "" {
		name = name + versionSpec
	}
	cmd := exec.Command(getPipCommand(), "install", name)
	return cmd.Run()
}

// InstallRequirementsFromFile installs all packages from a requirements.txt file.
func InstallRequirementsFromFile(path string) (string, error) {
	cmd := exec.Command(getPipCommand(), "install", "-r", path)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// GetPipList returns all installed packages as a map of name -> version.
func GetPipList() map[string]string {
	cmd := exec.Command(getPipCommand(), "list", "--format=columns")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	result := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Package") || strings.HasPrefix(line, "-") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			result[fields[0]] = fields[1]
		}
	}
	return result
}