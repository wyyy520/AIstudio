package environment

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
)

func DetectPython(ctx context.Context) PythonInfo {
	info := PythonInfo{}

	pythonCmd := "python3"
	if runtime.GOOS == "windows" {
		pythonCmd = "python"
	}

	cmd := exec.CommandContext(ctx, pythonCmd, "--version")
	output, err := cmd.Output()
	if err != nil {
		return info
	}

	info.Available = true
	version := strings.TrimSpace(string(output))
	parts := strings.Split(version, " ")
	if len(parts) >= 2 {
		info.Version = strings.TrimPrefix(parts[1], "v")
	}

	pipCmd := "pip3"
	if runtime.GOOS == "windows" {
		pipCmd = "pip"
	}
	if err := exec.CommandContext(ctx, pipCmd, "--version").Run(); err == nil {
		info.Pip = true
	}

	return info
}

func DetectCUDA(ctx context.Context) CUDAInfo {
	info := CUDAInfo{}

	nvidiaCmd := exec.CommandContext(ctx, "nvidia-smi")
	if err := nvidiaCmd.Run(); err != nil {
		return info
	}

	info.Available = true

	versionCmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=driver_version", "--format=csv,noheader")
	if out, err := versionCmd.Output(); err == nil {
		info.Version = strings.TrimSpace(string(out))
	}

	gpuCmd := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=name,memory.total", "--format=csv,noheader")
	if out, err := gpuCmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			parts := strings.Split(line, ",")
			if len(parts) >= 2 {
				info.GPUs = append(info.GPUs, GPUInfo{
					Name:   strings.TrimSpace(parts[0]),
					Memory: strings.TrimSpace(parts[1]),
				})
			}
		}
	}

	return info
}

func DetectDocker(ctx context.Context) DockerInfo {
	info := DockerInfo{}

	if err := exec.CommandContext(ctx, "docker", "--version").Run(); err != nil {
		return info
	}

	info.Available = true
	if out, err := exec.CommandContext(ctx, "docker", "--version").Output(); err == nil {
		info.Version = strings.TrimSpace(string(out))
	}

	if err := exec.CommandContext(ctx, "docker", "compose", "version").Run(); err == nil {
		info.Compose = true
		if out, err := exec.CommandContext(ctx, "docker", "compose", "version").Output(); err == nil {
			info.ComposeVer = strings.TrimSpace(string(out))
		}
	}

	return info
}

func DetectGit(ctx context.Context) GitInfo {
	info := GitInfo{}

	if err := exec.CommandContext(ctx, "git", "--version").Run(); err != nil {
		return info
	}

	info.Available = true
	if out, err := exec.CommandContext(ctx, "git", "--version").Output(); err == nil {
		info.Version = strings.TrimSpace(string(out))
	}

	return info
}

func DetectGo(ctx context.Context) GoInfo {
	info := GoInfo{}

	if err := exec.CommandContext(ctx, "go", "version").Run(); err != nil {
		return info
	}

	info.Available = true
	if out, err := exec.CommandContext(ctx, "go", "version").Output(); err == nil {
		info.Version = strings.TrimSpace(string(out))
	}

	return info
}

func DetectOS() OSInfo {
	return OSInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

func DetectAll(ctx context.Context) *EnvironmentReport {
	report := NewEnvironmentReport()

	osInfo := DetectOS()
	report.OSInfo = &osInfo
	report.Warnings = append(report.Warnings, "OS: "+osInfo.OS+"/"+osInfo.Arch)

	pythonInfo := DetectPython(ctx)
	report.PythonInfo = &pythonInfo
	if pythonInfo.Available {
		report.PythonVersion = pythonInfo.Version
		report.Commands["python"] = true
	} else {
		report.Commands["python"] = false
	}

	cudaInfo := DetectCUDA(ctx)
	report.CUDAInfo = &cudaInfo
	if cudaInfo.Available {
		report.GPUAvailable = true
		for _, gpu := range cudaInfo.GPUs {
			report.GPUNames = append(report.GPUNames, gpu.Name)
		}
	}

	dockerInfo := DetectDocker(ctx)
	report.DockerInfo = &dockerInfo
	if dockerInfo.Available {
		report.Commands["docker"] = true
	} else {
		report.Commands["docker"] = false
	}

	gitInfo := DetectGit(ctx)
	report.GitInfo = &gitInfo
	if gitInfo.Available {
		report.Commands["git"] = true
	} else {
		report.Commands["git"] = false
	}

	goInfo := DetectGo(ctx)
	report.GoInfo = &goInfo
	if goInfo.Available {
		report.Commands["go"] = true
	} else {
		report.Commands["go"] = false
	}

	return report
}
