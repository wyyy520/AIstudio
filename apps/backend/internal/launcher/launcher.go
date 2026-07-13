package launcher

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// ModuleStatus represents the current state of a running module.
type ModuleStatus string

const (
	StatusIdle     ModuleStatus = "idle"
	StatusStarting ModuleStatus = "starting"
	StatusRunning  ModuleStatus = "running"
	StatusStopping ModuleStatus = "stopping"
	StatusStopped  ModuleStatus = "stopped"
	StatusError    ModuleStatus = "error"
)

// ModuleInfo describes a managed module.
type ModuleInfo struct {
	Name    string       `json:"name"`
	Status  ModuleStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	PID     int          `json:"pid"`
}

// StartFunc is a function that starts a module.
type StartFunc func(ctx context.Context) error

// StopFunc is a function that stops a module.
type StopFunc func(ctx context.Context) error

// Module represents a managed component.
type Module struct {
	Info    ModuleInfo
	StartFn StartFunc
	StopFn  StopFunc
	order   int
}

// Launcher manages the lifecycle of all modules.
type Launcher struct {
	mu        sync.RWMutex
	modules   []*Module
	started   bool
	startTime time.Time
}

// NewLauncher creates a new Launcher.
func NewLauncher() *Launcher {
	return &Launcher{
		modules: make([]*Module, 0),
	}
}

// Register adds a module to the launcher with start/stop callbacks.
// Modules are started in registration order and stopped in reverse order.
func (l *Launcher) Register(name string, order int, startFn StartFunc, stopFn StopFunc) {
	l.mu.Lock()
	defer l.mu.Unlock()

	m := &Module{
		Info: ModuleInfo{
			Name:   name,
			Status: StatusIdle,
		},
		StartFn: startFn,
		StopFn:  stopFn,
		order:   order,
	}

	// Insert in order
	idx := 0
	for i, mod := range l.modules {
		if mod.order > order {
			idx = i
			break
		}
		idx = i + 1
	}
	if idx == len(l.modules) {
		l.modules = append(l.modules, m)
	} else {
		l.modules = append(l.modules[:idx], append([]*Module{m}, l.modules[idx:]...)...)
	}
}

// Start starts all registered modules in order.
func (l *Launcher) Start(ctx context.Context) error {
	l.mu.Lock()
	if l.started {
		l.mu.Unlock()
		return fmt.Errorf("launcher already started")
	}
	l.started = true
	l.startTime = time.Now()
	l.mu.Unlock()

	log.Println("[launcher] starting all modules...")

	for _, m := range l.modules {
		log.Printf("[launcher] starting module: %s", m.Info.Name)
		m.Info.Status = StatusStarting

		if err := l.startModule(ctx, m); err != nil {
			m.Info.Status = StatusError
			m.Info.Message = err.Error()
			log.Printf("[launcher] module %s failed to start: %v", m.Info.Name, err)
			return fmt.Errorf("failed to start %s: %w", m.Info.Name, err)
		}

		m.Info.Status = StatusRunning
		log.Printf("[launcher] module %s started successfully", m.Info.Name)
	}

	log.Println("[launcher] all modules started")
	return nil
}

func (l *Launcher) startModule(ctx context.Context, m *Module) error {
	startCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- m.StartFn(startCtx)
	}()

	select {
	case err := <-errCh:
		return err
	case <-startCtx.Done():
		return fmt.Errorf("start timeout after 30s")
	}
}

// Stop stops all modules in reverse order.
func (l *Launcher) Stop(ctx context.Context) {
	l.mu.Lock()
	if !l.started {
		l.mu.Unlock()
		return
	}
	l.started = false
	l.mu.Unlock()

	log.Println("[launcher] stopping all modules...")

	// Stop in reverse order
	for i := len(l.modules) - 1; i >= 0; i-- {
		m := l.modules[i]
		log.Printf("[launcher] stopping module: %s", m.Info.Name)
		m.Info.Status = StatusStopping

		stopCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if m.StopFn != nil {
			done := make(chan struct{})
			go func() {
				m.StopFn(stopCtx)
				close(done)
			}()

			select {
			case <-done:
				log.Printf("[launcher] module %s stopped gracefully", m.Info.Name)
			case <-stopCtx.Done():
				log.Printf("[launcher] module %s stop timeout, force killed", m.Info.Name)
			}
		}

		m.Info.Status = StatusStopped
	}

	log.Println("[launcher] all modules stopped")
}

// Status returns the status of all registered modules.
func (l *Launcher) Status() []ModuleInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]ModuleInfo, 0, len(l.modules))
	for _, m := range l.modules {
		result = append(result, m.Info)
	}
	return result
}

// HealthCheck checks all modules and returns their status.
// Returns true if all modules are running.
func (l *Launcher) HealthCheck() (bool, []ModuleInfo) {
	status := l.Status()
	allRunning := true
	for _, m := range status {
		if m.Status != StatusRunning {
			allRunning = false
		}
	}
	return allRunning, status
}

// Uptime returns the duration since the launcher started.
func (l *Launcher) Uptime() time.Duration {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if !l.started {
		return 0
	}
	return time.Since(l.startTime)
}

// IsRunning returns whether the launcher is in the started state.
func (l *Launcher) IsRunning() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.started
}