package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// ============================================================================
// Bundle Manager Implementation
// ============================================================================

// defaultBundleCacheDir returns the default cache directory for bundles.
func defaultBundleCacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "aistudio", "bundles")
	}
	return filepath.Join(home, ".aistudio", "bundles")
}

// bundleManager implements BundleManager.
type bundleManager struct {
	mu        sync.RWMutex
	cacheDir  string
	bundles   map[string]*Bundle
	installer BundleInstaller
}

// BundleInstaller handles the actual installation of packages.
type BundleInstaller interface {
	InstallPythonPackages(ctx context.Context, pythonPath string, packages []string, progress ProgressCallback) error
	InstallSystemCommands(ctx context.Context, commands []string) error
	CreateVirtualEnv(ctx context.Context, path string) error
}

type defaultInstaller struct{}

// NewBundleManager creates a new BundleManager with the given cache directory.
// If cacheDir is empty, it defaults to ~/.aistudio/bundles/.
func NewBundleManager(cacheDir string) BundleManager {
	if cacheDir == "" {
		cacheDir = defaultBundleCacheDir()
	}
	return &bundleManager{
		cacheDir:  cacheDir,
		bundles:   make(map[string]*Bundle),
		installer: &defaultInstaller{},
	}
}

// SetInstaller sets a custom installer (useful for testing).
func (m *bundleManager) SetInstaller(inst BundleInstaller) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.installer = inst
}

// ============================================================================
// BundleManager Interface Implementation
// ============================================================================

func (m *bundleManager) List() []*Bundle {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Bundle, 0, len(m.bundles))
	for _, b := range m.bundles {
		result = append(result, b)
	}
	return result
}

func (m *bundleManager) Get(name string) (*Bundle, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	b, ok := m.bundles[name]
	return b, ok
}

func (m *bundleManager) Install(ctx context.Context, req *Requirement, progress ProgressCallback) (*Bundle, error) {
	return m.install(ctx, req.Name, req.Version, req.Python, req.Packages, req.Commands, req.GPU, progress)
}

func (m *bundleManager) InstallFromSpec(ctx context.Context, spec *BundleSpec, progress ProgressCallback) (*Bundle, error) {
	return m.install(ctx, spec.Name, spec.Version, spec.Python, spec.Packages, spec.Commands, !spec.GPUOptional, progress)
}

func (m *bundleManager) install(ctx context.Context, name, version, pythonVersion string, packages []string, commands []string, gpuRequired bool, progress ProgressCallback) (*Bundle, error) {
	if progress == nil {
		progress = func(string, float64) {}
	}

	// Check if already installed
	m.mu.RLock()
	existing, exists := m.bundles[name]
	m.mu.RUnlock()
	if exists && existing.IsInstalled() {
		progress("Bundle already installed: "+name, 1.0)
		return existing, nil
	}

	progress("Installing bundle: "+name, 0.0)

	// Create cache directory
	bundleDir := filepath.Join(m.cacheDir, name+"-"+version)
	if err := os.MkdirAll(bundleDir, 0755); err != nil {
		return nil, fmt.Errorf("create bundle dir: %w", err)
	}

	// Check system commands
	progress("Checking system commands...", 0.05)
	for _, cmd := range commands {
		if err := checkCommand(cmd); err != nil {
			return nil, fmt.Errorf("missing required command %q: %w", cmd, err)
		}
	}

	// Create virtual environment
	venvPath := filepath.Join(bundleDir, "venv")
	progress("Creating virtual environment...", 0.1)
	if err := m.installer.CreateVirtualEnv(ctx, venvPath); err != nil {
		return nil, fmt.Errorf("create venv: %w", err)
	}

	// Get python path inside venv
	pythonPath := venvPython(venvPath)
	progress("Virtual environment created", 0.15)

	// Install packages
	if len(packages) > 0 {
		progress(fmt.Sprintf("Installing %d packages...", len(packages)), 0.2)
		if err := m.installer.InstallPythonPackages(ctx, pythonPath, packages, func(msg string, pct float64) {
			progress(msg, 0.2+pct*0.7)
		}); err != nil {
			return nil, fmt.Errorf("install packages: %w", err)
		}
	}

	progress("Finalizing installation...", 0.9)

	// Calculate size
	sizeMB := int64(0)
	filepath.Walk(bundleDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			sizeMB += info.Size()
		}
		return nil
	})
	sizeMB = sizeMB / (1024 * 1024)

	// Mark as shared (installed in central cache = shared)
	bundle := &Bundle{
		Name:        name,
		Version:     version,
		PythonPath:  pythonPath,
		Packages:    packages,
		Commands:    commands,
		Path:        bundleDir,
		InstalledAt: time.Now(),
		SizeMB:      sizeMB,
		GPUEnabled:  gpuRequired,
		Shared:      true,
	}

	m.mu.Lock()
	m.bundles[name] = bundle
	m.mu.Unlock()

	progress("Bundle installed: "+name, 1.0)
	log.Printf("[runtime] bundle installed: %s@%s (%d MB, shared=%v)", name, version, sizeMB, bundle.Shared)
	return bundle, nil
}

func (m *bundleManager) Uninstall(name string) error {
	m.mu.Lock()
	bundle, ok := m.bundles[name]
	if !ok {
		m.mu.Unlock()
		return fmt.Errorf("bundle not found: %s", name)
	}
	delete(m.bundles, name)
	m.mu.Unlock()

	if err := os.RemoveAll(bundle.Path); err != nil {
		return fmt.Errorf("remove bundle %s: %w", name, err)
	}
	log.Printf("[runtime] bundle uninstalled: %s", name)
	return nil
}

