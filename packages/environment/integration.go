package environment

import (
	"context"
	"fmt"

	"github.com/aistudio/packages/event"
)

type BundleManager interface {
	DetectInstalled(ctx context.Context, name, version string) bool
	Install(ctx context.Context, name, version string, packages, commands []string, gpu bool, progress func(string, float64)) error
	GetBundlePath(name, version string) string
}

type EnvironmentIntegration struct {
	envManager    *Manager
	bundleManager BundleManager
	eventBus      *event.EventBus
}

func NewEnvironmentIntegration(envManager *Manager, bundleManager BundleManager, eventBus *event.EventBus) *EnvironmentIntegration {
	return &EnvironmentIntegration{
		envManager:    envManager,
		bundleManager: bundleManager,
		eventBus:      eventBus,
	}
}

func (ei *EnvironmentIntegration) PrepareRuntime(ctx context.Context, req *Requirement, progressCb func(progress float64, msg string)) error {
	if progressCb == nil {
		progressCb = func(float64, string) {}
	}

	ei.emitEvent(event.TopicEnvDetecting, "detecting", 0, "Preparing environment for "+req.Name)
	progressCb(0, "Preparing environment for "+req.Name)

	if ei.bundleManager.DetectInstalled(ctx, req.Name, req.Version) {
		ei.emitEvent(event.TopicEnvBundleReady, "ready", 1.0, "Bundle already installed: "+req.Name)
		progressCb(1.0, "Environment ready")
		return nil
	}

	ei.emitEvent(event.TopicEnvDetecting, "detecting", 0.1, "Detecting system environment")
	progressCb(0.1, "Detecting system environment...")

	report := CheckRequirements(ctx, req)
	if report.Ready {
		ei.emitEvent(event.TopicEnvReady, "ready", 1.0, "All system requirements met")
		progressCb(1.0, "Environment ready")
		return nil
	}

	ei.emitEvent(event.TopicEnvInstallingBundle, "installing", 0.2, "Installing runtime bundle: "+req.Name)
	progressCb(0.2, "Installing bundle: "+req.Name)

	err := ei.bundleManager.Install(ctx, req.Name, req.Version, req.Packages, req.Commands, req.GPU, func(msg string, pct float64) {
		mappedProgress := 0.2 + pct*0.7
		ei.emitEvent(event.TopicEnvInstallingBundle, "installing", mappedProgress, msg)
		progressCb(mappedProgress, msg)
	})
	if err != nil {
		ei.emitEvent(event.TopicEnvError, "error", 0, "Installation failed: "+err.Error())
		return fmt.Errorf("prepare runtime: %w", err)
	}

	ei.emitEvent(event.TopicEnvBundleReady, "ready", 1.0, "Bundle installed and ready")
	progressCb(1.0, "Environment ready")
	return nil
}

func (ei *EnvironmentIntegration) EnsureEnvironment(ctx context.Context, req *Requirement) (*EnvironmentReport, error) {
	if ei.bundleManager.DetectInstalled(ctx, req.Name, req.Version) {
		report := NewEnvironmentReport()
		report.Ready = true
		return report, nil
	}

	report := CheckRequirements(ctx, req)
	if report.Ready {
		ei.emitEvent(event.TopicEnvReady, "ready", 1.0, "All requirements met")
		return report, nil
	}

	ei.emitEvent(event.TopicEnvInstallingBundle, "installing", 0.2, "Installing bundle: "+req.Name)
	err := ei.bundleManager.Install(ctx, req.Name, req.Version, req.Packages, req.Commands, req.GPU, func(msg string, pct float64) {
		ei.emitEvent(event.TopicEnvInstallingBundle, "installing", 0.2+pct*0.7, msg)
	})
	if err != nil {
		ei.emitEvent(event.TopicEnvError, "error", 0, "Install failed: "+err.Error())
		return report, fmt.Errorf("ensure environment: %w", err)
	}

	finalReport := NewEnvironmentReport()
	finalReport.Ready = true
	ei.emitEvent(event.TopicEnvBundleReady, "ready", 1.0, "Bundle installed ready")
	return finalReport, nil
}

func (ei *EnvironmentIntegration) emitEvent(topic event.Topic, status string, progress float64, message string) {
	if ei.eventBus != nil {
		ei.eventBus.Publish(topic, event.EnvEventData{
			BundleName: "",
			Status:     status,
			Progress:   progress,
			Message:    message,
		})
	}
}