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

func DetectPython() PythonInfo {
	info := PythonInfo{}

	cmd := exec.Command(getPythonCommand(), "--version")
	output, err := cmd.Output()
	if err == nil {
		version := strings.TrimSpace(string(output))
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
	}

	var pathCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		pathCmd = exec.Command("where", getPythonCommand())
	} else {
		pathCmd = exec.Command("which", getPythonCommand())
	}
	pathOutput, err := pathCmd.Output()
	if err == nil {
		info.Path = strings.TrimSpace(string(pathOutput))
	}

	pipCmd := exec.Command(getPipCommand(), "--version")
	err = pipCmd.Run()
	info.Pip = err == nil

	return info
}
