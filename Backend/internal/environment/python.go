package environment

import (
	"os/exec"
	"runtime"
	"strings"
)

func getPythonCommand() string {
	if runtime.GOOS == "windows" {
		return "python"
	}
	return "python3"
}

func getPipCommand() string {
	if runtime.GOOS == "windows" {
		return "pip"
	}
	return "pip3"
}

// DetectPython runs a full Python environment detection.
func DetectPython() PythonInfo {
	info := PythonInfo{}

	cmd := exec.Command(getPythonCommand(), "--version")
	output, err := cmd.Output()
	if err != nil {
		return info
	}

	info.Available = true
	version := strings.TrimSpace(string(output))

	// "Python 3.11.5" -> "3.11"
	parts := strings.Split(version, " ")
	if len(parts) >= 2 {
		ver := strings.TrimSpace(parts[1])
		verParts := strings.Split(ver, ".")
		if len(verParts) >= 2 {
			info.Version = verParts[0] + "." + verParts[1]
		} else {
			info.Version = ver
		}
	}

	// Determine Python path
	var pathCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		pathCmd = exec.Command("where", getPythonCommand())
	} else {
		pathCmd = exec.Command("which", getPythonCommand())
	}
	pathOutput, err := pathCmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(pathOutput)), "\n")
		if len(lines) > 0 {
			info.Path = strings.TrimSpace(lines[0])
		}
	}

	// Check pip
	pipCmd := exec.Command(getPipCommand(), "--version")
	err = pipCmd.Run()
	info.Pip = err == nil

	return info
}

// DetectPythonVersion returns the full version string (e.g. "3.11.5").
func DetectPythonVersion() string {
	cmd := exec.Command(getPythonCommand(), "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	version := strings.TrimSpace(string(output))
	parts := strings.Split(version, " ")
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

// CheckPipWorking verifies pip can actually install packages.
func CheckPipWorking() bool {
	cmd := exec.Command(getPipCommand(), "list", "--format=columns")
	err := cmd.Run()
	return err == nil
}