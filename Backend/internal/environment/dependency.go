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

func parseRequirementsFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var packages []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name := strings.Split(line, "==")[0]
		name = strings.Split(name, ">=")[0]
		name = strings.Split(name, "<=")[0]
		name = strings.TrimSpace(name)
		if name != "" {
			packages = append(packages, name)
		}
	}
	return packages, scanner.Err()
}

func checkPackageInstalled(name string) (string, bool) {
	cmd := exec.Command(getPipCommand(), "show", name)
	output, err := cmd.Output()
	if err != nil {
		return "", false
	}
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "Version:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Version:")), true
		}
	}
	return "", false
}

func DetectDependencies() []DependencyInfo {
	reqFile := findRequirementsFile()
	if reqFile == "" {
		return nil
	}

	packages, err := parseRequirementsFile(reqFile)
	if err != nil || len(packages) == 0 {
		return nil
	}

	var deps []DependencyInfo
	for _, pkg := range packages {
		version, installed := checkPackageInstalled(pkg)
		status := "missing"
		if installed {
			status = "installed"
		}
		deps = append(deps, DependencyInfo{
			Name:    pkg,
			Version: version,
			Status:  status,
		})
	}
	return deps
}

func InstallDependency(name string) error {
	cmd := exec.Command(getPipCommand(), "install", name)
	return cmd.Run()
}
