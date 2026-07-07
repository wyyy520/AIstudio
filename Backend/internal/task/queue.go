package task

import (
	"container/heap"
	"sync"
)

// TaskQueue is a priority queue for tasks.
// Higher priority tasks are dequeued first.
type TaskQueue struct {
	mu     sync.RWMutex
	items  priorityHeap
	waitCh chan struct{}
}

// priorityHeap implements heap.Interface for queue items.
type priorityHeap []*queueItem

type queueItem struct {
	task     *Task
	priority Priority
	index    int // index in the heap
}

// NewTaskQueue creates a new priority task queue.
func NewTaskQueue() *TaskQueue {
	q := &TaskQueue{
		items:  make(priorityHeap, 0),
		waitCh: make(chan struct{}, 1),
	}
	heap.Init(&q.items)
	return q
}

// Enqueue adds a task to the queue.
func (q *TaskQueue) Enqueue(task *Task) {
	q.mu.Lock()
	item := &queueItem{
		task:     task,
		priority: task.Priority,
	}
	heap.Push(&q.items, item)
	q.mu.Unlock()

	// Signal waiting workers
	select {
	case q.waitCh <- struct{}{}:
	default:
	}
}

// Dequeue removes and returns the highest priority task.
func (q *TaskQueue) Dequeue() *Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.items.Len() == 0 {
		return nil
	}
	item := heap.Pop(&q.items).(*queueItem)
	return item.task
}

// Peek returns the highest priority task without removing it.
func (q *TaskQueue) Peek() *Task {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.items.Len() == 0 {
		return nil
	}
	return q.items[0].task
}

// Size returns the number of tasks in the queue.
func (q *TaskQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.items.Len()
}

// Remove removes a task by ID from the queue.
func (q *TaskQueue) Remove(taskID string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, item := range q.items {
		if item.task.ID == taskID {
			heap.Remove(&q.items, i)
			return true
		}
	}
	return false
}

// WaitCh returns a channel that is signaled when a new task is added.
func (q *TaskQueue) WaitCh() <-chan struct{} {
	return q.waitCh
}

// --- heap.Interface implementation for priorityHeap ---

func (h priorityHeap) Len() int { return len(h) }

func (h priorityHeap) Less(i, j int) bool {
	// Higher priority items come first
	if h[i].priority != h[j].priority {
		return h[i].priority > h[j].priority
	}
	// For same priority, older items come first (FIFO)
	return h[i].task.CreatedAt.Before(h[j].task.CreatedAt)
}

func (h priorityHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *priorityHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*queueItem)
	item.index = n
	*h = append(*h, item)
}

func (h *priorityHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*h = old[0 : n-1]
	return item
}