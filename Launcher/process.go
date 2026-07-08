package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/aistudio/launcher/logger"
)

type Process struct {
	Name   string
	Cmd    *exec.Cmd
	Status string
	PID    int
}

type ProcessManager struct {
	mu       sync.RWMutex
	processes map[string]*Process
}

func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		processes: make(map[string]*Process),
	}
}

func (pm *ProcessManager) Start(name string, cmd *exec.Cmd, logFile string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.processes[name]; exists {
		return fmt.Errorf("process %s already running", name)
	}

	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		cmd.Stdout = f
		cmd.Stderr = f
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	pm.processes[name] = &Process{
		Name:   name,
		Cmd:    cmd,
		Status: "running",
		PID:    cmd.Process.Pid,
	}

	logger.Info("Process started", "name", name, "pid", cmd.Process.Pid)
	return nil
}

func (pm *ProcessManager) Stop(name string, timeout time.Duration) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p, exists := pm.processes[name]
	if !exists {
		return fmt.Errorf("process %s not found", name)
	}

	if p.Cmd.Process == nil {
		return fmt.Errorf("process %s has no process handle", name)
	}

	logger.Info("Stopping process", "name", name, "pid", p.PID)

	p.Status = "stopping"

	done := make(chan struct{})
	go func() {
		p.Cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Process stopped gracefully", "name", name)
	case <-time.After(timeout):
		logger.Warn("Process stop timeout, killing", "name", name)
		p.Cmd.Process.Kill()
		p.Status = "killed"
	}

	delete(pm.processes, name)
	return nil
}

func (pm *ProcessManager) Restart(name string, newCmd *exec.Cmd, logFile string, timeout time.Duration) error {
	if err := pm.Stop(name, timeout); err != nil {
		logger.Warn("Failed to stop process during restart", "name", name, "error", err)
	}
	return pm.Start(name, newCmd, logFile)
}

func (pm *ProcessManager) Check(name string) (bool, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	p, exists := pm.processes[name]
	if !exists {
		return false, nil
	}

	if p.Cmd.Process == nil {
		return false, nil
	}

	err := p.Cmd.Process.Signal(syscall.Signal(0))
	if err != nil {
		delete(pm.processes, name)
		return false, nil
	}

	return true, nil
}

func (pm *ProcessManager) GetPID(name string) (int, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	p, exists := pm.processes[name]
	if !exists {
		return 0, false
	}
	return p.PID, true
}

func (pm *ProcessManager) KillAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for name, p := range pm.processes {
		if p.Cmd.Process != nil {
			logger.Info("Killing process", "name", name, "pid", p.PID)
			p.Cmd.Process.Kill()
		}
		delete(pm.processes, name)
	}
}

func (pm *ProcessManager) List() map[string]int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]int)
	for name, p := range pm.processes {
		result[name] = p.PID
	}
	return result
}