package service

import (
	"context"
	"log"
	"time"

	"github.com/aistudio/backend/internal/task"
)

// AgentService handles AI agent chat and task delegation.
// This is a placeholder that will be extended with actual Agent Engine integration.
type AgentService struct {
	taskMgr *task.Manager
}

// NewAgentService creates a new AgentService.
func NewAgentService(taskMgr *task.Manager) *AgentService {
	return &AgentService{taskMgr: taskMgr}
}

// ChatRequest represents an agent chat request.
type ChatRequest struct {
	Message   string                 `json:"message"`
	ProjectID string                 `json:"projectId"`
	Context   map[string]interface{} `json:"context"`
}

// ChatResponse represents an agent chat response.
type ChatResponse struct {
	Reply   string                   `json:"reply"`
	Actions []AgentAction            `json:"actions"`
	Tools   []string                 `json:"tools"`
}

// AgentAction represents an action the agent suggests.
type AgentAction struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// Chat processes a chat message and returns an agent response.
// For now, it provides a rule-based reply and optionally creates tasks.
// In the future, this will connect to the Agent Engine / LLM.
func (s *AgentService) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	log.Printf("[agent] chat request: project=%s, message=%q", req.ProjectID, req.Message)

	// Simple intent detection (placeholder for LLM integration)
	reply, actions := s.processIntent(req.Message, req.ProjectID)

	// For actionable intents, create tasks
	for i, action := range actions {
		taskID, err := s.taskMgr.Submit(ctx,
			action.Type,
			action.Description,
			"agent",
			task.PriorityNormal,
			action.Parameters,
		)
		if err != nil {
			log.Printf("[agent] failed to create task for action %s: %v", action.Type, err)
			continue
		}
		log.Printf("[agent] created task %s for action: %s", taskID, action.Type)
		actions[i].Parameters["taskId"] = taskID
	}

	resp := &ChatResponse{
		Reply:   reply,
		Actions: actions,
		Tools:   []string{"workflow", "plugin", "task"},
	}

	log.Printf("[agent] chat response: reply=%q, actions=%d", resp.Reply, len(resp.Actions))
	return resp, nil
}

// processIntent does simple rule-based intent detection.
// This is a placeholder; replace with LLM/Agent engine later.
func (s *AgentService) processIntent(message, projectID string) (string, []AgentAction) {
	// Simple keyword matching
	switch {
	case containsAny(message, "run", "execute", "start", "deploy"):
		return "I'll help you run the workflow. Creating a task now.", []AgentAction{
			{Type: "run_workflow", Description: "Execute workflow", Parameters: map[string]interface{}{"projectId": projectID, "timestamp": time.Now().Unix()}},
		}

	case containsAny(message, "create", "new", "make", "build"):
		return "I'll create a new workflow for you.", []AgentAction{
			{Type: "create_workflow", Description: "Create new workflow", Parameters: map[string]interface{}{"projectId": projectID}},
		}

	case containsAny(message, "status", "progress", "check"):
		return "Checking the status of your tasks...", []AgentAction{
			{Type: "list_tasks", Description: "List all tasks", Parameters: map[string]interface{}{"projectId": projectID}},
		}

	case containsAny(message, "help", "what can you do", "capabilities"):
		return `I can help you with:
1. Run workflows and monitor their progress
2. Create new workflows from descriptions
3. Check task and workflow status
4. Install and manage plugins
5. Execute plugins with custom input

Just tell me what you'd like to do!`, []AgentAction{}

	case containsAny(message, "plugin", "install"):
		return "I can help you manage plugins. Which plugin would you like to install?", []AgentAction{
			{Type: "list_plugins", Description: "List available plugins", Parameters: map[string]interface{}{}},
		}

	default:
		return "I understand you need assistance. I can run workflows, check status, or manage plugins. Could you be more specific about what you'd like to do?", []AgentAction{}
	}
}

// containsAny checks if a string contains any of the given substrings.
func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if contains(s, sub) {
			return true
		}
	}
	return false
}

// contains is a simple case-insensitive substring check.
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		containsSubstring(toLower(s), toLower(substr))
}

// toLower converts a string to lowercase without importing strings.
func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

// containsSubstring checks if s contains substr.
func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Ensure ChatResponse implements a reasonable interface.
var _ = (*ChatResponse)(nil)