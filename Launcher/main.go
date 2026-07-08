package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aistudio/launcher/config"
	"github.com/aistudio/launcher/health"
	"github.com/aistudio/launcher/logger"
	"github.com/aistudio/launcher/process"
	"github.com/aistudio/launcher/service"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("       AIStudio Launcher v1.0.0         ")
	fmt.Println("========================================")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n[Launcher] Received shutdown signal")
		cancel()
	}()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[Launcher] Failed to load config: %v", err)
	}

	logDir := cfg.Paths.LogsDir
	if err := logger.Init(logDir); err != nil {
		log.Fatalf("[Launcher] Failed to initialize logger: %v", err)
	}
	logger.Info("Logger initialized", "log_dir", logDir)

	if err := checkDirectories(cfg); err != nil {
		logger.Error("Directory check failed", "error", err)
		log.Fatalf("[Launcher] Directory check failed: %v", err)
	}

	pm := process.NewProcessManager()

	launcher := service.NewLauncher(cfg, pm)

	if err := launcher.Start(ctx); err != nil {
		logger.Error("Failed to start services", "error", err)
		pm.KillAll()
		log.Fatalf("[Launcher] Failed to start: %v", err)
	}

	logger.Info("All services started successfully")
	fmt.Println("\n[Launcher] AIStudio is running...")
	fmt.Printf("[Launcher] Backend:   http://localhost:%s\n", cfg.Backend.Port)
	fmt.Printf("[Launcher] Frontend:  %s\n", cfg.Frontend.Path)
	fmt.Println("[Launcher] Press Ctrl+C to stop")

	health.StartMonitor(ctx, cfg, pm)

	<-ctx.Done()

	logger.Info("Shutting down...")
	launcher.Stop()
	logger.Info("Shutdown complete")
}

func checkDirectories(cfg *config.Config) error {
	dirs := []string{
		cfg.Paths.Root,
		cfg.Paths.BackendDir,
		cfg.Paths.FrontendDir,
		cfg.Paths.EngineDir,
		cfg.Paths.StorageDir,
		cfg.Paths.LogsDir,
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			logger.Warn("Directory does not exist, creating", "dir", dir)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}
	}

	return nil
}