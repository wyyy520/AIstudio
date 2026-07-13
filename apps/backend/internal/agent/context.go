package agent

import (
	"encoding/json"
	"sync"
	"time"
)

// ContextManager tracks the current session state for the Agent.
// It maintains the current project, task, conversation history, and user requirements.
type ContextManager struct {
	mu sync.RWMutex

	projectID    string
	userID       string
	goal         string
	history      []Message
	actionPlan   *ActionPlan
	executionLog []StepResult
	metadata     map[string]interface{}

	maxHistory int
}

// NewContextManager creates a new context manager.
func NewContextManager() *ContextManager {
	return &ContextManager{
		history:      make([]Message, 0),
		executionLog: make([]StepResult, 0),
		metadata:     make(map[string]interface{}),
		maxHistory:   50,
	}
}

// SetProject sets the current project context.
func (c *ContextManager) SetProject(projectID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.projectID = projectID
}

// ProjectID returns the current project ID.
func (c *ContextManager) ProjectID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.projectID
}

// SetUser sets the current user context.
func (c *ContextManager) SetUser(userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.userID = userID
}

// UserID returns the current user ID.
func (c *ContextManager) UserID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.userID
}

// SetGoal sets the current user goal.
func (c *ContextManager) SetGoal(goal string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.goal = goal
}

// Goal returns the current user goal.
func (c *ContextManager) Goal() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.goal
}

// AddMessage appends a message to the conversation history.
func (c *ContextManager) AddMessage(msg Message) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.history = append(c.history, msg)
	if len(c.history) > c.maxHistory {
		c.history = c.history[len(c.history)-c.maxHistory:]
	}
}

// GetHistory returns recent conversation history.
func (c *ContextManager) GetHistory(limit int) []Message {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if limit <= 0 || limit > len(c.history) {
		limit = len(c.history)
	}
	start := len(c.history) - limit
	if start < 0 {
		start = 0
	}
	result := make([]Message, len(c.history[start:]))
	copy(result, c.history[start:])
	return result
}

// SetActionPlan stores the current action plan.
func (c *ContextManager) SetActionPlan(plan *ActionPlan) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.actionPlan = plan
}

// ActionPlan returns the current action plan.
func (c *ContextManager) ActionPlan() *ActionPlan {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.actionPlan
}

// AddStepResult records a step execution result.
func (c *ContextManager) AddStepResult(result StepResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.executionLog = append(c.executionLog, result)
}

// StepResults returns all step execution results.
func (c *ContextManager) StepResults() []StepResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]StepResult, len(c.executionLog))
	copy(result, c.executionLog)
	return result
}

// SetMetadata stores arbitrary metadata.
func (c *ContextManager) SetMetadata(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metadata[key] = value
}

// GetMetadata retrieves metadata by key.
func (c *ContextManager) GetMetadata(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.metadata[key]
	return val, ok
}

// Reset clears the current context (except project/user).
func (c *ContextManager) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.goal = ""
	c.history = make([]Message, 0)
	c.actionPlan = nil
	c.executionLog = make([]StepResult, 0)
	c.metadata = make(map[string]interface{})
}

// Snapshot returns a copy of the current context state.
func (c *ContextManager) Snapshot() ContextSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return ContextSnapshot{
		ProjectID:  c.projectID,
		Goal:       c.goal,
		HistoryLen: len(c.history),
		StepsCount: len(c.executionLog),
		Timestamp:  time.Now(),
	}
}

// ContextSnapshot is a lightweight snapshot of the context state.
type ContextSnapshot struct {
	ProjectID  string    `json:"project_id"`
	Goal       string    `json:"goal"`
	HistoryLen int       `json:"history_len"`
	StepsCount int       `json:"steps_count"`
	Timestamp  time.Time `json:"timestamp"`
}

// ---- Project Context Types ----

// ProjectContext holds project-aware context for the agent
type ProjectContext struct {
	ProjectID    string            `json:"project_id"`
	Name         string            `json:"name"`
	Workflows    []WorkflowSummary `json:"workflows,omitempty"`
	RecentLogs   []LogEntry        `json:"recent_logs,omitempty"`
	RecentErrors []ErrorEntry      `json:"recent_errors,omitempty"`
	ActiveTasks  []TaskSummary     `json:"active_tasks,omitempty"`
}

type WorkflowSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Source    string `json:"source,omitempty"`
}

type ErrorEntry struct {
	Timestamp string `json:"timestamp"`
	Error     string `json:"error"`
	NodeID    string `json:"node_id,omitempty"`
}

type TaskSummary struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Name   string `json:"name,omitempty"`
}

func (c *ContextManager) SetProjectContext(pc *ProjectContext) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metadata["project_context"] = pc
}

func (c *ContextManager) GetProjectContext() *ProjectContext {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if pc, ok := c.metadata["project_context"]; ok {
		if projectCtx, ok := pc.(*ProjectContext); ok {
			return projectCtx
		}
	}
	return nil
}

// ---- Action Plan Types ----

// ActionPlan is the structured plan generated by the Agent.
type ActionPlan struct {
	Goal        string   `json:"goal"`
	Explanation string   `json:"explanation"`
	Steps       []Action `json:"steps"`
}

// Action represents a single step in the plan.
type Action struct {
	Tool   string                 `json:"tool"`
	Params map[string]interface{} `json:"params"`
	Reason string                 `json:"reason"`
}

// StepResult records the result of executing a single action.
type StepResult struct {
	Step      int                    `json:"step"`
	Tool      string                 `json:"tool"`
	Success   bool                   `json:"success"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// ---- Response Types ----

// AgentResponse is the top-level response from the Agent.
type AgentResponse struct {
	Goal        string       `json:"goal"`
	Explanation string       `json:"explanation"`
	Plan        []Action     `json:"plan"`
	Steps       []StepResult `json:"steps"`
	Status      string       `json:"status"` // "planned", "executing", "completed", "failed"
	Summary     string       `json:"summary"`
}

// ToJSON serializes the response to JSON bytes.
func (r *AgentResponse) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}