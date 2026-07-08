package task

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type testHandler struct{}

func (h *testHandler) Execute(ctx context.Context, task *Task) (interface{}, error) {
	return "done", nil
}

type failHandler struct{}

func (h *failHandler) Execute(ctx context.Context, task *Task) (interface{}, error) {
	return nil, fmt.Errorf("simulated task failure")
}

type progressHandler struct {
	onProgress func(taskID string, progress float64)
}

func (h *progressHandler) Execute(ctx context.Context, task *Task) (interface{}, error) {
	if h.onProgress != nil {
		h.onProgress(task.ID, 0.5)
	}
	return "done", nil
}

func TestStateTransitions(t *testing.T) {
	tests := []struct {
		current Status
		next    Status
		valid   bool
	}{
		{StatusWaiting, StatusRunning, true},
		{StatusWaiting, StatusCancelled, true},
		{StatusWaiting, StatusSuccess, false},
		{StatusRunning, StatusSuccess, true},
		{StatusRunning, StatusFailed, true},
		{StatusRunning, StatusCancelled, true},
		{StatusRunning, StatusWaiting, false},
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

	q.Enqueue(&Task{ID: "1", Priority: PriorityLow, CreatedAt: time.Now()})
	q.Enqueue(&Task{ID: "2", Priority: PriorityHigh, CreatedAt: time.Now()})
	q.Enqueue(&Task{ID: "3", Priority: PriorityUrgent, CreatedAt: time.Now()})
	q.Enqueue(&Task{ID: "4", Priority: PriorityNormal, CreatedAt: time.Now()})

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

	if task := q.Dequeue(); task != nil {
		t.Errorf("expected nil, got %s", task.ID)
	}
}

func TestManagerCreateAndStartTask(t *testing.T) {
	m := NewManager(2)
	m.RegisterHandler("test", &testHandler{})
	m.Start()
	defer m.Stop()

	taskID, err := m.CreateTask(context.Background(), "proj-1", "wf-1", TaskTypeWorkflow, "test-task", "test", PriorityNormal, nil)
	if err != nil {
		t.Fatalf("CreateTask() failed: %v", err)
	}

	if taskID == "" {
		t.Fatal("task ID should not be empty")
	}

	// Task should be in waiting state
	task, err := m.GetTask(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetTask() failed: %v", err)
	}
	if task.Status != StatusWaiting {
		t.Errorf("expected waiting, got %s", task.Status)
	}

	// Start the task
	if err := m.StartTask(context.Background(), taskID); err != nil {
		t.Fatalf("StartTask() failed: %v", err)
	}

	// Wait for completion
	time.Sleep(300 * time.Millisecond)

	task, err = m.GetTask(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetTask() failed: %v", err)
	}
	if task.Status != StatusSuccess {
		t.Errorf("expected success, got %s", task.Status)
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

	time.Sleep(300 * time.Millisecond)

	task, err := m.GetTask(context.Background(), id)
	if err != nil {
		t.Fatalf("GetTask() failed: %v", err)
	}

	if task.Status != StatusSuccess {
		t.Errorf("expected success, got %s", task.Status)
	}
}

func TestManagerUpdateProgress(t *testing.T) {
	m := NewManager(2)
	m.RegisterHandler("test", &testHandler{})
	m.Start()
	defer m.Stop()

	taskID, err := m.CreateTask(context.Background(), "proj-1", "wf-1", TaskTypeWorkflow, "progress-test", "test", PriorityNormal, nil)
	if err != nil {
		t.Fatalf("CreateTask() failed: %v", err)
	}

	if err := m.StartTask(context.Background(), taskID); err != nil {
		t.Fatalf("StartTask() failed: %v", err)
	}

	// Update progress
	if err := m.UpdateProgress(context.Background(), taskID, 0.5); err != nil {
		t.Fatalf("UpdateProgress() failed: %v", err)
	}

	task, _ := m.GetTask(context.Background(), taskID)
	if task.Progress != 0.5 {
		t.Errorf("expected progress 0.5, got %f", task.Progress)
	}
}

func TestManagerFinishTask(t *testing.T) {
	m := NewManager(2)
	m.RegisterHandler("test", &testHandler{})
	m.Start()
	defer m.Stop()

	taskID, err := m.CreateTask(context.Background(), "proj-1", "wf-1", TaskTypeWorkflow, "finish-test", "test", PriorityNormal, nil)
	if err != nil {
		t.Fatalf("CreateTask() failed: %v", err)
	}

	if err := m.StartTask(context.Background(), taskID); err != nil {
		t.Fatalf("StartTask() failed: %v", err)
	}

	// Wait a bit for the worker to pick it up
	time.Sleep(300 * time.Millisecond)

	task, _ := m.GetTask(context.Background(), taskID)
	if task.Status != StatusSuccess {
		t.Errorf("expected success, got %s", task.Status)
	}
	if task.Progress != 1.0 {
		t.Errorf("expected progress 1.0, got %f", task.Progress)
	}
}

func TestManagerFailTask(t *testing.T) {
	m := NewManager(2)
	m.RegisterHandler("fail", &failHandler{})
	m.Start()
	defer m.Stop()

	taskID, err := m.CreateTask(context.Background(), "proj-1", "wf-1", TaskTypeWorkflow, "fail-test", "fail", PriorityNormal, nil)
	if err != nil {
		t.Fatalf("CreateTask() failed: %v", err)
	}

	if err := m.StartTask(context.Background(), taskID); err != nil {
		t.Fatalf("StartTask() failed: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	task, _ := m.GetTask(context.Background(), taskID)
	if task.Status != StatusFailed {
		t.Errorf("expected failed, got %s", task.Status)
	}
}

func TestManagerCancel(t *testing.T) {
	m := NewManager(2)
	m.RegisterHandler("test", &testHandler{})
	m.Start()
	defer m.Stop()

	taskID, err := m.CreateTask(context.Background(), "proj-1", "wf-1", TaskTypeWorkflow, "cancel-test", "test", PriorityNormal, nil)
	if err != nil {
		t.Fatalf("CreateTask() failed: %v", err)
	}

	// Cancel before starting
	if err := m.Cancel(context.Background(), taskID); err != nil {
		t.Fatalf("Cancel() failed: %v", err)
	}

	task, _ := m.GetTask(context.Background(), taskID)
	if task.Status != StatusCancelled {
		t.Errorf("expected cancelled, got %s", task.Status)
	}
}

func TestManagerCancelRunning(t *testing.T) {
	m := NewManager(2)
	m.RegisterHandler("test", &testHandler{})
	m.Start()
	defer m.Stop()

	taskID, err := m.CreateTask(context.Background(), "proj-1", "wf-1", TaskTypeWorkflow, "cancel-running", "test", PriorityNormal, nil)
	if err != nil {
		t.Fatalf("CreateTask() failed: %v", err)
	}

	if err := m.StartTask(context.Background(), taskID); err != nil {
		t.Fatalf("StartTask() failed: %v", err)
	}

	// Cancel while running
	if err := m.Cancel(context.Background(), taskID); err != nil {
		t.Fatalf("Cancel() failed: %v", err)
	}

	task, _ := m.GetTask(context.Background(), taskID)
	if task.Status != StatusCancelled {
		t.Errorf("expected cancelled, got %s", task.Status)
	}
}

func TestMissingHandler(t *testing.T) {
	m := NewManager(2)
	m.Start()
	defer m.Stop()

	_, err := m.CreateTask(context.Background(), "proj-1", "wf-1", TaskTypeWorkflow, "no-handler", "nonexistent", PriorityNormal, nil)
	if err == nil {
		t.Fatal("expected error for missing handler, got nil")
	}
}

func TestListTasks(t *testing.T) {
	m := NewManager(2)
	m.RegisterHandler("test", &testHandler{})
	m.Start()
	defer m.Stop()

	id1, _ := m.Submit(context.Background(), "t1", "", "test", PriorityLow, nil)
	id2, _ := m.Submit(context.Background(), "t2", "", "test", PriorityHigh, nil)

	_ = id1
	_ = id2

	time.Sleep(300 * time.Millisecond)

	tasks, err := m.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks() failed: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestTaskStatusResponse(t *testing.T) {
	m := NewManager(2)
	m.RegisterHandler("test", &testHandler{})
	m.Start()
	defer m.Stop()

	taskID, err := m.CreateTask(context.Background(), "proj-1", "wf-1", TaskTypeWorkflow, "status-test", "test", PriorityNormal, nil)
	if err != nil {
		t.Fatalf("CreateTask() failed: %v", err)
	}

	statusResp, err := m.GetTaskStatus(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetTaskStatus() failed: %v", err)
	}

	if statusResp.TaskID != taskID {
		t.Errorf("expected task_id %s, got %s", taskID, statusResp.TaskID)
	}
	if statusResp.Status != StatusWaiting {
		t.Errorf("expected waiting, got %s", statusResp.Status)
	}
	if statusResp.Progress != 0 {
		t.Errorf("expected progress 0, got %f", statusResp.Progress)
	}
}

func TestEventBus(t *testing.T) {
	bus := NewEventBus()

	received := make(chan *TaskEvent, 1)
	bus.Subscribe(EventTaskCreated, func(event *TaskEvent) {
		received <- event
	})

	task := &Task{
		ID:       "test-task-id",
		Status:   StatusWaiting,
		Progress: 0,
	}

	bus.EmitTaskCreated(task)

	select {
	case event := <-received:
		if event.TaskID != "test-task-id" {
			t.Errorf("expected task_id test-task-id, got %s", event.TaskID)
		}
		if event.Type != EventTaskCreated {
			t.Errorf("expected event type %s, got %s", EventTaskCreated, event.Type)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestEventBusAllEvents(t *testing.T) {
	bus := NewEventBus()

	received := make(chan TaskEventType, 1)
	bus.SubscribeAll(func(event *TaskEvent) {
		received <- event.Type
	})

	task := &Task{
		ID:       "test-task-id",
		Status:   StatusWaiting,
		Progress: 0,
	}

	bus.EmitTaskCreated(task)
	select {
	case eventType := <-received:
		if eventType != EventTaskCreated {
			t.Errorf("expected %s, got %s", EventTaskCreated, eventType)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestPriorityString(t *testing.T) {
	tests := []struct {
		p        Priority
		expected string
	}{
		{PriorityLow, "low"},
		{PriorityNormal, "normal"},
		{PriorityHigh, "high"},
		{PriorityUrgent, "urgent"},
		{Priority(99), "unknown"},
	}

	for _, tc := range tests {
		if got := tc.p.String(); got != tc.expected {
			t.Errorf("Priority(%d).String() = %s, want %s", tc.p, got, tc.expected)
		}
	}
}

func TestIsTerminal(t *testing.T) {
	if IsTerminal(StatusWaiting) {
		t.Error("waiting should not be terminal")
	}
	if IsTerminal(StatusRunning) {
		t.Error("running should not be terminal")
	}
	if !IsTerminal(StatusSuccess) {
		t.Error("success should be terminal")
	}
	if !IsTerminal(StatusFailed) {
		t.Error("failed should be terminal")
	}
	if !IsTerminal(StatusCancelled) {
		t.Error("cancelled should be terminal")
	}
}

func TestIsActive(t *testing.T) {
	if !IsActive(StatusWaiting) {
		t.Error("waiting should be active")
	}
	if !IsActive(StatusRunning) {
		t.Error("running should be active")
	}
	if IsActive(StatusSuccess) {
		t.Error("success should not be active")
	}
	if IsActive(StatusFailed) {
		t.Error("failed should not be active")
	}
	if IsActive(StatusCancelled) {
		t.Error("cancelled should not be active")
	}
}