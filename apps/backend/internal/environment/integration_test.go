package environment

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/aistudio/backend/internal/eventbus"
	"github.com/aistudio/backend/internal/runtime"
)

// mockBundleManager implements runtime.BundleManager for testing.
type mockBundleManager struct {
	mu          sync.Mutex
	installed   map[string]*runtime.Bundle
	installFunc func(ctx context.Context, req *runtime.Requirement, progress runtime.ProgressCallback) (*runtime.Bundle, error)
	detectFunc  func(ctx context.Context, spec *runtime.BundleSpec) (*runtime.Bundle, bool)
}

func newMockBundleManager() *mockBundleManager {
	return &mockBundleManager{
		installed: make(map[string]*runtime.Bundle),
	}
}

func (m *mockBundleManager) List() []*runtime.Bundle {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]*runtime.Bundle, 0, len(m.installed))
	for _, b := range m.installed {
		result = append(result, b)
	}
	return result
}

func (m *mockBundleManager) Get(name string) (*runtime.Bundle, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.installed[name]
	return b, ok
}

func (m *mockBundleManager) Install(ctx context.Context, req *runtime.Requirement, progress runtime.ProgressCallback) (*runtime.Bundle, error) {
	if m.installFunc != nil {
		return m.installFunc(ctx, req, progress)
	}
	if progress == nil {
		progress = func(string, float64) {}
	}
	progress("installing", 0.5)
	progress("done", 1.0)
	bundle := &runtime.Bundle{
		Name:    req.Name,
		Version: req.Version,
		Path:    "/tmp/mock/" + req.Name,
	}
	m.mu.Lock()
	m.installed[req.Name] = bundle
	m.mu.Unlock()
	return bundle, nil
}

func (m *mockBundleManager) InstallFromSpec(ctx context.Context, spec *runtime.BundleSpec, progress runtime.ProgressCallback) (*runtime.Bundle, error) {
	return m.Install(ctx, &runtime.Requirement{
		Name:     spec.Name,
		Version:  spec.Version,
		Python:   spec.Python,
		Packages: spec.Packages,
		Commands: spec.Commands,
		GPU:      !spec.GPUOptional,
	}, progress)
}

func (m *mockBundleManager) Uninstall(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.installed, name)
	return nil
}

func (m *mockBundleManager) CachePath() string {
	return "/tmp/mock"
}

func (m *mockBundleManager) SharedBundles() []*runtime.Bundle {
	return m.List()
}

func (m *mockBundleManager) Clean(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.installed = make(map[string]*runtime.Bundle)
	return nil
}

