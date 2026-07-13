package agent

import (
	"time"

	"github.com/aistudio/packages/workflow"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	ToolCall  *ToolCall `json:"tool_call,omitempty"`
}

type Conversation struct {
	ID        string              `json:"id"`
	Messages  []Message           `json:"messages"`
	Workflow  *workflow.Workflow  `json:"workflow,omitempty"`
	Context   map[string]any      `json:"context,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type ToolCall struct {
	Tool      string            `json:"tool"`
	Arguments map[string]any    `json:"arguments"`
	Result    *ToolCallResult   `json:"result,omitempty"`
}

type ToolCallResult struct {
	Success bool                   `json:"success"`
	Data    map[string]any         `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type AgentConfig struct {
	Provider    string  `json:"provider"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	Timeout     string  `json:"timeout"`
	APIKey      string  `json:"-"`

	MaxHistory       int  `json:"max_history"`
	EnableStreaming  bool `json:"enable_streaming"`
	EnableRollback   bool `json:"enable_rollback"`
}

type WorkflowGenerationRequest struct {
	Description string            `json:"description"`
	Target      workflow.Target   `json:"target"`
	Constraints map[string]any    `json:"constraints,omitempty"`
	SkillID     string            `json:"skill_id,omitempty"`
}

type Plan struct {
	Goal        string       `json:"goal"`
	Explanation string       `json:"explanation"`
	Steps       []PlanStep   `json:"steps"`
}

type PlanStep struct {
	Action  string         `json:"action"`
	Params  map[string]any `json:"params"`
	Reason  string         `json:"reason"`
}

type StepResult struct {
	Step      int       `json:"step"`
	Action    string    `json:"action"`
	Success   bool      `json:"success"`
	Data      any       `json:"data,omitempty"`
	Error     string    `json:"error,omitempty"`
	Duration  string    `json:"duration,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}