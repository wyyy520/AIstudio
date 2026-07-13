package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aistudio/packages/event"
	"github.com/aistudio/packages/skill"
	"github.com/aistudio/packages/workflow"
	"github.com/google/uuid"
)

type Agent struct {
	planner   *Planner
	executor  *Executor
	memory    *Memory
	tools     *ToolRegistry
	skillMgr  *skill.SkillManager
	eventBus  *event.EventBus
}

func NewAgent(skillMgr *skill.SkillManager) *Agent {
	tools := NewToolRegistry()
	agent := &Agent{
		planner:  NewPlanner(),
		executor: NewExecutor(tools),
		memory:   NewMemory(50),
		tools:    tools,
		skillMgr: skillMgr,
	}

	agent.registerTools()
	return agent
}

func (a *Agent) WithLLM(llm LLMProvider) *Agent {
	a.planner.WithLLM(llm)
	return a
}

func (a *Agent) WithEventBus(bus *event.EventBus) *Agent {
	a.eventBus = bus
	a.executor.WithEventBus(bus)
	return a
}

func (a *Agent) WithMemory(mem *Memory) *Agent {
	a.memory = mem
	return a
}

func (a *Agent) ToolRegistry() *ToolRegistry {
	return a.tools
}

func (a *Agent) registerTools() {
	a.tools.Register(CreateNodeTool())
	a.tools.Register(ConnectNodesTool())
	a.tools.Register(FillConfigTool())
	a.tools.Register(ValidateWorkflowTool())
	a.tools.Register(ApplySkillTool(a.skillMgr))
	a.tools.Register(SearchSkillsTool(a.skillMgr))
}

func (a *Agent) Chat(ctx context.Context, conversationID string, message string) (*Message, error) {
	conv := a.memory.GetConversation(conversationID)
	if conv == nil {
		conv = a.memory.CreateConversation(conversationID)
	}

	userMsg := Message{
		Role:      RoleUser,
		Content:   message,
		Timestamp: time.Now(),
	}
	a.memory.AddMessage(conversationID, userMsg)

	target := workflow.TargetPython
	if conv.Workflow != nil {
		target = conv.Workflow.Target
	}

	plan, err := a.planner.Plan(ctx, message, target, a.tools.List())
	if err != nil {
		respMsg := Message{
			Role:    RoleAssistant,
			Content: fmt.Sprintf("I couldn't create a plan for your request: %v", err),
			Timestamp: time.Now(),
		}
		a.memory.AddMessage(conversationID, respMsg)
		return &respMsg, nil
	}

	toolCtx := &ToolContext{
		Workflow: conv.Workflow,
	}
	if toolCtx.Workflow == nil {
		toolCtx.Workflow = &workflow.Workflow{
			ID:       uuid.New().String(),
			Target:   target,
			Name:     plan.Goal,
			SchemaVersion: workflow.CurrentSchemaVersion,
			Nodes:    make([]workflow.Node, 0),
			Edges:    make([]workflow.Edge, 0),
			CreatedAt: time.Now(),
		}
	}

	if a.eventBus != nil {
		a.eventBus.Publish(event.Topic("agent:plan:created"), plan)
	}

	results := a.executor.Execute(ctx, plan, toolCtx)
	a.memory.SetWorkflow(conversationID, toolCtx.Workflow)

	successCount := 0
	var failures []string
	for _, r := range results {
		if r.Success {
			successCount++
		} else {
			failures = append(failures, fmt.Sprintf("%s: %s", r.Action, r.Error))
		}
	}

	summary := fmt.Sprintf("Completed %d/%d steps.", successCount, len(results))
	if len(failures) > 0 {
		summary += fmt.Sprintf(" Failures: %v", failures)
	}

	respMsg := Message{
		Role:    RoleAssistant,
		Content: summary,
		Timestamp: time.Now(),
	}
	a.memory.AddMessage(conversationID, respMsg)

	if a.eventBus != nil {
		a.eventBus.Publish(event.Topic("agent:chat:completed"), map[string]any{
			"conversation_id": conversationID,
			"results":         results,
			"workflow":        toolCtx.Workflow,
		})
	}

	return &respMsg, nil
}

func (a *Agent) GenerateWorkflow(ctx context.Context, description string, target workflow.Target) (*workflow.Workflow, error) {
	log.Printf("[agent] generating workflow: %s (target: %s)", description, target)

	plan, err := a.planner.Plan(ctx, description, target, a.tools.List())
	if err != nil {
		return nil, fmt.Errorf("planning failed: %w", err)
	}

	wf := &workflow.Workflow{
		ID:           uuid.New().String(),
		Name:         plan.Goal,
		Description:  description,
		Target:       target,
		Version:      1,
		SchemaVersion: workflow.CurrentSchemaVersion,
		Nodes:        make([]workflow.Node, 0),
		Edges:        make([]workflow.Edge, 0),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	toolCtx := &ToolContext{Workflow: wf}
	results := a.executor.Execute(ctx, plan, toolCtx)

	for _, r := range results {
		if !r.Success {
			return nil, fmt.Errorf("execution failed at step %d (%s): %s", r.Step, r.Action, r.Error)
		}
	}

	if len(wf.Nodes) == 0 {
		return nil, fmt.Errorf("workflow has no nodes")
	}

	if a.eventBus != nil {
		a.eventBus.Publish(event.Topic("agent:workflow:generated"), map[string]any{
			"workflow_id": wf.ID,
			"description": description,
			"target":      target,
		})
	}

	return wf, nil
}

func (a *Agent) ImproveWorkflow(ctx context.Context, wf *workflow.Workflow, feedback string) (*workflow.Workflow, error) {
	log.Printf("[agent] improving workflow %s: %s", wf.ID, feedback)

	plan, err := a.planner.Plan(ctx, feedback, wf.Target, a.tools.List())
	if err != nil {
		return nil, fmt.Errorf("planning improvement failed: %w", err)
	}

	toolCtx := &ToolContext{Workflow: wf}
	results := a.executor.Execute(ctx, plan, toolCtx)

	for _, r := range results {
		if !r.Success {
			return nil, fmt.Errorf("improvement failed at step %d (%s): %s", r.Step, r.Action, r.Error)
		}
	}

	wf.UpdatedAt = time.Now()

	if a.eventBus != nil {
		a.eventBus.Publish(event.Topic("agent:workflow:improved"), map[string]any{
			"workflow_id": wf.ID,
			"feedback":    feedback,
		})
	}

	return wf, nil
}