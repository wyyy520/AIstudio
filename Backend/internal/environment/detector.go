package environment

type PythonInfo struct {
	Version string `json:"version"`
	Path    string `json:"path"`
	Pip     bool   `json:"pip"`
}

type GPUInfo struct {
	Name   string `json:"name"`
	Memory string `json:"memory"`
}

type CUDAInfo struct {
	CUDA string     `json:"cuda"`
	GPUs []GPUInfo  `json:"gpu"`
}

type DependencyInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

type EnvironmentStatus struct {
	Python       PythonInfo       `json:"python"`
	CUDA         CUDAInfo         `json:"cuda"`
	Dependencies []DependencyInfo `json:"dependencies"`
}
