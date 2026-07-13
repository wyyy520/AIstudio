package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aistudio/packages/workflow"
)

type LLMProvider interface {
	Chat(ctx context.Context, systemPrompt string, messages []Message) (*Message, error)
}

type Planner struct {
	llm      LLMProvider
	messages []Message
}

func NewPlanner() *Planner {
	return &Planner{
		messages: make([]Message, 0),
	}
}

func (p *Planner) WithLLM(llm LLMProvider) *Planner {
	p.llm = llm
	return p
}

func (p *Planner) Plan(ctx context.Context, description string, target workflow.Target, tools []ToolDef) (*Plan, error) {
	if p.llm != nil {
		plan, err := p.planWithLLM(ctx, description, target, tools)
		if err == nil && plan != nil && len(plan.Steps) > 0 {
			return plan, nil
		}
	}

	return p.planWithRules(description, target), nil
}

func (p *Planner) planWithLLM(ctx context.Context, description string, target workflow.Target, tools []ToolDef) (*Plan, error) {
	toolDescs := make([]string, len(tools))
	for i, t := range tools {
		toolDescs[i] = fmt.Sprintf("- %s: %s", t.Name, t.Description)
	}

	systemPrompt := fmt.Sprintf(`You are an AI workflow planner. Given a user's requirement, determine the steps needed to create the workflow.

Available actions:
%s

Respond with a JSON object:
{
  "goal": "brief goal description",
  "explanation": "human readable explanation",
  "steps": [
    {"action": "action_name", "params": {...}, "reason": "why this step"}
  ]
}`, strings.Join(toolDescs, "\n"))

	messages := []Message{
		{Role: RoleSystem, Content: systemPrompt, Timestamp: time.Now()},
		{Role: RoleUser, Content: fmt.Sprintf("Create a workflow for: %s (target: %s)", description, target), Timestamp: time.Now()},
	}

	resp, err := p.llm.Chat(ctx, systemPrompt, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM planning failed: %w", err)
	}

	content := strings.TrimSpace(resp.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var plan Plan
	if err := json.Unmarshal([]byte(content), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse LLM plan: %w", err)
	}

	return &plan, nil
}

func (p *Planner) planWithRules(description string, target workflow.Target) *Plan {
	desc := strings.ToLower(description)

	goal := desc
	if len(goal) > 100 {
		goal = goal[:100]
	}

	return &Plan{
		Goal:        goal,
		Explanation: fmt.Sprintf("I'll help you create a %s workflow for %s.", target, description),
		Steps: []PlanStep{
			{
				Action: "search_skills",
				Params: map[string]any{"query": description},
				Reason: "Find relevant skill templates for your task",
			},
			{
				Action: "create_node",
				Params: map[string]any{"type": "data_loader", "name": "Data Loader",
					"config": map[string]any{"dataset": "${dataset_path}"}},
				Reason: "Add a data loader node to load your dataset",
			},
			{
				Action: "create_node",
				Params: map[string]any{"type": "data_preprocessor", "name": "Preprocessor",
					"config": map[string]any{}},
				Reason: "Add preprocessing for data preparation",
			},
			{
				Action: "create_node",
				Params: map[string]any{"type": "model_trainer", "name": "Trainer",
					"config": map[string]any{"epochs": 100, "batch_size": 32}},
				Reason: "Add a model training node",
			},
			{
				Action: "create_node",
				Params: map[string]any{"type": "model_evaluator", "name": "Evaluator",
					"config": map[string]any{}},
				Reason: "Add evaluation to measure model performance",
			},
			{
				Action: "connect_nodes",
				Params: map[string]any{"source_id": "${node:data_loader}", "target_id": "${node:preprocessor}"},
				Reason: "Connect data loader to preprocessor",
			},
			{
				Action: "connect_nodes",
				Params: map[string]any{"source_id": "${node:preprocessor}", "target_id": "${node:trainer}"},
				Reason: "Connect preprocessor to trainer",
			},
			{
				Action: "connect_nodes",
				Params: map[string]any{"source_id": "${node:trainer}", "target_id": "${node:evaluator}"},
				Reason: "Connect trainer to evaluator",
			},
			{
				Action: "validate_workflow",
				Params: map[string]any{},
				Reason: "Validate the workflow structure",
			},
		},
	}
}