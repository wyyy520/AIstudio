package runtime

import (
	"context"

	"github.com/aistudio/packages/environment"
)

type Runtime interface {
	Detect(ctx context.Context, req *Requirement) (*EnvironmentReport, error)
	Prepare(ctx context.Context, req *Requirement) error
	Execute(ctx context.Context, project *Project, config *RunConfig) (*RunResult, error)
	Stop(ctx context.Context, runID string) error
	Status(ctx context.Context, runID string) (*RunStatus, error)
	ListRunning() []*RunStatus
}

func RequirementFromSpec(spec *BundleSpec) *Requirement {
	return &environment.Requirement{
		Name:     spec.Name,
		Version:  spec.Version,
		Python:   spec.Python,
		Packages: spec.Packages,
		Commands: spec.Commands,
		GPU:      !spec.GPUOptional,
	}
}