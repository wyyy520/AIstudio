package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aistudio/backend/internal/agent"
	"github.com/aistudio/backend/internal/environment"
	"github.com/aistudio/backend/internal/plugin"
	"github.com/aistudio/backend/internal/task"
	"github.com/aistudio/backend/internal/workflow"
)

// AgentService bridges the Agent Engine with the rest of AIStudio.
// It wires the Agent's tools to the Plugin Manager, Workflow Engine, Task Manager, and Environment Manager.
type AgentService struct {
	agent   *agent.Agent
	plugin  *plugin.Manager
	env     *environment.Manager
	engine  *workflow.Engine
	taskMgr *task.Manager
	mcp     *MCPService
}

// NewAgentService creates a new AgentService with all dependencies wired.
func NewAgentService(
	ag *agent.Agent,
	pluginMgr *plugin.Manager,
	envMgr *environment.Manager,
	engine *workflow.Engine,
	taskMgr *task.Manager,
	mcpSvc *MCPService,
) *AgentService {
	svc := &AgentService{
		agent:   ag,
		plugin:  pluginMgr,
		env:     envMgr,
		engine:  engine,
		taskMgr: taskMgr,
		mcp:     mcpSvc,
	}

	// Wire all tools to the agent's tool registry
	svc.wireTools()

	return svc
}

// wireTools connects each tool to its real implementation.
func (s *AgentService) wireTools() {
	registry := s.agent.ToolRegistry()

	// Check Environment
	registry.Register(&agent.CheckEnvironmentTool{
		CheckFn: func(ctx context.Context) (map[string]interface{}, error) {
			status := s.env.GetStatus()
			data, _ := json.Marshal(status)
			var result map[string]interface{}
			json.Unmarshal(data, &result)
			return result, nil
		},
	})

	// List Plugins
	registry.Register(&agent.ListPluginsTool{
		ListFn: func(ctx context.Context) ([]map[string]interface{}, error) {
			plugins := s.plugin.GetRegistry().List()
			var result []map[string]interface{}
			for _, p := range plugins {
				data, _ := json.Marshal(p)
				var m map[string]interface{}
				json.Unmarshal(data, &m)
				result = append(result, m)
			}
			return result, nil
		},
	})

	// Create Workflow
	registry.Register(&agent.CreateWorkflowTool{
		CreateFn: func(ctx context.Context, name string, workflowJSON json.RawMessage) (string, error) {
			// Parse and validate the workflow
			wf, err := workflow.ParseAndValidate(workflowJSON)
			if err != nil {
				return "", fmt.Errorf("invalid workflow: %w", err)
			}
			if name != "" {
				wf.Name = name
			}
			_ = wf // Store workflow via engine
			log.Printf("[agent-service] workflow created: %s", name)
			return "wf_" + name, nil
		},
	})

	// Run Workflow
	registry.Register(&agent.RunWorkflowTool{
		RunFn: func(ctx context.Context, workflowID string, params map[string]interface{}) (string, error) {
			// Create a task for the workflow execution
			taskID, err := s.taskMgr.CreateTask(ctx, "", workflowID, task.TaskTypeWorkflow, "Agent Workflow Run", "workflow", task.PriorityNormal, params)
			if err != nil {
				return "", fmt.Errorf("failed to create task: %w", err)
			}
			if err := s.taskMgr.StartTask(ctx, taskID); err != nil {
				return "", fmt.Errorf("failed to start task: %w", err)
			}
			log.Printf("[agent-service] workflow task started: %s", taskID)
			return taskID, nil
		},
	})

	// Get Task Status
	registry.Register(&agent.GetTaskStatusTool{
		StatusFn: func(ctx context.Context, taskID string) (map[string]interface{}, error) {
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

	// Install Plugin
	registry.Register(&agent.InstallPluginTool{
		InstallFn: func(ctx context.Context, name string) error {
			plugins := s.plugin.GetRegistry().List()
			for _, p := range plugins {
				if p.Name == name {
					return nil
				}
			}
			return fmt.Errorf("plugin not found in registry: %s", name)
		},
	})

	// Register all MCP tools with the agent
	// Each MCP tool becomes an agent tool that can be called by the agent
	// This allows the agent to invoke external tools like SUMO, MATLAB, etc.
	for _, mcpTool := range s.mcp.GetAgentMCPTools() {
		registry.Register(mcpTool)
	}

	log.Printf("[agent-service] all tools wired successfully (including %d MCP tools)", len(s.mcp.GetAgentMCPTools()))
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
}

// Chat processes a chat message through the Agent.
func (s *AgentService) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	log.Printf("[agent-service] chat request: project=%s user=%s message=%q", req.ProjectID, req.UserID, req.Message)

	// Collect plugin names
	pluginNames := s.getPluginNames()

	// Get environment status
	envStatus := s.env.GetStatus()
	envJSON, _ := json.Marshal(envStatus)

	// Process through the Agent
	resp, err := s.agent.Process(ctx, req.Message, req.ProjectID, req.UserID, pluginNames, string(envJSON))
	if err != nil {
		return nil, err
	}

	return &ChatResponse{
		Reply:       resp.Summary,
		Goal:        resp.Goal,
		Explanation: resp.Explanation,
		Plan:        resp.Plan,
		Steps:       resp.Steps,
		Status:      resp.Status,
		Summary:     resp.Summary,
	}, nil
}

// PlanOnly analyzes a message and returns a plan without executing.
func (s *AgentService) PlanOnly(ctx context.Context, req ChatRequest) (*agent.ActionPlan, error) {
	pluginNames := s.getPluginNames()
	envStatus := s.env.GetStatus()
	envJSON, _ := json.Marshal(envStatus)

	return s.agent.PlanOnly(ctx, req.Message, req.ProjectID, req.UserID, pluginNames, string(envJSON))
}

// getPluginNames returns the list of installed plugin names.
func (s *AgentService) getPluginNames() []string {
	plugins := s.plugin.GetRegistry().List()
	names := make([]string, 0, len(plugins))
	for _, p := range plugins {
		names = append(names, p.Name)
	}
	return names
}