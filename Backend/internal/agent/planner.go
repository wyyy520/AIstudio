package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// Planner analyzes user requirements and generates action plans.
// It uses the LLM to understand intent and plan tool calls.
type Planner struct {
	llm       LLMProvider
	builder   *PromptBuilder
	memory    *Memory
}

// NewPlanner creates a new Planner.
func NewPlanner(llm LLMProvider, memory *Memory) *Planner {
	return &Planner{
		llm:     llm,
		builder: NewPromptBuilder(),
		memory:  memory,
	}
}

// Plan analyzes the user's request and produces an action plan.
// It uses the LLM for intent understanding and tool selection.
func (p *Planner) Plan(ctx ChatContext, userMessage string, tools []ToolInfo, plugins []string, envStatus string) (*ActionPlan, error) {
	log.Printf("[planner] analyzing request: %q", userMessage)

	p.builder.
		SetTools(tools).
		SetPlugins(plugins).
		SetEnvironment(envStatus).
		SetProjectID(ctx.ProjectID)

	// Try LLM-based planning first
	if p.llm != nil {
		plan, err := p.planWithLLM(ctx, userMessage)
		if err == nil && plan != nil && len(plan.Steps) > 0 {
			log.Printf("[planner] LLM plan generated: %d steps", len(plan.Steps))
			return plan, nil
		}
		log.Printf("[planner] LLM planning failed, falling back to rule-based: %v", err)
	}

	// Fall back to rule-based planning
	return p.planWithRules(userMessage, plugins), nil
}

// planWithLLM uses the LLM to generate a plan.
func (p *Planner) planWithLLM(ctx ChatContext, userMessage string) (*ActionPlan, error) {
	systemPrompt := p.builder.BuildSystemPrompt()

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	resp, err := p.llm.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM chat failed: %w", err)
	}

	// Parse the JSON response
	content := strings.TrimSpace(resp.Content)
	// Remove markdown code fences if present
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var plan ActionPlan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse LLM plan: %w (content: %s)", err, content[:min(len(content), 200)])
	}

	return &plan, nil
}

// planWithRules provides a fallback rule-based planning system.
// This is used when no LLM is available.
func (p *Planner) planWithRules(userMessage string, plugins []string) *ActionPlan {
	msg := strings.ToLower(userMessage)

	// Detect task intent
	switch {
	case containsAny(msg, "train", "training", "model", "yolo", "cnn", "detect", "detection", "classify", "classification"):
		return p.planTraining(msg, plugins)

	case containsAny(msg, "install", "plugin", "add"):
		return p.planInstall(msg, plugins)

	case containsAny(msg, "check", "environment", "env", "status", "python", "cuda"):
		return p.planEnvironmentCheck(msg)

	case containsAny(msg, "run", "execute", "start", "workflow"):
		return p.planRunWorkflow(msg)

	case containsAny(msg, "dataset", "data", "prepare", "preprocess"):
		return p.planDataPrep(msg, plugins)

	default:
		return p.planGeneral(msg)
	}
}

// planTraining generates a plan for model training tasks.
func (p *Planner) planTraining(msg string, plugins []string) *ActionPlan {
	goal := extractGoal(msg, "train AI model")

	plugin := "YOLO"
	for _, p := range plugins {
		pl := strings.ToLower(p)
		if strings.Contains(pl, "yolo") || strings.Contains(pl, "vision") {
			plugin = p
			break
		}
	}

	return &ActionPlan{
		Goal:        goal,
		Explanation: "I'll set up a training workflow for your model. First, let me check the environment, then create the workflow.",
		Steps: []Action{
			{
				Tool:   "check_environment",
				Params: map[string]interface{}{},
				Reason: "Ensure Python, CUDA, and dependencies are ready for training",
			},
			{
				Tool: "create_workflow",
				Params: map[string]interface{}{
					"name": goal,
					"workflow": map[string]interface{}{
						"name":        goal,
						"description": "Auto-generated training workflow",
						"plugin":      plugin,
						"nodes": []map[string]interface{}{
							{
								"id":     "dataset_input",
								"type":   "input",
								"plugin": "data-source",
								"label":  "Dataset",
								"config": map[string]interface{}{},
							},
							{
								"id":     "train_model",
								"type":   "vision",
								"plugin": plugin,
								"label":  "Train Model",
								"config": map[string]interface{}{
									"task": "train",
								},
							},
							{
								"id":     "export_output",
								"type":   "output",
								"plugin": "data-source",
								"label":  "Export Model",
								"config": map[string]interface{}{},
							},
						},
						"edges": []map[string]interface{}{
							{
								"id":     "e1",
								"source": map[string]string{"node_id": "dataset_input", "port_id": "output"},
								"target": map[string]string{"node_id": "train_model", "port_id": "input"},
							},
							{
								"id":     "e2",
								"source": map[string]string{"node_id": "train_model", "port_id": "output"},
								"target": map[string]string{"node_id": "export_output", "port_id": "input"},
							},
						},
					},
				},
				Reason: "Create a training workflow with your selected plugin",
			},
		},
	}
}

