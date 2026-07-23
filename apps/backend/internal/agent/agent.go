package agent

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Agent is the core AI Agent that processes natural language requests,
// plans actions, and executes them using available tools.
type Agent struct {
	planner  *Planner
	executor *Executor
	memory   *Memory
	context  *ContextManager
	tools    *ToolRegistry
	llm      LLMProvider
}

// NewAgent creates a new Agent instance.
// llm may be nil for rule-based-only operation.
func NewAgent(llm LLMProvider, memory *Memory) *Agent {
	ctx := NewContextManager()
	tools := NewToolRegistry()

	return &Agent{
		planner:  NewPlanner(llm, memory),
		executor: NewExecutor(tools, ctx),
		memory:   memory,
		context:  ctx,
		tools:    tools,
		llm:      llm,
	}
}

// ToolRegistry returns the agent's tool registry for external wiring.
func (a *Agent) ToolRegistry() *ToolRegistry {
	return a.tools
}

// Context returns the agent's context manager.
func (a *Agent) Context() *ContextManager {
	return a.context
}

// Memory returns the agent's memory store.
func (a *Agent) Memory() *Memory {
	return a.memory
}

// StreamEvent represents a real-time event emitted during agent processing.
type StreamEvent struct {
	Type    string      `json:"type"`    // "token", "action", "done", "error"
	Content string      `json:"content"` // human-readable text content
	Meta    interface{} `json:"meta"`    // optional structured metadata
}

// StreamCallback is called with each StreamEvent during processing.
// The callback should be non-blocking; if nil, events are discarded.
type StreamCallback func(evt StreamEvent)

// Process handles a user message end-to-end (non-streaming).
// Use StreamProcess for real-time SSE feedback.
func (a *Agent) Process(ctx context.Context, message, projectID, userID string, plugins []string, envStatus string) (*AgentResponse, error) {
	return a.StreamProcess(ctx, message, projectID, userID, plugins, envStatus, nil)
}

