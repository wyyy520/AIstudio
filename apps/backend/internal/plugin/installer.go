// Package plugin — Plugin Installer
//
// The Installer handles the full plugin lifecycle derived from a manifest URL
// or local directory: fetch manifest → download sources → install dependencies
// → register with the plugin registry.  It supports asynchronous installs
// that can be polled for progress via GetInstallStatus.
package plugin

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Types & constants
// ---------------------------------------------------------------------------

// InstallPhase marks the current stage of an install task.
type InstallPhase string

const (
	PhaseCheckDeps InstallPhase = "check_deps"
	PhaseDownload  InstallPhase = "download"
	PhaseInstall   InstallPhase = "install"
	PhaseRegister  InstallPhase = "register"
)

// InstallStatus carries the live state of an ongoing or completed install.
type InstallStatus struct {
	TaskID      string       `json:"task_id"`
	PluginName  string       `json:"plugin_name"`
	Phase       InstallPhase `json:"phase"`
	Progress    float64      `json:"progress"`
	Log         []string     `json:"log"`
	Error       string       `json:"error,omitempty"`
	StartedAt   time.Time    `json:"started_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
}

// InstallTask represents one async install operation.  Callers can wait on
// Done or poll GetInstallStatus.
type InstallTask struct {
	ID     string
	Plugin string
	Status InstallStatus
	Done   chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

// Installer orchestrates plugin acquisition and registration.
type Installer struct {
	pluginsDir string
	tasks      map[string]*InstallTask
	registry   *Registry
	mu         sync.RWMutex
}

// NewInstaller creates an Installer that stores plugins under pluginsDir.
func NewInstaller(pluginsDir string, registry *Registry) *Installer {
	return &Installer{
		pluginsDir: pluginsDir,
		tasks:      make(map[string]*InstallTask),
		registry:   registry,
	}
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// Install starts an async install from manifestURL (URL or local path).
func (inst *Installer) Install(ctx context.Context, manifestURL string) (*InstallTask, error) {
	taskID := uuid.New().String()
	ctx, cancel := context.WithCancel(ctx)

	task := &InstallTask{
		ID:   taskID,
		Done: make(chan struct{}),
		Status: InstallStatus{
			TaskID:    taskID,
			Phase:     PhaseCheckDeps,
			Progress:  0,
			Log:       []string{},
			StartedAt: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	inst.mu.Lock()
	inst.tasks[taskID] = task
	inst.mu.Unlock()

	go inst.runInstall(task, manifestURL)
	return task, nil
}

// InstallSync is a convenience wrapper that blocks until install completes.
func (inst *Installer) InstallSync(ctx context.Context, manifestURL string) error {
	task, err := inst.Install(ctx, manifestURL)
	if err != nil {
		return err
	}
	select {
	case <-task.Done:
		if task.Status.Error != "" {
			return fmt.Errorf("install failed: %s", task.Status.Error)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Uninstall removes a plugin directory and unregisters it.
func (inst *Installer) Uninstall(ctx context.Context, pluginName string) error {
	pluginDir := filepath.Join(inst.pluginsDir, pluginName)
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return fmt.Errorf("plugin not found: %s", pluginName)
	}
	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("remove plugin directory: %w", err)
	}
	return nil
}

// GetInstallStatus returns a snapshot of an install task, or nil.
func (inst *Installer) GetInstallStatus(taskID string) *InstallStatus {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	task, ok := inst.tasks[taskID]
	if !ok {
		return nil
	}
	s := task.Status
	return &s
}

// GetInstallStatusByName finds the first task matching pluginName.
func (inst *Installer) GetInstallStatusByName(pluginName string) *InstallStatus {
	inst.mu.RLock()
	defer inst.mu.RUnlock()
	for _, t := range inst.tasks {
		if t.Plugin == pluginName {
			s := t.Status
			return &s
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Install pipeline (runs in background goroutine)
// ---------------------------------------------------------------------------

func (inst *Installer) runInstall(task *InstallTask, source string) {
	defer close(task.Done)
	defer func() {
		now := time.Now()
		task.Status.CompletedAt = &now
	}()

	// Phase 1 — fetch manifest
	inst.log(task, PhaseCheckDeps, 0.1, "fetching manifest...")
	mv, err := inst.fetchManifest(source)
	if err != nil {
		task.Status.Error = fmt.Sprintf("fetch manifest: %v", err)
		return
	}
	pluginName := mv.Name
	task.Plugin = pluginName

	// Phase 2 — download / copy sources
	inst.log(task, PhaseDownload, 0.3, "downloading plugin %s...", pluginName)
	pluginDir := filepath.Join(inst.pluginsDir, pluginName)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		task.Status.Error = fmt.Sprintf("create plugin dir: %v", err)
		return
	}
	if err := inst.downloadPlugin(task.ctx, pluginDir, source); err != nil {
		task.Status.Error = fmt.Sprintf("download plugin: %v", err)
		return
	}

	// Phase 3 — pip install + install.py
	inst.log(task, PhaseInstall, 0.6, "installing dependencies...")
	if err := inst.installDependencies(task, pluginDir); err != nil {
		task.Status.Error = fmt.Sprintf("install dependencies: %v", err)
		return
	}

	// Phase 4 — register
	inst.log(task, PhaseRegister, 0.9, "registering plugin...")
	p, err := loadPluginFromManifest(filepath.Join(pluginDir, "plugin.json"))
	if err != nil {
		task.Status.Error = fmt.Sprintf("manifest missing in installed plugin: %v", err)
		return
	}
	if err := inst.registry.Register(p); err != nil {
		task.Status.Error = fmt.Sprintf("register plugin: %v", err)
		return
	}

	task.Status.PluginName = pluginName
	inst.log(task, PhaseRegister, 1.0, "plugin installed successfully")
}

// ---------------------------------------------------------------------------
// Manifest retrieval
// ---------------------------------------------------------------------------

func (inst *Installer) fetchManifest(source string) (*ManifestV2, error) {
	// Try URL first, fall back to local file
	resp, err := http.Get(source)
	if err != nil {
		data, readErr := os.ReadFile(source)
		if readErr != nil {
			return nil, fmt.Errorf("fetch url: %w, read local: %v", err, readErr)
		}
		var mv ManifestV2
		if err := json.Unmarshal(data, &mv); err != nil {
			return nil, fmt.Errorf("parse local manifest: %w", err)
		}
		if err := validateManifest(&mv); err != nil {
			return nil, err
		}
		return &mv, nil
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	var mv ManifestV2
	if err := json.Unmarshal(data, &mv); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if err := validateManifest(&mv); err != nil {
		return nil, err
	}
	return &mv, nil
}

func validateManifest(mv *ManifestV2) error {
	if mv.ID == "" {
		return fmt.Errorf("manifest missing required field: id")
	}
	if mv.Name == "" {
		return fmt.Errorf("manifest missing required field: name")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Source acquisition — download / copy / extract
// ---------------------------------------------------------------------------

// downloadPlugin routes to the correct strategy: URL download, local dir copy,
// or archive extraction.
func (inst *Installer) downloadPlugin(ctx context.Context, destDir, source string) error {
	if isRemoteURL(source) {
		return inst.downloadFromURL(ctx, destDir, source)
	}
	srcInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("source not accessible: %w", err)
	}
	if srcInfo.IsDir() {
		return inst.copyDirectory(source, destDir)
	}
	return inst.extractArchive(source, destDir)
}

// downloadFromURL fetches a remote file, saves it to a temp location, and
// either extracts (zip) or copies the raw file into destDir.
func (inst *Installer) downloadFromURL(ctx context.Context, destDir, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "plugin-*.download")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write download: %w", err)
	}
	tmpFile.Close()

	// If it's a zip archive, extract; otherwise copy as-is
	if strings.HasSuffix(strings.ToLower(url), ".zip") {
		return unzip(tmpPath, destDir)
	}
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(destDir, filepath.Base(url)), data, 0644)
}

// copyDirectory recursively copies src → dst.
func (inst *Installer) copyDirectory(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, relPath)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
}

// extractArchive extracts .zip archives to destDir.
func (inst *Installer) extractArchive(path, destDir string) error {
	if strings.ToLower(filepath.Ext(path)) == ".zip" {
		return unzip(path, destDir)
	}
	return fmt.Errorf("unsupported archive format %q — extract manually to %s", filepath.Ext(path), destDir)
}

// unzip extracts a .zip file to destDir with ZipSlip protection.
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	dest = filepath.Clean(dest)

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		// Prevent ZipSlip (path traversal)
		destWithSep := dest + string(os.PathSeparator)
		if fpath != dest && !strings.HasPrefix(fpath, destWithSep) {
			return fmt.Errorf("illegal file path in zip: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, copyErr := io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Dependency installation
// ---------------------------------------------------------------------------

// installDependencies runs pip install (requirements.txt) and install.py for a plugin.
func (inst *Installer) installDependencies(task *InstallTask, pluginDir string) error {
	// Step 1 — pip install -r requirements.txt
	reqFile := filepath.Join(pluginDir, "requirements.txt")
	if _, err := os.Stat(reqFile); err == nil {
		appendLog(task, "installing pip dependencies from requirements.txt...")
		cmd := execCommand("pip", "install", "-r", reqFile)
		cmd.Dir = pluginDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			appendLog(task, "pip install warning: %v", err)
			if len(out) > 0 {
				appendLog(task, "%s", string(out))
			}
			// Non‑fatal — plugin may still work without optional deps
		} else {
			appendLog(task, "pip dependencies installed successfully")
		}
	}

	// Step 2 — run install.py
	installScript := filepath.Join(pluginDir, "install.py")
	if _, err := os.Stat(installScript); err == nil {
		appendLog(task, "running install.py...")
		cmd := execCommand("python", installScript)
		cmd.Dir = pluginDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			appendLog(task, "%s", string(out))
			return fmt.Errorf("install.py failed: %w", err)
		}
		appendLog(task, "install.py completed successfully")
	}

	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// execCommand is a thin wrapper so tests can replace os/exec usage.
var execCommand = func(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

func isRemoteURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func (inst *Installer) log(task *InstallTask, phase InstallPhase, progress float64, format string, args ...interface{}) {
	task.Status.Phase = phase
	task.Status.Progress = progress
	msg := fmt.Sprintf(format, args...)
	task.Status.Log = append(task.Status.Log, msg)
}

func appendLog(task *InstallTask, format string, args ...interface{}) {
	task.Status.Log = append(task.Status.Log, fmt.Sprintf(format, args...))
}
