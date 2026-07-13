package service

import (
	"context"
	"log"

	"github.com/aistudio/backend/internal/environment"
	"github.com/aistudio/backend/internal/runtime"
	"github.com/aistudio/backend/internal/workflow"
)

// EnvironmentService provides environment management operations.
type EnvironmentService struct {
	manager     *environment.Manager
	integration *environment.EnvironmentIntegration
}

// NewEnvironmentService creates a new EnvironmentService from the container.
func NewEnvironmentService(c *Container) *EnvironmentService {
	mgr := environment.NewManager()
	return &EnvironmentService{
		manager:     mgr,
		integration: c.EnvIntegration,
	}
}

// GetStatus returns the current environment status.
func (s *EnvironmentService) GetStatus() environment.EnvironmentStatus {
	status := s.manager.GetStatus()
	if envStatus, ok := status.(environment.EnvironmentStatus); ok {
		return envStatus
	}
	log.Println("[env-service] GetStatus returned unexpected type")
	return environment.EnvironmentStatus{}
}

// Check runs a full environment check.
func (s *EnvironmentService) Check() interface{} {
	return s.manager.Check()
}

// GetRepairPlan returns a repair plan.
func (s *EnvironmentService) GetRepairPlan() interface{} {
	return nil
}

// Repair executes the repair plan.
func (s *EnvironmentService) Repair() interface{} {
	return nil
}

// InstallDependency installs a single dependency.
func (s *EnvironmentService) InstallDependency(name string) error {
	return s.manager.InstallDependency(name)
}

// GetLogs returns all environment operation logs.
func (s *EnvironmentService) GetLogs() []environment.LogEntry {
	return s.manager.GetLogs()
}

// ClearLogs clears all environment operation logs.
func (s *EnvironmentService) ClearLogs() {
	s.manager.ClearLogs()
}

// Prepare ensures the runtime environment is ready for the given workflow on the
// specified target platform. It delegates to EnvironmentIntegration to detect,
// install, and verify the runtime bundle.
func (s *EnvironmentService) Prepare(ctx context.Context, wf *workflow.Workflow, target workflow.Target) error {
	if s.integration == nil {
		log.Println("[env-service] EnvironmentIntegration not available, skipping Prepare")
		return nil
	}

	req := targetToRequirement(target, wf)
	return s.integration.PrepareRuntime(ctx, req, nil)
}

// targetToRequirement builds a runtime Requirement from a workflow target.
func targetToRequirement(target workflow.Target, wf *workflow.Workflow) *runtime.Requirement {
	req := &runtime.Requirement{
		Name:    string(target),
		Version: "1.0.0",
		GPU:     false,
	}

	switch target {
	case workflow.TargetPython:
		req.Python = ">=3.8"
		req.Packages = collectPythonPackages(wf)
	case workflow.TargetMATLAB:
		req.Commands = []string{"matlab"}
	case workflow.TargetROS2:
		req.Commands = []string{"ros2", "python3"}
		req.Packages = []string{"rclpy"}
	case workflow.TargetDocker:
		req.Commands = []string{"docker"}
	case workflow.TargetCPP:
		req.Commands = []string{"g++", "cmake"}
	case workflow.TargetSTM32:
		req.Commands = []string{"arm-none-eabi-gcc", "cmake"}
	}

	return req
}

// collectPythonPackages extracts unique pip package requirements from workflow nodes.
func collectPythonPackages(wf *workflow.Workflow) []string {
	seen := make(map[string]bool)
	var pkgs []string
	for _, node := range wf.Nodes {
		if node.Config != nil {
			if pkg, ok := node.Config["package"].(string); ok && !seen[pkg] {
				seen[pkg] = true
				pkgs = append(pkgs, pkg)
			}
			if deps, ok := node.Config["pip_packages"].([]interface{}); ok {
				for _, d := range deps {
					if pkg, ok := d.(string); ok && !seen[pkg] {
						seen[pkg] = true
						pkgs = append(pkgs, pkg)
					}
				}
			}
		}
	}
	return pkgs
}