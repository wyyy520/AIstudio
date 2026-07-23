// Package agent implements the AI Agent — the intelligent assistant that
// builds and modifies Workflows through natural language conversation.
//
// Architecture (EngStudio.md §7):
//
//	User message → Planner (LLM) → Plan (sequence of ToolActions)
//	→ Executor → ToolRegistry (CreateNode, ConnectNodes, FillConfig, ...)
//	→ Workflow (mutated in-place) → Memory (conversation history)
//
// The Agent is stateless; conversation state lives in Memory. Each Chat call:
//  1. Retrieves or creates a Conversation from Memory
//  2. Sends user message + tool list to Planner (LLM)
//  3. Executes the returned Plan via ToolRegistry
//  4. Returns a summary message to the user
//
// Key design decisions:
//   - LLM-agnostic: Planner accepts any LLMProvider implementation
//   - Tool-based: All workflow mutations go through Tools (safe, auditable)
//   - Event-driven: Publishes to EventBus for UI updates
//   - Memory-bound: Conversations auto-expire; max 50 messages retained
//
// Usage:
//
//	agent := agent.NewAgent(skillManager).
//	    WithLLM(myLLM).
//	    WithEventBus(bus)
//	reply, _ := agent.Chat(ctx, "session-1", "Build an image classifier")
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

// ============================================================================
// Agent — the top-level AI assistant
// ============================================================================

// Agent combines planning, execution, memory, and tools into a single
// conversational interface for workflow creation and modification.
type Agent struct {
	planner   *Planner       // LLM-driven action planner
	executor  *Executor      // Tool action executor
	memory    *Memory        // Conversation history (50 msg cap)
	tools     *ToolRegistry  // Available tools (CreateNode, ConnectNodes, ...)
	skillMgr  *skill.SkillManager // Skill/template manager
	eventBus  *event.EventBus     // Optional event bus for UI updates
}

// ============================================================================
// Construction — fluent builder pattern
// ============================================================================

// NewAgent creates a new Agent with the given SkillManager.
// Tools are auto-registered: CreateNode, ConnectNodes, FillConfig,
// ValidateWorkflow, ApplySkill, SearchSkills.
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

// ============================================================================
// Tool Registration — all tools available to the Planner
// ============================================================================

// registerTools registers the 6 built-in tools that the Agent can use.
// Tools are the ONLY way the Agent can modify a Workflow.
func (a *Agent) registerTools() {
	a.tools.Register(CreateNodeTool())
	a.tools.Register(ConnectNodesTool())
	a.tools.Register(FillConfigTool())
	a.tools.Register(ValidateWorkflowTool())
	a.tools.Register(ApplySkillTool(a.skillMgr))
	a.tools.Register(SearchSkillsTool(a.skillMgr))
}

// ============================================================================
// Chat — the main conversational entry point
// ============================================================================

// Chat processes a user message and returns an assistant reply.
//
// Steps:
//  1. Load or create Conversation from Memory
//  2. Append user message to conversation history
//  3. Determine target (defaults to Python if no workflow exists)
//  4. Ask Planner to generate a Plan from the message + available tools
//  5. Execute the Plan via ToolRegistry (mutates the Workflow)
//  6. Return a summary message (success/failure count)
//  7. Publish agent:chat:completed event
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

// ============================================================================
// GenerateWorkflow — create a complete workflow from a description
// ============================================================================

// GenerateWorkflow creates a brand-new Workflow from a plain-text description.
//
// Steps:
//  1. Plan: ask the LLM to decompose the description into tool actions
//  2. Create a blank Workflow with the given target
//  3. Execute the Plan against the blank Workflow
//  4. Validate: fail if the workflow has no nodes
//  5. Publish agent:workflow:generated event
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

// ============================================================================
// ImproveWorkflow — modify an existing workflow based on feedback
// ============================================================================

// ImproveWorkflow applies user feedback to modify an existing Workflow.
//
// Steps:
//  1. Plan improvements from the feedback string
//  2. Execute improvement actions against the existing Workflow
//  3. Update the UpdatedAt timestamp
//  4. Publish agent:workflow:improved event
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