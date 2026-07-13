package runtime

import (
	"context"
)

type localRuntime struct {
	executor CommandExecutor
}

func NewLocalRuntime(executor CommandExecutor) Runtime {
	return &localRuntime{executor: executor}
}

func (r *localRuntime) Detect(ctx context.Context, req *Requirement) (*EnvironmentReport, error) {
	report := &EnvironmentReport{
		Ready:    true,
		Packages: make(map[string]string),
		Commands: make(map[string]bool),
		Issues:   []*Issue{},
		Warnings: []string{},
	}
	return report, nil
}

func (r *localRuntime) Prepare(ctx context.Context, req *Requirement) error {
	return nil
}

func (r *localRuntime) Execute(ctx context.Context, project *Project, config *RunConfig) (*RunResult, error) {
	if config == nil {
		config = &RunConfig{
			ProjectDir: project.RootPath,
			EntryPoint: project.EntryPoint,
		}
	}
	result := r.executor.Execute(ctx, *config)
	return &RunResult{
		RunID:      result.RunID,
		Status:     result.Status,
		ExitCode:   result.ExitCode,
		Stdout:     result.Stdout,
		Stderr:     result.Stderr,
		Duration:   result.Duration,
		StartedAt:  result.StartedAt,
		CompletedAt: result.CompletedAt,
		Error:      result.Error,
	}, nil
}

func (r *localRuntime) Stop(ctx context.Context, runID string) error {
	return r.executor.Stop(runID)
}

func (r *localRuntime) Status(ctx context.Context, runID string) (*RunStatus, error) {
	status, _ := r.executor.Status(runID)
	return status, nil
}

func (r *localRuntime) ListRunning() []*RunStatus {
	return r.executor.ListRunning()
}