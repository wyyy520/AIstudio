package service

import (
	"context"

	"github.com/aistudio/backend/internal/runtime"
)

// BundleService provides runtime bundle management operations.
type BundleService struct {
	manager runtime.BundleManager
}

// NewBundleService creates a new BundleService.
func NewBundleService(manager runtime.BundleManager) *BundleService {
	return &BundleService{manager: manager}
}

// List returns all installed bundles.
func (s *BundleService) List() []*runtime.Bundle {
	return s.manager.List()
}

// Get returns a bundle by name.
func (s *BundleService) Get(name string) (*runtime.Bundle, bool) {
	return s.manager.Get(name)
}

// Install installs a runtime bundle with progress reporting.
func (s *BundleService) Install(ctx context.Context, req *runtime.Requirement, progress runtime.ProgressCallback) (*runtime.Bundle, error) {
	return s.manager.Install(ctx, req, progress)
}

// InstallFromSpec installs a bundle from a BundleSpec.
func (s *BundleService) InstallFromSpec(ctx context.Context, spec *runtime.BundleSpec, progress runtime.ProgressCallback) (*runtime.Bundle, error) {
	return s.manager.InstallFromSpec(ctx, spec, progress)
}

// Uninstall removes a runtime bundle.
func (s *BundleService) Uninstall(name string) error {
	return s.manager.Uninstall(name)
}

// CachePath returns the cache directory for bundles.
func (s *BundleService) CachePath() string {
	return s.manager.CachePath()
}

// Clean removes all cached bundles.
func (s *BundleService) Clean(ctx context.Context) error {
	return s.manager.Clean(ctx)
}

// DetectInstalled checks which bundles are installed without installing.
func (s *BundleService) DetectInstalled(ctx context.Context, spec *runtime.BundleSpec) (*runtime.Bundle, bool) {
	return s.manager.DetectInstalled(ctx, spec)
}

// SharedBundles returns bundles that can be shared across projects.
func (s *BundleService) SharedBundles() []*runtime.Bundle {
	return s.manager.SharedBundles()
}
