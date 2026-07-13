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

	// Detect workflow intent
	switch {
	case containsAny(msg, "install", "plugin", "add"):
		return p.planInstall(msg, plugins)

	case containsAny(msg, "check", "environment", "env", "status", "python", "cuda"):
		return p.planEnvironmentCheck(msg)

	case containsAny(msg, "list", "available", "skill", "template"):
		return p.planListSkills(msg)

	default:
		return p.planCreateWorkflow(msg, plugins)
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

// planListSkills generates a plan for listing skills.
func (p *Planner) planListSkills(msg string) *ActionPlan {
	return &ActionPlan{
		Goal:        "list available skills",
		Explanation: "I'll show you the available workflow templates and skills.",
		Steps: []Action{
			{
				Tool:   "list_skills",
				Params: map[string]interface{}{},
				Reason: "Show available workflow skills and templates",
			},
		},
	}
}

// planCreateWorkflow generates a plan for creating a workflow.
func (p *Planner) planCreateWorkflow(msg string, plugins []string) *ActionPlan {
	goal := extractGoal(msg, "create workflow")

	return &ActionPlan{
		Goal:        goal,
		Explanation: "I'll help you create a workflow. First, let me check your environment and available plugins, then create the workflow.",
		Steps: []Action{
			{
				Tool:   "check_environment",
				Params: map[string]interface{}{},
				Reason: "Check environment readiness",
			},
			{
				Tool:   "list_plugins",
				Params: map[string]interface{}{},
				Reason: "Check available plugins for the workflow",
			},
			{
				Tool:   "list_skills",
				Params: map[string]interface{}{},
				Reason: "Check available workflow templates",
			},
			{
				Tool: "create_workflow",
				Params: map[string]interface{}{
					"name":     goal,
					"workflow": map[string]interface{}{},
				},
				Reason: "Create the workflow with selected plugins and nodes",
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