package environment

type Manager struct {
	pythonDetector func() PythonInfo
	cudaDetector   func() CUDAInfo
	depDetector    func() []DependencyInfo
	installer      func(string) error
}

func NewManager() *Manager {
	return &Manager{
		pythonDetector: DetectPython,
		cudaDetector:   DetectCUDA,
		depDetector:    DetectDependencies,
		installer:      InstallDependency,
	}
}

func (m *Manager) CheckEnvironment() EnvironmentStatus {
	return EnvironmentStatus{
		Python:       m.pythonDetector(),
		CUDA:         m.cudaDetector(),
		Dependencies: m.depDetector(),
	}
}

func (m *Manager) InstallDependency(name string) error {
	return m.installer(name)
}

func (m *Manager) GetStatus() EnvironmentStatus {
	return m.CheckEnvironment()
}
