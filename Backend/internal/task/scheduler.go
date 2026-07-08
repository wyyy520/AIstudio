package task

import (
	"context"
	"log"
	"sync"
	"time"
)

// Scheduler manages periodic task scheduling and maintenance.
// It detects waiting tasks, starts them, and performs cleanup of old tasks.
type Scheduler struct {
	manager  *Manager
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	interval time.Duration
}

// NewScheduler creates a new task scheduler.
func NewScheduler(manager *Manager) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		manager:  manager,
		ctx:      ctx,
		cancel:   cancel,
		interval: 5 * time.Second, // check for waiting tasks every 5 seconds
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
			s.detectAndStartWaitingTasks()
			s.maintenance()
		}
	}
}

// detectAndStartWaitingTasks finds tasks in waiting status and starts them.
func (s *Scheduler) detectAndStartWaitingTasks() {
	s.manager.mu.Lock()
	defer s.manager.mu.Unlock()

	for id, task := range s.manager.tasks {
		if task.Status == StatusWaiting {
			// Validate the transition first
			if err := ValidateTransition(task.Status, StatusRunning); err != nil {
				continue
			}

			task.Status = StatusRunning
			now := time.Now()
			task.StartTime = &now
			task.UpdatedAt = now

			// Persist status change
			if s.manager.repo != nil {
				if err := s.manager.repo.Update(task); err != nil {
					log.Printf("[scheduler] failed to persist task %s start: %v", id, err)
				}
			}

			// Enqueue for execution
			s.manager.queue.Enqueue(task)

			// Emit event
			s.manager.events.EmitTaskStarted(task)

			log.Printf("[scheduler] detected waiting task, started: %s", id)
		}
	}
}

// maintenance performs periodic checks:
// - Recovers stuck tasks (running too long)
// - Cleans up old completed tasks
// - Reserved for future CPU/GPU resource scheduling
func (s *Scheduler) maintenance() {
	s.manager.mu.Lock()
	defer s.manager.mu.Unlock()

	now := time.Now()
	timeout := 30 * time.Minute

	for id, task := range s.manager.tasks {
		if task.Status == StatusRunning && task.StartTime != nil {
			if now.Sub(*task.StartTime) > timeout {
				log.Printf("[scheduler] task %s timed out (running > %v), marking as failed", id, timeout)
				task.Status = StatusFailed
				task.Error = "task timed out"
				task.UpdatedAt = now
				task.EndTime = &now

				if s.manager.repo != nil {
					_ = s.manager.repo.Update(task)
				}

				s.manager.events.EmitTaskFailed(task)
			}
		}

		// Clean up completed tasks older than 24 hours
		if IsTerminal(task.Status) && task.EndTime != nil {
			if now.Sub(*task.EndTime) > 24*time.Hour {
				log.Printf("[scheduler] cleaning up old task: %s", id)
				delete(s.manager.tasks, id)
				if s.manager.repo != nil {
					_ = s.manager.repo.Delete(id)
				}
			}
		}
	}
}