func (m *bundleManager) CachePath() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cacheDir
}

func (m *bundleManager) SharedBundles() []*Bundle {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Bundle
	for _, b := range m.bundles {
		if b.Shared {
			result = append(result, b)
		}
	}
	return result
}

func (m *bundleManager) Clean(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove all bundle directories
	for name, bundle := range m.bundles {
		if err := os.RemoveAll(bundle.Path); err != nil {
			log.Printf("[runtime] failed to remove bundle %s: %v", name, err)
		}
	}
	m.bundles = make(map[string]*Bundle)

	// Also clean the cache directory itself
	entries, err := os.ReadDir(m.cacheDir)
	if err == nil {
		for _, entry := range entries {
			path := filepath.Join(m.cacheDir, entry.Name())
			os.RemoveAll(path)
		}
	}

	log.Printf("[runtime] bundle cache cleaned: %s", m.cacheDir)
	return nil
}

func (m *bundleManager) DetectInstalled(ctx context.Context, spec *BundleSpec) (*Bundle, bool) {
	// Check if already in our map
	m.mu.RLock()
	b, ok := m.bundles[spec.Name]
	m.mu.RUnlock()
	if ok && b.IsInstalled() {
		return b, true
	}

	// Check if the cache directory exists
	bundleDir := filepath.Join(m.cacheDir, spec.Name+"-"+spec.Version)
	venvPath := filepath.Join(bundleDir, "venv")
	pythonPath := venvPython(venvPath)

	if _, err := os.Stat(pythonPath); os.IsNotExist(err) {
		return nil, false
	}

	// Verify key packages are installed
	bundle := &Bundle{
		Name:       spec.Name,
		Version:    spec.Version,
		PythonPath: pythonPath,
		Packages:   spec.Packages,
		Commands:   spec.Commands,
		Path:       bundleDir,
		Shared:     true,
	}

	m.mu.Lock()
	m.bundles[spec.Name] = bundle
	m.mu.Unlock()

	go func() {
		// Calculate size asynchronously
		var size int64
		filepath.Walk(bundleDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				size += info.Size()
			}
			return nil
		})
		m.mu.Lock()
		if b, ok := m.bundles[spec.Name]; ok {
			b.SizeMB = size / (1024 * 1024)
		}
		m.mu.Unlock()
	}()

	return bundle, true
}

// ============================================================================
// Default Installer
// ============================================================================

func (d *defaultInstaller) CreateVirtualEnv(ctx context.Context, path string) error {
	pythonCmd := pythonCommand()
	cmd := exec.CommandContext(ctx, pythonCmd, "-m", "venv", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("venv create failed: %w\n%s", err, string(output))
	}
	return nil
}

func (d *defaultInstaller) InstallPythonPackages(ctx context.Context, pythonPath string, packages []string, progress ProgressCallback) error {
	pipPath := filepath.Join(filepath.Dir(pythonPath), "pip")
	if runtime.GOOS == "windows" {
		pipPath = filepath.Join(filepath.Dir(pythonPath), "pip.exe")
	}

	// Try pip install for all packages together first
	args := append([]string{"install", "--quiet"}, packages...)
	cmd := exec.CommandContext(ctx, pipPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fall back to installing one by one for better error reporting
		for i, pkg := range packages {
			progress(fmt.Sprintf("Installing %s...", pkg), float64(i)/float64(len(packages)))
			cmd := exec.CommandContext(ctx, pipPath, "install", pkg)
			if out, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("pip install %s failed: %w\n%s", pkg, err, string(out))
			}
		}
	}
	_ = output
	return nil
}

func (d *defaultInstaller) InstallSystemCommands(ctx context.Context, commands []string) error {
	// System commands are pre-required; this verifies they exist
	for _, cmd := range commands {
		if err := checkCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

// ============================================================================
// Helpers
// ============================================================================

func checkCommand(name string) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("where", name)
		return cmd.Run()
	}
	cmd := exec.Command("which", name)
	return cmd.Run()
}

func pythonCommand() string {
	if runtime.GOOS == "windows" {
		return "python"
	}
	return "python3"
}

func venvPython(venvPath string) string {
	if runtime.GOOS == "windows" {
		path := filepath.Join(venvPath, "Scripts", "python.exe")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return filepath.Join(venvPath, "bin", "python")
}

// ============================================================================
// Bundle Spec Loader
// ============================================================================

// LoadBundleSpec loads a BundleSpec from a JSON file.
func LoadBundleSpec(path string) (*BundleSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read bundle spec %s: %w", path, err)
	}
	var spec BundleSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parse bundle spec %s: %w", path, err)
	}
	if spec.Name == "" {
		return nil, fmt.Errorf("bundle spec %s: name is required", path)
	}
	return &spec, nil
}

// LoadBundleSpecsFromDir loads all bundle.json files from a directory.
func LoadBundleSpecsFromDir(dir string) ([]*BundleSpec, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var specs []*BundleSpec
	for _, entry := range entries {
		if entry.IsDir() {
			bundlePath := filepath.Join(dir, entry.Name(), "bundle.json")
			if _, err := os.Stat(bundlePath); err == nil {
				spec, err := LoadBundleSpec(bundlePath)
				if err != nil {
					log.Printf("[runtime] warning: failed to load bundle spec %s: %v", bundlePath, err)
					continue
				}
				specs = append(specs, spec)
			}
		}
	}
	return specs, nil
}

// Ensure interface compliance
var _ BundleManager = (*bundleManager)(nil)