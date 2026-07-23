package runtime

import (
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// ProcessManager manages running processes with resource monitoring.
// Per silu.md 5.7, it handles process lifecycle and monitoring.
type ProcessManager struct {
	mu            sync.RWMutex
	managed       map[string]*ManagedProcess
	maxConcurrent int
}

// ManagedProcess represents a managed running process.
type ManagedProcess struct {
	RunID      string
	Cmd        *exec.Cmd
	Status     RunStatusEnum
	StartedAt  time.Time
	PID        int
	MemoryMB   int64
	CPUPercent float64
}

// NewProcessManager creates a new ProcessManager.
func NewProcessManager(maxConcurrent int) *ProcessManager {
	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}
	return &ProcessManager{
		managed:       make(map[string]*ManagedProcess),
		maxConcurrent: maxConcurrent,
	}
}

// Track registers a process for management.
func (pm *ProcessManager) Track(runID string, cmd *exec.Cmd) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.managed) >= pm.maxConcurrent {
		return fmt.Errorf("max concurrent processes reached (%d)", pm.maxConcurrent)
	}

	pid := 0
	if cmd.Process != nil {
		pid = cmd.Process.Pid
	}

	pm.managed[runID] = &ManagedProcess{
		RunID:     runID,
		Cmd:       cmd,
		Status:    RunStatusRunning,
		StartedAt: time.Now(),
		PID:       pid,
	}
	return nil
}

// UpdateStatus updates the status of a managed process.
func (pm *ProcessManager) UpdateStatus(runID string, status RunStatusEnum) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if p, ok := pm.managed[runID]; ok {
		p.Status = status
	}
}

// Get returns a managed process by run ID.
func (pm *ProcessManager) Get(runID string) (*ManagedProcess, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	p, ok := pm.managed[runID]
	return p, ok
}

// Remove stops tracking a process without killing it.
func (pm *ProcessManager) Remove(runID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.managed, runID)
}

// ListRunning returns all managed processes.
func (pm *ProcessManager) ListRunning() []*ManagedProcess {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	result := make([]*ManagedProcess, 0, len(pm.managed))
	for _, p := range pm.managed {
		result = append(result, p)
	}
	return result
}

// Count returns the number of managed processes.
func (pm *ProcessManager) Count() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.managed)
}
