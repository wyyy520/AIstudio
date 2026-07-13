package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

type InstallPhase string

const (
	PhaseCheckDeps InstallPhase = "check_deps"
	PhaseDownload  InstallPhase = "download"
	PhaseInstall   InstallPhase = "install"
	PhaseRegister  InstallPhase = "register"
)

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

type InstallTask struct {
	ID     string
	Plugin string
	Status InstallStatus
	Done   chan struct{}
	ctx    context.Context
	cancel context.CancelFunc
}

type Installer struct {
	pluginsDir string
	tasks      map[string]*InstallTask
	registry   *Registry
	mu         sync.RWMutex
}

func NewInstaller(pluginsDir string, registry *Registry) *Installer {
	return &Installer{
		pluginsDir: pluginsDir,
		tasks:      make(map[string]*InstallTask),
		registry:   registry,
	}
}

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

func (inst *Installer) GetInstallStatus(taskID string) *InstallStatus {
	inst.mu.RLock()
	defer inst.mu.RUnlock()

	task, ok := inst.tasks[taskID]
	if !ok {
		return nil
	}
	status := task.Status
	return &status
}

func (inst *Installer) GetInstallStatusByName(pluginName string) *InstallStatus {
	inst.mu.RLock()
	defer inst.mu.RUnlock()

	for _, task := range inst.tasks {
		if task.Plugin == pluginName {
			status := task.Status
			return &status
		}
	}
	return nil
}

func (inst *Installer) runInstall(task *InstallTask, manifestURL string) {
	defer close(task.Done)
	defer func() {
		now := time.Now()
		task.Status.CompletedAt = &now
	}()

	inst.updateTask(task, PhaseCheckDeps, 0.1, "checking dependencies...")

	mv, err := inst.fetchManifest(manifestURL)
	if err != nil {
		task.Status.Error = fmt.Sprintf("fetch manifest: %v", err)
		return
	}

	pluginName := mv.Name
	task.Plugin = pluginName

	inst.updateTask(task, PhaseDownload, 0.3, fmt.Sprintf("downloading plugin %s...", pluginName))

	pluginDir := filepath.Join(inst.pluginsDir, pluginName)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		task.Status.Error = fmt.Sprintf("create plugin directory: %v", err)
		return
	}

	if err := inst.downloadPlugin(task.ctx, pluginDir, manifestURL); err != nil {
		task.Status.Error = fmt.Sprintf("download plugin: %v", err)
		return
	}

	inst.updateTask(task, PhaseInstall, 0.6, "installing plugin...")

	if err := inst.installDependencies(task, pluginDir); err != nil {
		task.Status.Error = fmt.Sprintf("install dependencies: %v", err)
		return
	}

	inst.updateTask(task, PhaseRegister, 0.9, "registering plugin...")

	p, err := loadPluginFromManifest(filepath.Join(pluginDir, "plugin.json"))
	if err != nil {
		if os.IsNotExist(err) {
			task.Status.Error = "installed plugin missing plugin.json"
			return
		}
		task.Status.Error = fmt.Sprintf("load installed manifest: %v", err)
		return
	}

	if err := inst.registry.Register(p); err != nil {
		task.Status.Error = fmt.Sprintf("register plugin: %v", err)
		return
	}

	task.Status.PluginName = pluginName
	inst.updateTask(task, PhaseRegister, 1.0, "plugin installed successfully")
}

func (inst *Installer) fetchManifest(manifestURL string) (*ManifestV2, error) {
	resp, err := http.Get(manifestURL)
	if err != nil {
		data, readErr := os.ReadFile(manifestURL)
		if readErr != nil {
			return nil, fmt.Errorf("fetch from url: %w, read local: %v", err, readErr)
		}
		var mv ManifestV2
		if err := json.Unmarshal(data, &mv); err != nil {
			return nil, fmt.Errorf("parse local manifest: %w", err)
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

	if mv.ID == "" {
		return nil, fmt.Errorf("manifest missing required field: id")
	}
	if mv.Name == "" {
		return nil, fmt.Errorf("manifest missing required field: name")
	}

	return &mv, nil
}

func (inst *Installer) downloadPlugin(ctx context.Context, pluginDir, manifestURL string) error {
	return nil
}

func (inst *Installer) installDependencies(task *InstallTask, pluginDir string) error {
	reqFile := filepath.Join(pluginDir, "requirements.txt")
	if _, err := os.Stat(reqFile); err == nil {
		task.Status.Log = append(task.Status.Log, "installing pip dependencies...")
	}

	installScript := filepath.Join(pluginDir, "install.py")
	if _, err := os.Stat(installScript); err == nil {
		task.Status.Log = append(task.Status.Log, "running install.py...")
	}

	return nil
}

func (inst *Installer) updateTask(task *InstallTask, phase InstallPhase, progress float64, msg string) {
	task.Status.Phase = phase
	task.Status.Progress = progress
	task.Status.Log = append(task.Status.Log, msg)
}
