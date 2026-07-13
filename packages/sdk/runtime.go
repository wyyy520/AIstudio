package sdk

import (
	"context"

	"github.com/aistudio/packages/runtime"
)

type RunResult = runtime.RunResult
type RunStatus = runtime.RunStatus
type BundleSpec = runtime.BundleSpec
type Bundle = runtime.Bundle

func Execute(project *runtime.Project, runConfig *runtime.RunConfig) (*RunResult, error) {
	exec := runtime.NewLocalExecutor()
	return exec.Execute(context.Background(), *runConfig), nil
}

func GetStatus(runID string) (*RunStatus, error) {
	exec := runtime.NewLocalExecutor()
	status, ok := exec.Status(runID)
	if !ok {
		return nil, nil
	}
	return status, nil
}

func InstallBundle(spec *BundleSpec) (*Bundle, error) {
	mgr := runtime.NewBundleManager("")
	req := runtime.RequirementFromSpec(spec)
	return mgr.Install(context.Background(), req, nil)
}