// StreamProcess handles a user message end-to-end with real-time streaming callbacks:
//  1. Plan: Analyze the message and generate an action plan
//  2. Execute: Run the plan steps sequentially (each step fires "step_start" / "step_done")
//  3. Respond: Build summary and stream via "token" events, then "done"
func (a *Agent) StreamProcess(ctx context.Context, message, projectID, userID string, plugins []string, envStatus string, cb StreamCallback) (*AgentResponse, error) {
	start := time.Now()
	log.Printf("[agent] processing message: project=%s user=%s message=%q", projectID, userID, message)

	emit := func(evt StreamEvent) {
		if cb != nil {
			cb(evt)
		}
	}

	// Set up context
	a.context.SetProject(projectID)
	a.context.SetUser(userID)
	a.context.SetGoal("")

	// Add user message to history
	a.context.AddMessage(Message{Role: "user", Content: message})

	// Save conversation entry
	if a.memory != nil {
		_ = a.memory.SaveConversation(ConversationEntry{
			ProjectID: projectID,
			UserID:    userID,
			Role:      "user",
			Content:   message,
			CreatedAt: time.Now(),
		})
	}

	// ---- Phase 1: Plan ----
	emit(StreamEvent{Type: "action", Content: "Analyzing your request...", Meta: map[string]string{"phase": "planning"}})

	toolInfos := a.tools.List()
	chatCtx := ChatContext{ProjectID: projectID, UserID: userID}

	plan, err := a.planner.Plan(chatCtx, message, toolInfos, plugins, envStatus)
	if err != nil {
		log.Printf("[agent] planning failed: %v", err)
		emit(StreamEvent{Type: "action", Content: "Planning failed, using fallback...", Meta: map[string]string{"phase": "planning_error"}})
		return &AgentResponse{
			Goal:        "error",
			Explanation: fmt.Sprintf("Failed to understand your request: %v", err),
			Status:      "failed",
		}, nil
	}

	a.context.SetGoal(plan.Goal)
	a.context.SetActionPlan(plan)

	emit(StreamEvent{Type: "action", Content: plan.Explanation, Meta: map[string]interface{}{
		"phase": "plan_ready",
		"goal":  plan.Goal,
		"steps": len(plan.Steps),
	}})

	// Save agent's plan to conversation
	if a.memory != nil {
		_ = a.memory.SaveConversation(ConversationEntry{
			ProjectID: projectID,
			UserID:    userID,
			Role:      "agent",
			Content:   plan.Explanation,
			Goal:      plan.Goal,
			CreatedAt: time.Now(),
		})
	}

	// ---- Phase 2: Execute ----
	steps := a.executor.ExecuteWithCallback(ctx, plan, false, func(stepNum int, action Action, result StepResult) {
		if result.Success {
			emit(StreamEvent{Type: "action", Content: fmt.Sprintf("Step %d/%d: %s - OK", stepNum, len(plan.Steps), action.Tool), Meta: map[string]interface{}{
				"phase":   "step_done",
				"step":    stepNum,
				"tool":    action.Tool,
				"success": true,
			}})
		} else {
			emit(StreamEvent{Type: "action", Content: fmt.Sprintf("Step %d/%d: %s - Failed: %s", stepNum, len(plan.Steps), action.Tool, result.Error), Meta: map[string]interface{}{
				"phase":   "step_done",
				"step":    stepNum,
				"tool":    action.Tool,
				"success": false,
			}})
		}
	})

	// ---- Phase 3: Build Response ----
	status := "completed"
	allSuccess := true
	for _, s := range steps {
		if !s.Success {
			allSuccess = false
			break
		}
	}
	if !allSuccess {
		status = "failed"
	}

	summary := buildSummary(plan, steps, allSuccess)

	// Stream summary tokens naturally for real-time display
	for _, r := range summary {
		emit(StreamEvent{Type: "token", Content: string(r)})
	}

	// Add agent response to history
	a.context.AddMessage(Message{Role: "assistant", Content: summary})

	duration := time.Since(start)
	log.Printf("[agent] request completed in %v: status=%s steps=%d", duration, status, len(steps))

	emit(StreamEvent{Type: "done", Content: "", Meta: map[string]interface{}{
		"status":  status,
		"steps":   len(steps),
		"goal":    plan.Goal,
		"elapsed": duration.String(),
	}})

	return &AgentResponse{
		Goal:        plan.Goal,
		Explanation: plan.Explanation,
		Plan:        plan.Steps,
		Steps:       steps,
		Status:      status,
		Summary:     summary,
	}, nil
}

// PlanOnly analyzes the message and returns a plan without executing it.
func (a *Agent) PlanOnly(ctx context.Context, message, projectID, userID string, plugins []string, envStatus string) (*ActionPlan, error) {
	a.context.SetProject(projectID)
	a.context.SetUser(userID)

	toolInfos := a.tools.List()
	chatCtx := ChatContext{ProjectID: projectID, UserID: userID}

	plan, err := a.planner.Plan(chatCtx, message, toolInfos, plugins, envStatus)
	if err != nil {
		return nil, err
	}

	a.context.SetActionPlan(plan)
	return plan, nil
}

// Reset clears the agent's current context.
func (a *Agent) Reset() {
	a.context.Reset()
}

// buildSummary creates a human-readable summary of the execution.
func buildSummary(plan *ActionPlan, steps []StepResult, allSuccess bool) string {
	if allSuccess {
		return fmt.Sprintf("Successfully completed: %s. Executed %d steps.", plan.Goal, len(steps))
	}

	successCount := 0
	var failures []string
	for _, s := range steps {
		if s.Success {
			successCount++
		} else {
			failures = append(failures, fmt.Sprintf("%s: %s", s.Tool, s.Error))
		}
	}

	summary := fmt.Sprintf("Partially completed: %s. %d/%d steps succeeded.", plan.Goal, successCount, len(steps))
	if len(failures) > 0 {
		summary += fmt.Sprintf(" Failures: %v", failures)
	}
	return summary
}