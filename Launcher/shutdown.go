package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/aistudio/launcher/logger"
	"github.com/aistudio/launcher/process"
)

type ShutdownManager struct {
	pm          *process.ProcessManager
	shutdownCtx context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

func NewShutdownManager(pm *process.ProcessManager) *ShutdownManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ShutdownManager{
		pm:          pm,
		shutdownCtx: ctx,
		cancel:      cancel,
	}
}

func (sm *ShutdownManager) Start() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("Received shutdown signal")
		sm.Shutdown()
	}()
}

func (sm *ShutdownManager) Shutdown() {
	logger.Info("Starting shutdown sequence...")
	sm.cancel()

	sm.shutdownOrder()
}

func (sm *ShutdownManager) shutdownOrder() {
	order := []string{"frontend", "engine", "backend"}
	timeouts := map[string]time.Duration{
		"frontend": 5 * time.Second,
		"engine":   10 * time.Second,
		"backend":  10 * time.Second,
	}

	for _, name := range order {
		logger.Info("Stopping service", "service", name)
		if err := sm.pm.Stop(name, timeouts[name]); err != nil {
			logger.Warn("Failed to stop service", "service", name, "error", err)
		}
	}

	sm.pm.KillAll()
	logger.Info("All services stopped")

	logger.Close()
}

func GracefulShutdown(pm *process.ProcessManager, timeout time.Duration) error {
	logger.Info("Starting graceful shutdown...")

	order := []string{"frontend", "engine", "backend"}
	timeouts := map[string]time.Duration{
		"frontend": timeout,
		"engine":   timeout,
		"backend":  timeout,
	}

	var errs []error
	for _, name := range order {
		logger.Info("Stopping service", "service", name)
		if err := pm.Stop(name, timeouts[name]); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
			logger.Warn("Failed to stop service", "service", name, "error", err)
		}
	}

	pm.KillAll()
	logger.Close()

	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	logger.Info("Graceful shutdown complete")
	return nil
}