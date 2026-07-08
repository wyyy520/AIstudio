package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/aistudio/launcher/config"
	"github.com/aistudio/launcher/health"
	"github.com/aistudio/launcher/logger"
	"github.com/aistudio/launcher/process"
)

type Launcher struct {
	cfg *config.Config
	pm  *process.ProcessManager
}

func NewLauncher(cfg *config.Config, pm *process.ProcessManager) *Launcher {
	return &Launcher{
		cfg: cfg,
		pm:  pm,
	}
}

func (l *Launcher) Start(ctx context.Context) error {
	logger.Info("Starting AIStudio services...")

	if err := l.startBackend(ctx); err != nil {
		return fmt.Errorf("failed to start backend: %w", err)
	}

	if err := l.waitForBackendHealth(ctx); err != nil {
		return fmt.Errorf("backend health check failed: %w", err)
	}

	if err := l.startEngine(ctx); err != nil {
		return fmt.Errorf("failed to start engine: %w", err)
	}

	if err := l.waitForEngineHealth(ctx); err != nil {
		return fmt.Errorf("engine health check failed: %w", err)
	}

	if err := l.startFrontend(ctx); err != nil {
		return fmt.Errorf("failed to start frontend: %w", err)
	}

	return nil
}

func (l *Launcher) Stop() {
	logger.Info("Stopping AIStudio services...")

	l.pm.Stop("frontend", 5*time.Second)
	l.pm.Stop("engine", 10*time.Second)
	l.pm.Stop("backend", 10*time.Second)
}

func (l *Launcher) startBackend(ctx context.Context) error {
	logger.Info("Starting Backend service...")

	backendPath := filepath.Join(l.cfg.Paths.BackendDir, "cmd.exe")
	if _, err := os.Stat(backendPath); os.IsNotExist(err) {
		backendPath = filepath.Join(l.cfg.Paths.BackendDir, "cmd")
	}

	backendLog := filepath.Join(l.cfg.Paths.LogsDir, "backend.log")

	cmd := exec.CommandContext(ctx, backendPath)
	cmd.Dir = l.cfg.Paths.BackendDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("AISTUDIO_CONFIG=%s", filepath.Join(l.cfg.Paths.ConfigDir, "backend.yaml")),
	)

	if err := l.pm.Start("backend", cmd, backendLog); err != nil {
		return err
	}

	logger.Info("Backend started", "path", backendPath)
	return nil
}

func (l *Launcher) waitForBackendHealth(ctx context.Context) error {
	logger.Info("Waiting for Backend health...")

	healthURL := fmt.Sprintf("http://localhost:%s/health", l.cfg.Backend.Port)
	timeout := time.Duration(l.cfg.Launcher.StartupTimeout) * time.Second

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ticker.C:
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				if health.CheckHTTP(healthURL) {
					logger.Info("Backend is healthy")
					return nil
				}
			case <-time.After(timeout):
				return fmt.Errorf("backend health check timeout after %v", timeout)
			}
		}
	}
}

func (l *Launcher) startEngine(ctx context.Context) error {
	logger.Info("Starting Python Engine service...")

	pythonPath := l.cfg.Engine.PythonPath
	if _, err := exec.LookPath(pythonPath); err != nil {
		logger.Warn("Python not found in PATH, using default", "python_path", pythonPath)
		pythonPath = "python"
	}

	runnerPath := filepath.Join(l.cfg.Paths.Root, l.cfg.Engine.RunnerScript)
	engineLog := filepath.Join(l.cfg.Paths.LogsDir, "engine.log")

	cmd := exec.CommandContext(ctx, pythonPath, runnerPath)
	cmd.Dir = filepath.Join(l.cfg.Paths.Root, l.cfg.Paths.EngineDir)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("PYTHONPATH=%s", filepath.Join(l.cfg.Paths.Root, l.cfg.Paths.EngineDir)),
		"PYTHONUNBUFFERED=1",
	)

	if err := l.pm.Start("engine", cmd, engineLog); err != nil {
		return err
	}

	logger.Info("Engine started", "python", pythonPath, "runner", runnerPath)
	return nil
}

