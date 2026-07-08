package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aistudio/launcher/config"
	"github.com/aistudio/launcher/logger"
	"github.com/aistudio/launcher/process"
)

type HealthChecker struct {
	cfg *config.Config
	pm  *process.ProcessManager
}

func StartMonitor(ctx context.Context, cfg *config.Config, pm *process.ProcessManager) {
	hc := &HealthChecker{
		cfg: cfg,
		pm:  pm,
	}

	interval := time.Duration(cfg.Launcher.HealthCheckInterval) * time.Second
	if interval < 1*time.Second {
		interval = 5 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Info("Health monitor started", "interval", interval)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Health monitor stopped")
			return
		case <-ticker.C:
			hc.check()
		}
	}
}

func (hc *HealthChecker) check() {
	hc.checkBackend()
	hc.checkEngine()
}

func (hc *HealthChecker) checkBackend() {
	healthURL := fmt.Sprintf("http://localhost:%s/health", hc.cfg.Backend.Port)

	running, err := hc.pm.Check("backend")
	if err != nil || !running {
		logger.Warn("Backend process not running")
		if hc.cfg.Launcher.AutoRestart {
			logger.Info("Attempting to restart Backend...")
		}
		return
	}

	if !CheckHTTP(healthURL) {
		logger.Warn("Backend health check failed")
		if hc.cfg.Launcher.AutoRestart {
			logger.Info("Attempting to restart Backend...")
		}
		return
	}

	logger.Debug("Backend health check passed")
}

func (hc *HealthChecker) checkEngine() {
	healthURL := fmt.Sprintf("http://localhost:%s/health", hc.cfg.Engine.Port)

	running, err := hc.pm.Check("engine")
	if err != nil || !running {
		logger.Warn("Engine process not running")
		if hc.cfg.Launcher.AutoRestart {
			logger.Info("Attempting to restart Engine...")
		}
		return
	}

	if !CheckHTTP(healthURL) {
		logger.Warn("Engine health check failed")
		if hc.cfg.Launcher.AutoRestart {
			logger.Info("Attempting to restart Engine...")
		}
		return
	}

	logger.Debug("Engine health check passed")
}

func CheckHTTP(url string) bool {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}