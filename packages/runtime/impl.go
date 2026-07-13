package runtime

import (
	"context"
	"fmt"

	"github.com/aistudio/packages/environment"
	"github.com/aistudio/packages/event"
)

type runtimeImpl struct {
	bundleMgr BundleManager
	executor  CommandExecutor
	envMgr    *environment.Manager
	eventBus  *event.EventBus
}

func NewRuntime(bundleMgr BundleManager, executor CommandExecutor, envMgr *environment.Manager, eventBus *event.EventBus) Runtime {
	return &runtimeImpl{
		bundleMgr: bundleMgr,
		executor:  executor,
		envMgr:    envMgr,
		eventBus:  eventBus,
	}
}

func (r *runtimeImpl) Detect(ctx context.Context, req *Requirement) (*EnvironmentReport, error) {
	if req == nil {
		return nil, fmt.Errorf("requirement is nil")
	}

	r.emitRuntimeEvent("", "", "detecting", 0, "Checking environment requirements...")

	report := environment.CheckRequirements(ctx, req)
	if report == nil {
		return nil, fmt.Errorf("environment check returned nil")
	}

	if report.Ready {
		r.emitRuntimeEvent("", "", "ready", 1.0, "All environment requirements met")
	} else {
		r.emitRuntimeEvent("", "", "issues_found", 0.5,
			fmt.Sprintf("Found %d issues", len(report.Issues)))
	}

	return report, nil
}

func (r *runtimeImpl) Prepare(ctx context.Context, req *Requirement) error {
	if req == nil {
		return fmt.Errorf("requirement is nil")
	}

	r.emitRuntimeEvent("", "", "preparing", 0, "Preparing environment for "+req.Name)

	report := environment.CheckRequirements(ctx, req)
	if report.Ready {
		r.emitRuntimeEvent("", "", "ready", 1.0, "Environment already meets requirements")
		return nil
	}

	_, err := r.bundleMgr.Install(ctx, req, func(msg string, pct float64) {
		r.emitRuntimeEvent("", "", "installing", 0.2+pct*0.8, msg)
	})
	if err != nil {
		r.emitRuntimeEvent("", "", "failed", 0, "Preparation failed: "+err.Error())
		return fmt.Errorf("prepare runtime: %w", err)
	}

	r.emitRuntimeEvent("", "", "ready", 1.0, "Environment prepared successfully")
	return nil
}

func (r *runtimeImpl) Execute(ctx context.Context, project *Project, config *RunConfig) (*RunResult, error) {
	if project == nil {
		return nil, fmt.Errorf("project is nil")
	}
	if config == nil {
		return nil, fmt.Errorf("run config is nil")
	}

	if config.ProjectDir == "" {
		config.ProjectDir = project.RootPath
	}

	r.emitRuntimeEvent("", project.ID, "started", 0, "Starting execution: "+project.Name)

	if config.EntryPoint == "" {
		config.EntryPoint = project.EntryPoint
	}
	if config.EntryPoint == "" {
		config.EntryPoint = "main.py"
	}

	if config.Env == nil {
		config.Env = make(map[string]string)
	}
	config.Env["AISTUDIO_PROJECT_ID"] = project.ID
	config.Env["AISTUDIO_PROJECT_NAME"] = project.Name

	result := r.executor.Execute(ctx, *config)
	if result == nil {
		return nil, fmt.Errorf("executor returned nil result")
	}

	switch result.Status {
	case RunStatusCompleted:
		r.emitRuntimeEvent(result.RunID, project.ID, "completed", 1.0,
			fmt.Sprintf("Execution completed in %v (exit code %d)", result.Duration, result.ExitCode))
	case RunStatusFailed:
		r.emitRuntimeEvent(result.RunID, project.ID, "failed", 0,
			fmt.Sprintf("Execution failed: %s", result.Error))
	case RunStatusTimeout:
		r.emitRuntimeEvent(result.RunID, project.ID, "timeout", 0,
			fmt.Sprintf("Execution timed out after %v", result.Duration))
	case RunStatusStopped:
		r.emitRuntimeEvent(result.RunID, project.ID, "stopped", 0, "Execution stopped by user")
	default:
		r.emitRuntimeEvent(result.RunID, project.ID, "running", 0.5, "Execution in progress")
	}

	return result, nil
}

func (r *runtimeImpl) Stop(ctx context.Context, runID string) error {
	if runID == "" {
		return fmt.Errorf("runID is empty")
	}

	err := r.executor.Stop(runID)
	if err != nil {
		return err
	}

	r.emitRuntimeEvent(runID, "", "stopped", 0, "Execution stopped")
	return nil
}

func (r *runtimeImpl) Status(ctx context.Context, runID string) (*RunStatus, error) {
	if runID == "" {
		return nil, fmt.Errorf("runID is empty")
	}

	status, ok := r.executor.Status(runID)
	if !ok {
		return nil, fmt.Errorf("run not found: %s", runID)
	}
	return status, nil
}

func (r *runtimeImpl) ListRunning() []*RunStatus {
	return r.executor.ListRunning()
}

func (r *runtimeImpl) emitRuntimeEvent(runID, projectID, status string, progress float64, message string) {
	if r.eventBus == nil {
		return
	}
	r.eventBus.Publish(event.TopicRuntimeStarted, event.RuntimeEventData{
		RunID:     runID,
		ProjectID: projectID,
		Status:    status,
		Progress:  progress,
		Message:   message,
	})
}

func (r *runtimeImpl) LogEntryToEvent(entry LogEntry) {
	if r.eventBus == nil {
		return
	}
	r.eventBus.Publish(event.TopicRuntimeLog, event.RuntimeEventData{
		RunID:   entry.RunID,
		Status:  "running",
		Message: entry.Message,
	})
	r.eventBus.Publish(event.TopicLogEntry, event.LogEventData{
		Level:   entry.Level,
		Message: entry.Message,
		Source:  entry.Source,
		Raw:     entry.Raw,
	})
}

var _ Runtime = (*runtimeImpl)(nil)