package environment

import (
	"os/exec"
	"regexp"
	"strings"
)

func nvidiaSmiAvailable() bool {
	err := exec.Command("nvidia-smi").Run()
	return err == nil
}

func getCUDAToolkitVersion() string {
	cmd := exec.Command("nvidia-smi")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	re := regexp.MustCompile(`CUDA Version:\s*(\d+\.\d+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// DetectCUDA runs a full CUDA/GPU detection.
func DetectCUDA() CUDAInfo {
	info := CUDAInfo{
		GPUs: []GPUInfo{},
	}

	if !nvidiaSmiAvailable() {
		return info
	}

	info.Available = true
	info.Version = getCUDAToolkitVersion()

	// Query GPU details
	gpuCmd := exec.Command("nvidia-smi", "--query-gpu=name,memory.total", "--format=csv,noheader")
	gpuOutput, err := gpuCmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(gpuOutput)), "\n")
		for _, line := range lines {
			parts := strings.Split(line, ", ")
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

// DetectGPUCount returns the number of available GPUs.
func DetectGPUCount() int {
	info := DetectCUDA()
	return len(info.GPUs)
}

// DetectGPUMemory returns the total GPU memory in MiB as a string.
func DetectGPUMemory(index int) string {
	info := DetectCUDA()
	if index < len(info.GPUs) {
		return info.GPUs[index].Memory
	}
	return ""
}

// HasCUDA checks if CUDA is available at all.
func HasCUDA() bool {
	return nvidiaSmiAvailable()
}