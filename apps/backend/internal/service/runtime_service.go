package service

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aistudio/backend/internal/compiler"
	"github.com/aistudio/backend/internal/project"
	"github.com/aistudio/backend/internal/runtime"
	"github.com/aistudio/backend/internal/workflow"
)

// RuntimeService provides runtime execution operations.
type RuntimeService struct {
	runtime       runtime.Runtime
	bundleManager runtime.BundleManager
	executor      runtime.CommandExecutor
	compiler      compiler.Compiler
	projectMgr    *project.Manager
}

// NewRuntimeService creates a new RuntimeService.
func NewRuntimeService(rt runtime.Runtime, bm runtime.BundleManager, exec runtime.CommandExecutor, comp compiler.Compiler, proj *project.Manager) *RuntimeService {
	return &RuntimeService{
		runtime:       rt,
		bundleManager: bm,
		executor:      exec,
		compiler:      comp,
		projectMgr:    proj,
	}
}

// Plan creates a compilation plan without generating files.
func (s *RuntimeService) Plan(ctx context.Context, projectID string, target workflow.Target, projectName string) (*compiler.CompilePlan, error) {
	if s.compiler == nil {
		return nil, fmt.Errorf("compiler not initialized")
	}

	workflowPath := s.projectMgr.GetWorkflowPath(projectID)

	opts := compiler.CompileOptions{
		Target:      target,
		ProjectName: projectName,
		DryRun:      true,
	}

	return s.compiler.Plan(ctx, workflowPath, opts)
}

// Compile generates project files from a workflow.
func (s *RuntimeService) Compile(ctx context.Context, projectID string, target workflow.Target, projectName string) (*compiler.CompileResult, error) {
	if s.compiler == nil {
		return nil, fmt.Errorf("compiler not initialized")
	}

	workflowPath := s.projectMgr.GetWorkflowPath(projectID)

	outputDir := filepath.Join("projects", projectID, "output")
	if projectName != "" {
		outputDir = filepath.Join(outputDir, projectName)
	}

	opts := compiler.CompileOptions{
		OutputDir:   outputDir,
		Target:      target,
		ProjectName: projectName,
	}

	return s.compiler.Compile(ctx, workflowPath, opts)
}

// Detect checks if the environment meets the runtime requirements.
func (s *RuntimeService) Detect(ctx context.Context, req *runtime.Requirement) (*runtime.EnvironmentReport, error) {
	return s.runtime.Detect(ctx, req)
}

// Prepare installs the runtime bundle and prepares the environment.
func (s *RuntimeService) Prepare(ctx context.Context, req *runtime.Requirement) error {
	return s.runtime.Prepare(ctx, req)
}

// Execute runs a project with the given configuration.
func (s *RuntimeService) Execute(ctx context.Context, project *runtime.Project, config *runtime.RunConfig) (*runtime.RunResult, error) {
	return s.runtime.Execute(ctx, project, config)
}

// Stop terminates a running execution.
func (s *RuntimeService) Stop(ctx context.Context, runID string) error {
	return s.runtime.Stop(ctx, runID)
}

// Status returns the current status of a running execution.
func (s *RuntimeService) Status(ctx context.Context, runID string) (*runtime.RunStatus, error) {
	return s.runtime.Status(ctx, runID)
}

// ListRunning returns all currently running executions.
func (s *RuntimeService) ListRunning() []*runtime.RunStatus {
	return s.runtime.ListRunning()
}

// ExecuteCommand runs a raw command via the executor (lower-level).
func (s *RuntimeService) ExecuteCommand(ctx context.Context, config runtime.RunConfig) *runtime.RunResult {
	return s.executor.Execute(ctx, config)
}