func (m *mockBundleManager) DetectInstalled(ctx context.Context, spec *runtime.BundleSpec) (*runtime.Bundle, bool) {
	if m.detectFunc != nil {
		return m.detectFunc(ctx, spec)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.installed[spec.Name]
	return b, ok
}

// TestEnvironmentIntegration_EnsureEnvironment_BundleAlreadyInstalled verifies
// that EnsureEnvironment returns immediately when the bundle is already installed.
func TestEnvironmentIntegration_EnsureEnvironment_BundleAlreadyInstalled(t *testing.T) {
	mgr := NewManager()
	bm := newMockBundleManager()
	eb := eventbus.New()
	ei := NewEnvironmentIntegration(mgr, bm, eb)

	// Pre-install a bundle
	bm.installed["test-bundle"] = &runtime.Bundle{
		Name:    "test-bundle",
		Version: "1.0.0",
		Path:    "/tmp/mock/test-bundle",
	}

	req := &runtime.Requirement{
		Name:    "test-bundle",
		Version: "1.0.0",
	}

	ctx := context.Background()
	report, err := ei.EnsureEnvironment(ctx, req)
	if err != nil {
		t.Fatalf("EnsureEnvironment failed: %v", err)
	}
	if !report.Ready {
		t.Errorf("expected report to be ready, got Ready=false")
	}
}

// TestEnvironmentIntegration_EnsureEnvironment_InstallsWhenMissing verifies
// that EnsureEnvironment installs the bundle when it is missing.
func TestEnvironmentIntegration_EnsureEnvironment_InstallsWhenMissing(t *testing.T) {
	mgr := NewManager()
	bm := newMockBundleManager()
	eb := eventbus.New()
	ei := NewEnvironmentIntegration(mgr, bm, eb)

	req := &runtime.Requirement{
		Name:    "test-bundle",
		Version: "1.0.0",
	}

	ctx := context.Background()
	report, err := ei.EnsureEnvironment(ctx, req)
	if err != nil {
		t.Fatalf("EnsureEnvironment failed: %v", err)
	}
	if !report.Ready {
		t.Errorf("expected report to be ready, got Ready=false")
	}

	// Verify the bundle was installed
	_, ok := bm.Get("test-bundle")
	if !ok {
		t.Errorf("expected bundle 'test-bundle' to be installed")
	}
}

// TestEnvironmentIntegration_PrepareRuntime_ProgressCallback verifies
// that PrepareRuntime calls the progress callback with expected values.
func TestEnvironmentIntegration_PrepareRuntime_ProgressCallback(t *testing.T) {
	mgr := NewManager()
	bm := newMockBundleManager()
	eb := eventbus.New()
	ei := NewEnvironmentIntegration(mgr, bm, eb)

	req := &runtime.Requirement{
		Name:    "progress-bundle",
		Version: "1.0.0",
	}

	var progressEvents []struct {
		pct float64
		msg string
	}

	ctx := context.Background()
	err := ei.PrepareRuntime(ctx, req, func(pct float64, msg string) {
		progressEvents = append(progressEvents, struct {
			pct float64
			msg string
		}{pct, msg})
	})
	if err != nil {
		t.Fatalf("PrepareRuntime failed: %v", err)
	}

	if len(progressEvents) == 0 {
		t.Fatal("expected at least 1 progress event, got 0")
	}

	// First event should be 0%
	if progressEvents[0].pct != 0 {
		t.Errorf("expected first progress to be 0, got %f", progressEvents[0].pct)
	}

	// Last event should be 1.0 (complete)
	last := progressEvents[len(progressEvents)-1]
	if last.pct != 1.0 {
		t.Errorf("expected last progress to be 1.0, got %f", last.pct)
	}
}

// TestEnvironmentIntegration_EventBusEvents verifies that events are published
// on the EventBus during EnsureEnvironment.
func TestEnvironmentIntegration_EventBusEvents(t *testing.T) {
	mgr := NewManager()
	bm := newMockBundleManager()
	eb := eventbus.New()
	ei := NewEnvironmentIntegration(mgr, bm, eb)

	var receivedEvents []eventbus.Event
	eb.Subscribe(eventbus.TopicEnvDetecting, func(e eventbus.Event) {
		receivedEvents = append(receivedEvents, e)
	})
	eb.Subscribe(eventbus.TopicEnvInstallingBundle, func(e eventbus.Event) {
		receivedEvents = append(receivedEvents, e)
	})
	eb.Subscribe(eventbus.TopicEnvBundleReady, func(e eventbus.Event) {
		receivedEvents = append(receivedEvents, e)
	})

	req := &runtime.Requirement{
		Name:    "event-bundle",
		Version: "1.0.0",
	}

	ctx := context.Background()
	_, err := ei.EnsureEnvironment(ctx, req)
	if err != nil {
		t.Fatalf("EnsureEnvironment failed: %v", err)
	}

	if len(receivedEvents) == 0 {
		t.Error("expected at least 1 EventBus event, got 0")
	}
}

// TestEnvironmentIntegration_InstallFailure verifies that installation errors
// are propagated correctly.
func TestEnvironmentIntegration_InstallFailure(t *testing.T) {
	mgr := NewManager()
	bm := newMockBundleManager()
	installErr := "mock install failure"
	bm.installFunc = func(ctx context.Context, req *runtime.Requirement, progress runtime.ProgressCallback) (*runtime.Bundle, error) {
		return nil, fmt.Errorf("%s", installErr)
	}
	eb := eventbus.New()
	ei := NewEnvironmentIntegration(mgr, bm, eb)

	req := &runtime.Requirement{
		Name:    "fail-bundle",
		Version: "1.0.0",
	}

	ctx := context.Background()
	_, err := ei.EnsureEnvironment(ctx, req)
	if err == nil {
		t.Fatal("expected error from EnsureEnvironment, got nil")
	}
}

// TestManager_CheckAndPrepare verifies the Manager.CheckAndPrepare bridge method.
func TestManager_CheckAndPrepare(t *testing.T) {
	mgr := NewManager()
	bm := newMockBundleManager()

	req := &runtime.Requirement{
		Name:    "check-bundle",
		Version: "1.0.0",
	}

	ctx := context.Background()
	report, err := mgr.CheckAndPrepare(ctx, req, bm)
	if err != nil {
		t.Fatalf("CheckAndPrepare failed: %v", err)
	}

	if report == nil {
		t.Fatal("expected non-nil report")
	}

	// Verify the bundle was installed
	_, ok := bm.Get("check-bundle")
	if !ok {
		t.Errorf("expected bundle 'check-bundle' to be installed")
	}
}
