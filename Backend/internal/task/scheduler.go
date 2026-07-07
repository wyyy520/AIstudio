package task

import (
	"context"
	"log"
	"sync"
	"time"
)

// Scheduler manages periodic task scheduling and maintenance.
type Scheduler struct {
	manager *Manager
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	interval time.Duration
}

// NewScheduler creates a new task scheduler.
func NewScheduler(manager *Manager) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		manager:  manager,
		ctx:      ctx,
		cancel:   cancel,
		interval: 30 * time.Second, // default maintenance interval
	}
}

// Start begins the scheduler's maintenance loop.
func (s *Scheduler) Start() {
	s.wg.Add(1)
	go s.run()
	log.Printf("[scheduler] started (interval: %v)", s.interval)
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
	log.Println("[scheduler] stopped")
}

func (s *Scheduler) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.maintenance()
		}
	}
}

// maintenance performs periodic checks:
// - Recovers stuck tasks (running too long)
// - Cleans up old completed tasks
func (s *Scheduler) maintenance() {
	s.manager.mu.Lock()
	defer s.manager.mu.Unlock()

	now := time.Now()
	timeout := 30 * time.Minute

	for id, task := range s.manager.tasks {
		if task.Status == StatusRunning && task.StartedAt != nil {
			if now.Sub(*task.StartedAt) > timeout {
				log.Printf("[scheduler] task %s timed out (running > %v), marking as failed", id, timeout)
				task.Status = StatusFailed
				task.Error = "task timed out"
				task.UpdatedAt = now
			}
		}

		// Clean up completed tasks older than 24 hours
		if IsTerminal(task.Status) && task.CompletedAt != nil {
			if now.Sub(*task.CompletedAt) > 24*time.Hour {
				log.Printf("[scheduler] cleaning up old task: %s", id)
				delete(s.manager.tasks, id)
			}
		}
	}
}