// planInstall generates a plan for plugin installation.
func (p *Planner) planInstall(msg string, plugins []string) *ActionPlan {
	goal := extractGoal(msg, "install plugin")

	pluginName := ""
	for _, p := range plugins {
		pl := strings.ToLower(p)
		if strings.Contains(msg, pl) {
			pluginName = p
			break
		}
	}

	plan := &ActionPlan{
		Goal:        goal,
		Explanation: "I'll help you install the plugin.",
		Steps: []Action{
			{
				Tool:   "list_plugins",
				Params: map[string]interface{}{},
				Reason: "Check available plugins first",
			},
		},
	}

	if pluginName != "" {
		plan.Steps = append(plan.Steps, Action{
			Tool:   "install_plugin",
			Params: map[string]interface{}{"plugin_name": pluginName},
			Reason: "Install the requested plugin",
		})
	}

	return plan
}

// planEnvironmentCheck generates a plan for environment checking.
func (p *Planner) planEnvironmentCheck(msg string) *ActionPlan {
	return &ActionPlan{
		Goal:        "check environment",
		Explanation: "I'll check your AI development environment status.",
		Steps: []Action{
			{
				Tool:   "check_environment",
				Params: map[string]interface{}{},
				Reason: "Check Python, CUDA, and dependencies status",
			},
		},
	}
}

// planRunWorkflow generates a plan for running a workflow.
func (p *Planner) planRunWorkflow(msg string) *ActionPlan {
	return &ActionPlan{
		Goal:        "run workflow",
		Explanation: "I'll run the specified workflow for you.",
		Steps: []Action{
			{
				Tool:   "check_environment",
				Params: map[string]interface{}{},
				Reason: "Ensure environment is ready before running",
			},
			{
				Tool:   "run_workflow",
				Params: map[string]interface{}{"workflow_id": "from_context"},
				Reason: "Execute the workflow",
			},
		},
	}
}

// planDataPrep generates a plan for data preparation tasks.
func (p *Planner) planDataPrep(msg string, plugins []string) *ActionPlan {
	goal := extractGoal(msg, "prepare data")

	return &ActionPlan{
		Goal:        goal,
		Explanation: "I'll help you prepare and process your dataset.",
		Steps: []Action{
			{
				Tool:   "check_environment",
				Params: map[string]interface{}{},
				Reason: "Check environment for data processing tools",
			},
			{
				Tool: "create_workflow",
				Params: map[string]interface{}{
					"name": goal,
					"workflow": map[string]interface{}{
						"name":        goal,
						"description": "Auto-generated data preparation workflow",
						"nodes": []map[string]interface{}{
							{
								"id":     "data_input",
								"type":   "input",
								"plugin": "data-source",
								"label":  "Input Data",
								"config": map[string]interface{}{},
							},
							{
								"id":     "data_process",
								"type":   "logic",
								"plugin": "logic",
								"label":  "Process Data",
								"config": map[string]interface{}{
									"operation": "preprocess",
								},
							},
							{
								"id":     "data_output",
								"type":   "output",
								"plugin": "data-source",
								"label":  "Output Data",
								"config": map[string]interface{}{},
							},
						},
						"edges": []map[string]interface{}{
							{
								"id":     "e1",
								"source": map[string]string{"node_id": "data_input", "port_id": "output"},
								"target": map[string]string{"node_id": "data_process", "port_id": "input"},
							},
							{
								"id":     "e2",
								"source": map[string]string{"node_id": "data_process", "port_id": "output"},
								"target": map[string]string{"node_id": "data_output", "port_id": "input"},
							},
						},
					},
				},
				Reason: "Create a data processing workflow",
			},
		},
	}
}

// planGeneral generates a generic help plan.
func (p *Planner) planGeneral(msg string) *ActionPlan {
	return &ActionPlan{
		Goal:        "understand request",
		Explanation: "I can help you with AI tasks. Let me check what's available and suggest next steps.",
		Steps: []Action{
			{
				Tool:   "list_plugins",
				Params: map[string]interface{}{},
				Reason: "Show available plugins and capabilities",
			},
			{
				Tool:   "check_environment",
				Params: map[string]interface{}{},
				Reason: "Check current environment status",
			},
		},
	}
}

// ---- Helper functions ----

func extractGoal(msg, defaultGoal string) string {
	// Remove common prefixes
	goal := strings.TrimSpace(msg)
	for _, prefix := range []string{"help me ", "i want to ", "can you ", "please ", "帮我", "请"} {
		goal = strings.TrimPrefix(strings.ToLower(goal), prefix)
	}
	if len(goal) > 100 {
		goal = goal[:100]
	}
	if goal == "" {
		goal = defaultGoal
	}
	return goal
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}