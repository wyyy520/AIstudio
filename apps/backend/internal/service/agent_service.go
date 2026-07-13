// Package service provides business logic services for AIStudio.
//
// AgentService bridges the Agent Engine with the rest of AIStudio.
// It wires the Agent's tools to the Skill Manager, Compiler, and Runtime.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aistudio/backend/internal/agent"
	"github.com/aistudio/backend/internal/skill"
	"github.com/aistudio/backend/internal/task"
	"github.com/aistudio/backend/internal/workflow"
)

// AgentService bridges the Agent Engine with the rest of AIStudio.
type AgentService struct {
	agent   *agent.Agent
	skill   *skill.Manager
	taskMgr *task.Manager
	mcp     *MCPService
	llm     agent.LLMProvider
}

// NewAgentService creates a new AgentService with all dependencies wired.
func NewAgentService(skillMgr *skill.Manager) *AgentService {
	return &AgentService{
		skill: skillMgr,
	}
}

// WithAgent sets the Agent instance.
func (s *AgentService) WithAgent(ag *agent.Agent) *AgentService {
	s.agent = ag
	return s
}

// WithTaskManager sets the Task Manager instance.
func (s *AgentService) WithTaskManager(tm *task.Manager) *AgentService {
	s.taskMgr = tm
	return s
}

// WithMCPService sets the MCP Service instance.
func (s *AgentService) WithMCPService(mcp *MCPService) *AgentService {
	s.mcp = mcp
	return s
}

// WithLLMProvider sets the LLM provider.
func (s *AgentService) WithLLMProvider(llm agent.LLMProvider) *AgentService {
	s.llm = llm
	return s
}

// WireTools connects all agent tools to their real implementations.
func (s *AgentService) WireTools() {
	if s.agent == nil {
		log.Println("[agent-service] warning: agent not set, skipping tool wiring")
		return
	}

	registry := s.agent.ToolRegistry()

	// List Skills
	registry.Register(&agent.ListSkillsTool{
		ListFn: func(ctx context.Context) ([]map[string]interface{}, error) {
			skills := s.skill.List()
			result := make([]map[string]interface{}, 0, len(skills))
			for _, sk := range skills {
				params := sk.Parameters()
				paramNames := make([]string, 0, len(params))
				for _, p := range params {
					paramNames = append(paramNames, p.Name)
				}
				result = append(result, map[string]interface{}{
					"id":          sk.ID(),
					"name":        sk.Name(),
					"category":    sk.Category(),
					"description": sk.Description(),
					"version":     sk.Version(),
					"parameters":  paramNames,
				})
			}
			return result, nil
		},
	})

	// Apply Skill
	registry.Register(&agent.ApplySkillTool{
		ApplyFn: func(ctx context.Context, skillID string, params map[string]interface{}) (string, error) {
			wf, err := s.skill.Apply(skillID, params)
			if err != nil {
				return "", fmt.Errorf("failed to apply skill: %w", err)
			}
			wfJSON, _ := json.Marshal(wf)
			log.Printf("[agent-service] skill applied: %s (workflow: %s)", skillID, wf.ID)
			return string(wfJSON), nil
		},
	})

	// Create Workflow
	registry.Register(&agent.CreateWorkflowTool{
		CreateFn: func(ctx context.Context, name string, workflowJSON json.RawMessage) (string, error) {
			wf, err := workflow.Parse(workflowJSON)
			if err != nil {
				return "", fmt.Errorf("invalid workflow: %w", err)
			}
			if name != "" {
				wf.Name = name
			}
			if wf.ID == "" {
				wf.ID = fmt.Sprintf("wf_%d", time.Now().UnixMilli())
			}
			log.Printf("[agent-service] workflow created: %s (id: %s)", name, wf.ID)
			return wf.ID, nil
		},
	})

	// Get Task Status
	registry.Register(&agent.GetTaskStatusTool{
		StatusFn: func(ctx context.Context, taskID string) (map[string]interface{}, error) {
			if s.taskMgr == nil {
				return nil, fmt.Errorf("task manager not available")
			}
			t, err := s.taskMgr.GetTask(ctx, taskID)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"id":       t.ID,
				"status":   t.Status,
				"progress": t.Progress,
				"type":     t.Type,
			}, nil
		},
	})

	// Register MCP tools if available
	if s.mcp != nil {
		for _, mcpTool := range s.mcp.GetAgentMCPTools() {
			registry.Register(mcpTool)
		}
	}

	log.Printf("[agent-service] all tools wired successfully")
}

// ChatRequest represents an agent chat request.
type ChatRequest struct {
	Message   string                 `json:"message"`
	ProjectID string                 `json:"projectId"`
	UserID    string                 `json:"userId"`
	Context   map[string]interface{} `json:"context"`
}

