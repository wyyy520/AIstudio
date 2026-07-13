package e2e

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/aistudio/backend/internal/task"
)

func TestTaskLifecycle(t *testing.T) {
	m := task.NewManager(2)
	m.RegisterHandler("test", &testHandler{})
	m.Start()
	defer m.Stop()

	taskID, err := m.CreateTask(context.Background(), "proj-1", "wf-1", task.TaskTypeWorkflow, "lifecycle-test", "test", task.PriorityNormal, nil)
	if err != nil {
		t.Fatalf("CreateTask() failed: %v", err)
	}

	if taskID == "" {
		t.Fatal("task ID should not be empty")
	}

	created, err := m.GetTask(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetTask() failed: %v", err)
	}
	if created.Status != task.StatusWaiting {
		t.Errorf("expected waiting, got %s", created.Status)
	}

	if err := m.StartTask(context.Background(), taskID); err != nil {
		t.Fatalf("StartTask() failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	completed, err := m.GetTask(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetTask() failed: %v", err)
	}
	if completed.Status != task.StatusSuccess {
		t.Errorf("expected success, got %s (error: %s)", completed.Status, completed.Error)
	}
	if completed.Progress != 1.0 {
		t.Errorf("expected progress 1.0, got %f", completed.Progress)
	}
}

func TestTaskFailure(t *testing.T) {
	m := task.NewManager(2)
	m.RegisterHandler("fail", &failHandler{})
	m.Start()
	defer m.Stop()

	taskID, err := m.Submit(context.Background(), "fail-test", "a task that fails", "fail", task.PriorityNormal, nil)
	if err != nil {
		t.Fatalf("Submit() failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	result, err := m.GetTask(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetTask() failed: %v", err)
	}
	if result.Status != task.StatusFailed {
		t.Errorf("expected failed, got %s", result.Status)
	}
	if result.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestTaskCancellation(t *testing.T) {
	m := task.NewManager(2)
	m.RegisterHandler("test", &testHandler{})
	m.Start()
	defer m.Stop()

	taskID, err := m.CreateTask(context.Background(), "proj-1", "wf-1", task.TaskTypeWorkflow, "cancel-test", "test", task.PriorityNormal, nil)
	if err != nil {
		t.Fatalf("CreateTask() failed: %v", err)
	}

	if err := m.Cancel(context.Background(), taskID); err != nil {
		t.Fatalf("Cancel() failed: %v", err)
	}

	cancelled, err := m.GetTask(context.Background(), taskID)
	if err != nil {
		t.Fatalf("GetTask() failed: %v", err)
	}
	if cancelled.Status != task.StatusCancelled {
		t.Errorf("expected cancelled, got %s", cancelled.Status)
	}
}

func TestTaskStatusResponse(t *testing.T) {
	m := task.NewManager(2)
	m.RegisterHandler("test", &testHandler{})
	m.Start()
	defer m.Stop()

	taskID, err := m.CreateTask(context.Background(), "p", "w", task.TaskTypeWorkflow, "status-test", "test", task.PriorityNormal, nil)
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
	if statusResp.Status != task.StatusWaiting {
		t.Errorf("expected waiting, got %s", statusResp.Status)
	}
	if statusResp.Progress != 0 {
		t.Errorf("expected progress 0, got %f", statusResp.Progress)
	}
}

func TestTaskEventOnCompletion(t *testing.T) {
	m := task.NewManager(2)
	m.RegisterHandler("test", &testHandler{})

	var mu sync.Mutex
	events := make([]task.TaskEventType, 0)
	m.EventBus().SubscribeAll(func(event *task.TaskEvent) {
		mu.Lock()
		events = append(events, event.Type)
		mu.Unlock()
	})

	m.Start()
	defer m.Stop()

	_, err := m.Submit(context.Background(), "event-test", "", "test", task.PriorityNormal, nil)
	if err != nil {
		t.Fatalf("Submit() failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
}

type testHandler struct{}

func (h *testHandler) Execute(ctx context.Context, tsk *task.Task) (interface{}, error) {
	return "done", nil
}

type failHandler struct{}

func (h *failHandler) Execute(ctx context.Context, tsk *task.Task) (interface{}, error) {
	return nil, nil
}
