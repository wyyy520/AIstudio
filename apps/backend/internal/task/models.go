package task

import "time"

// Priority defines task execution priority.
type Priority int

const (
	PriorityLow    Priority = 0
	PriorityNormal Priority = 1
	PriorityHigh   Priority = 2
	PriorityUrgent Priority = 3
)

func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	case PriorityUrgent:
		return "urgent"
	default:
		return "unknown"
	}
}

// Status defines task lifecycle status.
type Status string

const (
	StatusWaiting   Status = "waiting"
	StatusRunning   Status = "running"
	StatusSuccess   Status = "success"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

// TaskType defines the category of work a task represents.
type TaskType string

const (
	TaskTypeWorkflow TaskType = "workflow"
	TaskTypeAgent    TaskType = "agent"
	TaskTypePlugin   TaskType = "plugin"
)

// Task represents a unit of async work in the AIStudio system.
// It maps to the task-schema.md protocol definition.
type Task struct {
	ID          string      `json:"id"`
	ProjectID   string      `json:"projectId"`
	WorkflowID  string      `json:"workflowId"`
	Type        TaskType    `json:"type"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Status      Status      `json:"status"`
	Progress    float64     `json:"progress"`
	Priority    Priority    `json:"priority"`
	Handler     string      `json:"handler"`
	Payload     interface{} `json:"payload,omitempty"`
	Result      interface{} `json:"result,omitempty"`
	Error       string      `json:"error,omitempty"`
	StartTime   *time.Time  `json:"startedAt,omitempty"`
	EndTime     *time.Time  `json:"completedAt,omitempty"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

// TaskResult holds the result of a completed task.
type TaskResult struct {
	TaskID string      `json:"id"`
	Status Status      `json:"status"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// TaskStatusResponse is the API response for GET /api/task/:id/status
type TaskStatusResponse struct {
	TaskID   string  `json:"id"`
	Status   Status  `json:"status"`
	Progress float64 `json:"progress"`
	Error    string  `json:"error,omitempty"`
}