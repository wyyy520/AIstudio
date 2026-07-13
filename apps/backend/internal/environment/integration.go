package environment

import (
	"context"
	"fmt"

	"github.com/aistudio/backend/internal/eventbus"
	"github.com/aistudio/backend/internal/runtime"
)

// EnvironmentIntegration wires Environment Manager + Bundle Manager + EventBus
// to provide a unified environment preparation API for the service layer.
type EnvironmentIntegration struct {
	envManager    *Manager
	bundleManager runtime.BundleManager
	eventBus      *eventbus.EventBus
}

// NewEnvironmentIntegration creates a new EnvironmentIntegration.
func NewEnvironmentIntegration(envManager *Manager, bundleManager runtime.BundleManager, eventBus *eventbus.EventBus) *EnvironmentIntegration {
	return &EnvironmentIntegration{
		envManager:    envManager,
		bundleManager: bundleManager,
		eventBus:      eventBus,
	}
}

// PrepareRuntime prepares the runtime environment for the given Requirement.
// Progress is reported through both the EventBus and an optional callback.
func (ei *EnvironmentIntegration) PrepareRuntime(ctx context.Context, req *runtime.Requirement, progressCb func(progress float64, msg string)) error {
	if progressCb == nil {
		progressCb = func(float64, string) {}
	}

	ei.emitEvent(eventbus.TopicEnvDetecting, "detecting", 0, "Preparing environment for "+req.Name)
	progressCb(0, "Preparing environment for "+req.Name)

	// Check if bundle is already installed
	spec := &runtime.BundleSpec{
		Name:        req.Name,
		Version:     req.Version,
		Python:      req.Python,
		Packages:    req.Packages,
		Commands:    req.Commands,
		GPUOptional: !req.GPU,
	}
	if bundle, ok := ei.bundleManager.DetectInstalled(ctx, spec); ok {
		ei.emitEvent(eventbus.TopicEnvBundleReady, "ready", 1.0, "Bundle already installed: "+bundle.Name)
		progressCb(1.0, "Environment ready")
		return nil
	}

	// Detect system environment
	ei.emitEvent(eventbus.TopicEnvDetecting, "detecting", 0.1, "Detecting system environment")
	progressCb(0.1, "Detecting system environment...")

	report := runtime.CheckRequirements(ctx, req)
	if report.Ready {
		ei.emitEvent(eventbus.TopicEnvReady, "ready", 1.0, "All system requirements met")
		progressCb(1.0, "Environment ready")
		return nil
	}

	// Install bundle with progress
	ei.emitEvent(eventbus.TopicEnvInstallingBundle, "installing", 0.2, "Installing runtime bundle: "+req.Name)
	progressCb(0.2, "Installing bundle: "+req.Name)

	_, err := ei.bundleManager.Install(ctx, req, func(msg string, pct float64) {
		mappedProgress := 0.2 + pct*0.7
		ei.emitEvent(eventbus.TopicEnvInstallingBundle, "installing", mappedProgress, msg)
		progressCb(mappedProgress, msg)
	})
	if err != nil {
		ei.emitEvent(eventbus.TopicEnvError, "error", 0, "Installation failed: "+err.Error())
		return fmt.Errorf("prepare runtime: %w", err)
	}

	ei.emitEvent(eventbus.TopicEnvBundleReady, "ready", 1.0, "Bundle installed and ready")
	progressCb(1.0, "Environment ready")
	return nil
}

// EnsureEnvironment checks if the runtime requirement is met, installs if necessary,
// and returns the environment report. This is a simpler synchronous variant.
func (ei *EnvironmentIntegration) EnsureEnvironment(ctx context.Context, req *runtime.Requirement) (*runtime.EnvironmentReport, error) {
	// Check if bundle is already installed
	spec := &runtime.BundleSpec{
		Name:        req.Name,
		Version:     req.Version,
		Python:      req.Python,
		Packages:    req.Packages,
		Commands:    req.Commands,
		GPUOptional: !req.GPU,
	}
	if _, ok := ei.bundleManager.DetectInstalled(ctx, spec); ok {
		report := runtime.NewEnvironmentReport(ctx)
		report.Ready = true
		return report, nil
	}

	// Detect and check requirements (but always install the bundle if not detected)
	report := runtime.CheckRequirements(ctx, req)

	// Install bundle regardless of system requirements check,
	// because the bundle itself contains the runtime environment.
	ei.emitEvent(eventbus.TopicEnvInstallingBundle, "installing", 0.2, "Installing bundle: "+req.Name)
	bundle, err := ei.bundleManager.Install(ctx, req, func(msg string, pct float64) {
		ei.emitEvent(eventbus.TopicEnvInstallingBundle, "installing", 0.2+pct*0.7, msg)
	})
	if err != nil {
		ei.emitEvent(eventbus.TopicEnvError, "error", 0, "Install failed: "+err.Error())
		return report, fmt.Errorf("ensure environment: %w", err)
	}

	// Re-detect after install
	finalReport := runtime.NewEnvironmentReport(ctx)
	finalReport.Ready = true
	ei.emitEvent(eventbus.TopicEnvBundleReady, "ready", 1.0, "Bundle installed: "+bundle.Name+"@"+bundle.Version)
	return finalReport, nil
}

func (ei *EnvironmentIntegration) GetManager() *Manager {
	return ei.envManager
}

func (ei *EnvironmentIntegration) emitEvent(topic eventbus.Topic, status string, progress float64, message string) {
	if ei.eventBus != nil {
		ei.eventBus.Publish(topic, eventbus.EnvEventData{
			BundleName: "",
			Status:     status,
			Progress:   progress,
			Message:    message,
		})
	}
}