func (l *Launcher) waitForEngineHealth(ctx context.Context) error {
	logger.Info("Waiting for Engine health...")

	healthURL := fmt.Sprintf("http://localhost:%s/health", l.cfg.Engine.Port)
	timeout := time.Duration(l.cfg.Launcher.StartupTimeout) * time.Second

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ticker.C:
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				if health.CheckHTTP(healthURL) {
					logger.Info("Engine is healthy")
					return nil
				}
			case <-time.After(timeout):
				logger.Warn("Engine health check timeout, continuing anyway")
				return nil
			}
		}
	}
}

func (l *Launcher) startFrontend(ctx context.Context) error {
	logger.Info("Starting Frontend service...")

	frontendPath := l.cfg.Frontend.Path
	if frontendPath == "" {
		frontendPath = filepath.Join(l.cfg.Paths.FrontendDir, "dist", "index.html")
	}

	tauriExe := filepath.Join(l.cfg.Paths.FrontendDir, "Frontend.exe")
	if _, err := os.Stat(tauriExe); err != nil {
		tauriExe = findTauriExe(l.cfg.Paths.FrontendDir)
	}

	frontendLog := filepath.Join(l.cfg.Paths.LogsDir, "frontend.log")

	var cmd *exec.Cmd
	if _, err := os.Stat(tauriExe); err == nil {
		cmd = exec.CommandContext(ctx, tauriExe)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("VITE_API_BASE_URL=%s", l.cfg.Frontend.BackendAddr),
		)
	} else {
		cmd = exec.CommandContext(ctx, "cmd", "/c", "start", "", frontendPath)
	}

	cmd.Dir = l.cfg.Paths.FrontendDir

	if err := l.pm.Start("frontend", cmd, frontendLog); err != nil {
		return err
	}

	logger.Info("Frontend started", "path", frontendPath)
	return nil
}

func findTauriExe(dir string) string {
	exePath := filepath.Join(dir, "Frontend.exe")
	if _, err := os.Stat(exePath); err == nil {
		return exePath
	}

	releasePath := filepath.Join(dir, "target", "release", "Frontend.exe")
	if _, err := os.Stat(releasePath); err == nil {
		return releasePath
	}

	debugPath := filepath.Join(dir, "target", "debug", "Frontend.exe")
	if _, err := os.Stat(debugPath); err == nil {
		return debugPath
	}

	return ""
}

func (l *Launcher) RestartBackend(ctx context.Context) error {
	logger.Info("Restarting Backend...")

	backendPath := filepath.Join(l.cfg.Paths.BackendDir, "cmd.exe")
	backendLog := filepath.Join(l.cfg.Paths.LogsDir, "backend.log")

	cmd := exec.CommandContext(ctx, backendPath)
	cmd.Dir = l.cfg.Paths.BackendDir

	return l.pm.Restart("backend", cmd, backendLog, 10*time.Second)
}

func (l *Launcher) RestartEngine(ctx context.Context) error {
	logger.Info("Restarting Engine...")

	pythonPath := l.cfg.Engine.PythonPath
	runnerPath := filepath.Join(l.cfg.Paths.Root, l.cfg.Engine.RunnerScript)
	engineLog := filepath.Join(l.cfg.Paths.LogsDir, "engine.log")

	cmd := exec.CommandContext(ctx, pythonPath, runnerPath)
	cmd.Dir = filepath.Join(l.cfg.Paths.Root, l.cfg.Paths.EngineDir)

	return l.pm.Restart("engine", cmd, engineLog, 10*time.Second)
}

func (l *Launcher) RestartFrontend(ctx context.Context) error {
	logger.Info("Restarting Frontend...")

	return l.startFrontend(ctx)
}

func CheckHTTP(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}