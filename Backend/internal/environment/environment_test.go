package environment

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPython(t *testing.T) {
	info := DetectPython()

	if info.Version == "" {
		t.Log("Python not found in test environment (this may be expected)")
	} else {
		t.Logf("Python version: %s", info.Version)
		t.Logf("Python path: %s", info.Path)
		t.Logf("Pip available: %v", info.Pip)
	}
}

func TestDetectCUDA(t *testing.T) {
	info := DetectCUDA()

	if info.CUDA == "" && len(info.GPUs) == 0 {
		t.Log("CUDA/GPU not detected (this may be expected)")
	} else {
		t.Logf("CUDA version: %s", info.CUDA)
		t.Logf("GPU count: %d", len(info.GPUs))
		for i, gpu := range info.GPUs {
			t.Logf("  GPU %d: %s (%s)", i, gpu.Name, gpu.Memory)
		}
	}
}

func TestParseRequirementsFile(t *testing.T) {
	tmpDir := t.TempDir()
	reqFile := filepath.Join(tmpDir, "requirements.txt")

	content := `# AI Studio Engine
torch>=2.1.0
transformers==4.36.0
numpy>=1.24.0
opencv-python>=4.8.0

# empty line above
`
	if err := os.WriteFile(reqFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write requirements file: %v", err)
	}

	packages, err := parseRequirementsFile(reqFile)
	if err != nil {
		t.Fatalf("parseRequirementsFile failed: %v", err)
	}

	expected := []string{"torch", "transformers", "numpy", "opencv-python"}
	if len(packages) != len(expected) {
		t.Fatalf("expected %d packages, got %d: %v", len(expected), len(packages), packages)
	}
	for i, pkg := range expected {
		if packages[i] != pkg {
			t.Errorf("expected %q at index %d, got %q", pkg, i, packages[i])
		}
	}
}

func TestParseRequirementsFile_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	reqFile := filepath.Join(tmpDir, "empty.txt")

	if err := os.WriteFile(reqFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	packages, err := parseRequirementsFile(reqFile)
	if err != nil {
		t.Fatalf("parseRequirementsFile failed: %v", err)
	}
	if len(packages) != 0 {
		t.Errorf("expected 0 packages for empty file, got %d", len(packages))
	}
}

func TestParseRequirementsFile_CommentsOnly(t *testing.T) {
	tmpDir := t.TempDir()
	reqFile := filepath.Join(tmpDir, "comments.txt")

	content := `# just a comment
# another comment
`
	if err := os.WriteFile(reqFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	packages, err := parseRequirementsFile(reqFile)
	if err != nil {
		t.Fatalf("parseRequirementsFile failed: %v", err)
	}
	if len(packages) != 0 {
		t.Errorf("expected 0 packages for comments-only file, got %d", len(packages))
	}
}

func TestFindRequirementsFile(t *testing.T) {
	path := findRequirementsFile()
	t.Logf("Requirements file found: %q", path)
}

func TestManagerWithMockDetectors(t *testing.T) {
	mgr := NewManager()

	mgr.pythonDetector = func() PythonInfo {
		return PythonInfo{Version: "3.12", Path: "/usr/bin/python3", Pip: true}
	}
	mgr.cudaDetector = func() CUDAInfo {
		return CUDAInfo{
			CUDA: "12.1",
			GPUs: []GPUInfo{
				{Name: "NVIDIA RTX 4090", Memory: "24564 MiB"},
			},
		}
	}
	mgr.depDetector = func() []DependencyInfo {
		return []DependencyInfo{
			{Name: "torch", Version: "2.1.0", Status: "installed"},
			{Name: "transformers", Version: "", Status: "missing"},
		}
	}

	status := mgr.CheckEnvironment()

	if status.Python.Version != "3.12" {
		t.Errorf("expected Python 3.12, got %s", status.Python.Version)
	}
	if !status.Python.Pip {
		t.Errorf("expected pip to be true")
	}
	if status.CUDA.CUDA != "12.1" {
		t.Errorf("expected CUDA 12.1, got %s", status.CUDA.CUDA)
	}
	if len(status.CUDA.GPUs) != 1 {
		t.Errorf("expected 1 GPU, got %d", len(status.CUDA.GPUs))
	}
	if len(status.Dependencies) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(status.Dependencies))
	}
	if status.Dependencies[0].Status != "installed" {
		t.Errorf("expected torch to be installed, got %s", status.Dependencies[0].Status)
	}
	if status.Dependencies[1].Status != "missing" {
		t.Errorf("expected transformers to be missing, got %s", status.Dependencies[1].Status)
	}
}

func TestManagerGetStatus(t *testing.T) {
	mgr := NewManager()

	mgr.pythonDetector = func() PythonInfo {
		return PythonInfo{Version: "3.11", Path: "", Pip: false}
	}
	mgr.cudaDetector = func() CUDAInfo {
		return CUDAInfo{}
	}
	mgr.depDetector = func() []DependencyInfo {
		return nil
	}

	status := mgr.GetStatus()

	if status.Python.Version != "3.11" {
		t.Errorf("expected Python 3.11, got %s", status.Python.Version)
	}
}

func TestManagerInstallDependency(t *testing.T) {
	installed := ""
	mgr := NewManager()
	mgr.installer = func(name string) error {
		installed = name
		return nil
	}

	if err := mgr.InstallDependency("torch"); err != nil {
		t.Fatalf("InstallDependency failed: %v", err)
	}
	if installed != "torch" {
		t.Errorf("expected 'torch' to be installed, got %q", installed)
	}
}
