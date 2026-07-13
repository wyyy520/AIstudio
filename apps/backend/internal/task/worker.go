package task

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// Worker represents a single worker that processes tasks from the queue.
type Worker struct {
	id       int
	queue    *TaskQueue
	handlers map[string]TaskHandler
	manager  *Manager
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewWorker creates a new worker.
func NewWorker(id int, queue *TaskQueue) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		id:       id,
		queue:    queue,
		handlers: make(map[string]TaskHandler),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// SetManager sets the task manager reference for lifecycle callbacks.
func (w *Worker) SetManager(mgr *Manager) {
	w.manager = mgr
}

// RegisterHandler registers a task handler for a specific task type.
func (w *Worker) RegisterHandler(name string, handler TaskHandler) {
	w.handlers[name] = handler
}

// Start begins processing tasks in a goroutine.
func (w *Worker) Start() {
	w.wg.Add(1)
	go w.run()
	log.Printf("[worker-%d] started", w.id)
}

// Stop gracefully stops the worker.
func (w *Worker) Stop() {
	w.cancel()
	w.wg.Wait()
	log.Printf("[worker-%d] stopped", w.id)
}

func (w *Worker) run() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-w.queue.WaitCh():
			w.processNext()
		}
	}
}

func (w *Worker) processNext() {
	task := w.queue.Dequeue()
	if task == nil {
		return
	}

	log.Printf("[worker-%d] processing task: %s (handler: %s)", w.id, task.ID, task.Handler)

	// Find handler
	handler, ok := w.handlers[task.Handler]
	if !ok {
		errMsg := fmt.Sprintf("no handler registered for: %s", task.Handler)
		log.Printf("[worker-%d] %s", w.id, errMsg)
		if w.manager != nil {
			_ = w.manager.FailTask(w.ctx, task.ID, errMsg)
		}
		return
	}

	// Execute
	result, err := handler.Execute(w.ctx, task)
	if err != nil {
		log.Printf("[worker-%d] task %s failed: %v", w.id, task.ID, err)
		if w.manager != nil {
			_ = w.manager.FailTask(w.ctx, task.ID, err.Error())
		}
	} else {
		log.Printf("[worker-%d] task %s completed successfully", w.id, task.ID)
		if w.manager != nil {
			_ = w.manager.FinishTask(w.ctx, task.ID, result)
		}
	}
}

// WorkerPool manages a pool of workers.
type WorkerPool struct {
	workers []*Worker
	queue   *TaskQueue
	mu      sync.RWMutex
}

// NewWorkerPool creates a pool with the specified number of workers.
func NewWorkerPool(numWorkers int, queue *TaskQueue) *WorkerPool {
	pool := &WorkerPool{
		workers: make([]*Worker, 0, numWorkers),
		queue:   queue,
	}

	for i := 0; i < numWorkers; i++ {
		worker := NewWorker(i, queue)
		pool.workers = append(pool.workers, worker)
	}

	return pool
}

// SetManager sets the task manager reference on all workers.
func (p *WorkerPool) SetManager(mgr *Manager) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, w := range p.workers {
		w.SetManager(mgr)
	}
}

// RegisterHandler registers a handler on all workers.
func (p *WorkerPool) RegisterHandler(name string, handler TaskHandler) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, w := range p.workers {
		w.RegisterHandler(name, handler)
	}
}

// Start starts all workers.
func (p *WorkerPool) Start() {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, w := range p.workers {
		w.Start()
	}
	log.Printf("[worker-pool] all %d workers started", len(p.workers))
}

// Stop gracefully stops all workers.
func (p *WorkerPool) Stop() {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, w := range p.workers {
		w.Stop()
	}
	log.Printf("[worker-pool] all %d workers stopped", len(p.workers))
}