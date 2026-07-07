package task

import (
	"context"
	"testing"
	"time"
)

type testHandler struct{}

func (h *testHandler) Execute(ctx context.Context, task *Task) (interface{}, error) {
	return "done", nil
}

type failHandler struct{}

func (h *failHandler) Execute(ctx context.Context, task *Task) (interface{}, error) {
	return nil, ctx.Err()
}

func TestStateTransitions(t *testing.T) {
	tests := []struct {
		current Status
		next    Status
		valid   bool
	}{
		{StatusPending, StatusRunning, true},
		{StatusPending, StatusCancelled, true},
		{StatusPending, StatusSuccess, false},
		{StatusRunning, StatusSuccess, true},
		{StatusRunning, StatusFailed, true},
		{StatusRunning, StatusCancelled, true},
		{StatusRunning, StatusPending, false},
		{StatusSuccess, StatusRunning, false},
		{StatusFailed, StatusRunning, false},
		{StatusCancelled, StatusRunning, false},
	}

	for _, tc := range tests {
		err := ValidateTransition(tc.current, tc.next)
		if tc.valid && err != nil {
			t.Errorf("expected valid transition %s -> %s, got error: %v", tc.current, tc.next, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("expected invalid transition %s -> %s, but no error", tc.current, tc.next)
		}
	}
}

func TestTaskQueue(t *testing.T) {
	q := NewTaskQueue()

	// Push tasks with different priorities
	q.Enqueue(&Task{ID: "1", Priority: PriorityLow, CreatedAt: time.Now()})
	q.Enqueue(&Task{ID: "2", Priority: PriorityHigh, CreatedAt: time.Now()})
	q.Enqueue(&Task{ID: "3", Priority: PriorityUrgent, CreatedAt: time.Now()})
	q.Enqueue(&Task{ID: "4", Priority: PriorityNormal, CreatedAt: time.Now()})

	// Verify order: urgent first, then high, then normal, then low
	expected := []string{"3", "2", "4", "1"}
	for _, id := range expected {
		task := q.Dequeue()
		if task == nil {
			t.Fatal("expected task but got nil")
		}
		if task.ID != id {
			t.Errorf("expected task %s, got %s", id, task.ID)
		}
	}

	// Queue should be empty
	if task := q.Dequeue(); task != nil {
		t.Errorf("expected nil, got %s", task.ID)
	}
}

func TestManagerSubmitAndGet(t *testing.T) {
	m := NewManager(2)
	m.RegisterHandler("test", &testHandler{})
	m.Start()
	defer m.Stop()

	id, err := m.Submit(context.Background(), "test-task", "a test task", "test", PriorityNormal, nil)
	if err != nil {
		t.Fatalf("Submit() failed: %v", err)
	}

	if id == "" {
		t.Fatal("task ID should not be empty")
	}

	// Wait for completion
	time.Sleep(200 * time.Millisecond)

	task, err := m.GetTask(context.Background(), id)
	if err != nil {
		t.Fatalf("GetTask() failed: %v", err)
	}

	if task.Status != StatusSuccess {
		t.Errorf("expected success, got %s", task.Status)
	}
}

func TestManagerCancel(t *testing.T) {
	m := NewManager(2)
	m.RegisterHandler("test", &testHandler{})
	m.Start()
	defer m.Stop()

	id, err := m.Submit(context.Background(), "cancel-task", "cancel test", "test", PriorityNormal, nil)
	if err != nil {
		t.Fatalf("Submit() failed: %v", err)
	}

	// Cancel immediately
	if err := m.Cancel(context.Background(), id); err != nil {
		t.Fatalf("Cancel() failed: %v", err)
	}

	task, _ := m.GetTask(context.Background(), id)
	if task.Status != StatusCancelled {
		t.Errorf("expected cancelled, got %s", task.Status)
	}
}

func TestMissingHandler(t *testing.T) {
	m := NewManager(2)
	m.Start()
	defer m.Stop()

	_, err := m.Submit(context.Background(), "no-handler", "no handler test", "nonexistent", PriorityNormal, nil)
	if err == nil {
		t.Fatal("expected error for missing handler, got nil")
	}
}

func TestListTasks(t *testing.T) {
	m := NewManager(2)
	m.RegisterHandler("test", &testHandler{})
	m.Start()
	defer m.Stop()

	m.Submit(context.Background(), "t1", "", "test", PriorityLow, nil)
	m.Submit(context.Background(), "t2", "", "test", PriorityHigh, nil)

	time.Sleep(200 * time.Millisecond)

	tasks, err := m.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks() failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}