// ChatResponse represents an agent chat response.
type ChatResponse struct {
	Reply       string              `json:"reply"`
	Goal        string              `json:"goal"`
	Explanation string              `json:"explanation"`
	Plan        []agent.Action      `json:"plan"`
	Steps       []agent.StepResult  `json:"steps"`
	Status      string              `json:"status"`
	Summary     string              `json:"summary"`
	WorkflowID  string              `json:"workflow_id,omitempty"`
	TaskID      string              `json:"task_id,omitempty"`
	Workflow    json.RawMessage     `json:"workflow,omitempty"`
}

// Chat processes a chat message through the Agent.
func (s *AgentService) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	log.Printf("[agent-service] chat request: project=%s user=%s message=%q", req.ProjectID, req.UserID, req.Message)

	if s.agent == nil {
		return nil, fmt.Errorf("agent not initialized")
	}

	// Get available skills
	skills := s.skill.List()
	skillNames := make([]string, 0, len(skills))
	for _, sk := range skills {
		skillNames = append(skillNames, sk.Name())
	}

	// Process through the Agent
	resp, err := s.agent.Process(ctx, req.Message, req.ProjectID, req.UserID, skillNames, "")
	if err != nil {
		return nil, err
	}

	chatResp := &ChatResponse{
		Reply:       resp.Summary,
		Goal:        resp.Goal,
		Explanation: resp.Explanation,
		Plan:        resp.Plan,
		Steps:       resp.Steps,
		Status:      resp.Status,
		Summary:     resp.Summary,
	}

	// Extract workflow_id and task_id from step results
	for _, step := range resp.Steps {
		if step.Success && step.Data != nil {
			if wfID, ok := step.Data["workflow_id"].(string); ok {
				chatResp.WorkflowID = wfID
			}
			if taskID, ok := step.Data["task_id"].(string); ok {
				chatResp.TaskID = taskID
			}
		}
	}

	return chatResp, nil
}

// WorkflowGenResult contains the result of generating a workflow from natural language.
type WorkflowGenResult struct {
	WorkflowJSON map[string]interface{} `json:"workflow_json"`
	Nodes        int                    `json:"nodes"`
	Edges        int                    `json:"edges"`
	Valid        bool                   `json:"valid"`
	Error        string                 `json:"error,omitempty"`
}

// GenerateWorkflowFromNL generates a workflow from a natural language description.
func (s *AgentService) GenerateWorkflowFromNL(ctx context.Context, projectID string, naturalLanguage string) (*WorkflowGenResult, error) {
	if s.llm == nil {
		return nil, fmt.Errorf("LLM provider not available")
	}

	prompt := fmt.Sprintf(`You are an AI workflow generator for AIStudio. 
Available node types and their purposes:
- data_loader/data_preprocessor/data_augmentation/data_split: Data processing
- model_trainer/model_evaluator/model_exporter/model_inference: Model operations
- feature_extractor/hyperparameter_tuning: Feature engineering
- visualization/metric_computation: Analysis
- control.condition/control.loop/control.switch/control.retry: Logic control
- vision.yolo_train/vision.yolo_inference/vision.resnet/vision.efficientnet: Computer vision
- nlp.transformer/nlp.llm/nlp.bert/nlp.lstm: Natural language processing
- speech.asr/speech.tts: Speech processing
- deployment.api_server/deployment.docker: Deployment
- system.python_env/system.install_dep: System

Generate a workflow JSON for the following request: %s

Return ONLY valid JSON matching the workflow schema with "nodes" and "edges" arrays.`, naturalLanguage)

	resp, err := s.llm.Chat(agent.ChatContext{ProjectID: projectID}, []agent.Message{{Role: "user", Content: prompt}})
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	var wfJSON map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &wfJSON); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	nodes, _ := wfJSON["nodes"].([]interface{})
	edges, _ := wfJSON["edges"].([]interface{})

	result := &WorkflowGenResult{
		WorkflowJSON: wfJSON,
		Nodes:        len(nodes),
		Edges:        len(edges),
	}

	if result.Nodes == 0 {
		result.Error = "no nodes generated"
		return result, nil
	}

	raw, _ := json.Marshal(wfJSON)
	wf, err := workflow.Parse(raw)
	if err != nil {
		result.Error = fmt.Sprintf("validation failed: %v", err)
		return result, nil
	}

	valid := workflow.Validate(wf)
	result.Valid = valid.Valid
	if !valid.Valid && len(valid.Errors) > 0 {
		result.Error = valid.Errors[0].Error()
	}

	return result, nil
}

// PlanOnly analyzes a message and returns a plan without executing.
func (s *AgentService) PlanOnly(ctx context.Context, req ChatRequest) (*agent.ActionPlan, error) {
	if s.agent == nil {
		return nil, fmt.Errorf("agent not initialized")
	}
	skills := s.skill.List()
	skillNames := make([]string, 0, len(skills))
	for _, sk := range skills {
		skillNames = append(skillNames, sk.Name())
	}
	return s.agent.PlanOnly(ctx, req.Message, req.ProjectID, req.UserID, skillNames, "")
}

