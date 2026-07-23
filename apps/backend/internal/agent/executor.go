package agent

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Executor executes the action plan step by step.
// It calls tools via the ToolRegistry and records results.
type Executor struct {
	tools   *ToolRegistry
	context *ContextManager
}

// NewExecutor creates a new Executor.
func NewExecutor(tools *ToolRegistry, ctx *ContextManager) *Executor {
	return &Executor{
		tools:   tools,
		context: ctx,
	}
}

// Execute runs all steps in the action plan sequentially.
// It stops on the first failure unless continueOnError is true.
func (e *Executor) Execute(ctx context.Context, plan *ActionPlan, continueOnError bool) []StepResult {
	return e.ExecuteWithCallback(ctx, plan, continueOnError, nil)
}

// StepCallback is called after each action step completes.
type StepCallback func(stepNum int, action Action, result StepResult)

// ExecuteWithCallback runs all steps with a callback fired after each step completes.
// The callback is called synchronously in the execution goroutine.
func (e *Executor) ExecuteWithCallback(ctx context.Context, plan *ActionPlan, continueOnError bool, cb StepCallback) []StepResult {
	log.Printf("[executor] executing plan: %s (%d steps)", plan.Goal, len(plan.Steps))

	var results []StepResult

	for i, action := range plan.Steps {
		stepNum := i + 1
		log.Printf("[executor] step %d/%d: %s", stepNum, len(plan.Steps), action.Tool)

		// Check context cancellation
		select {
		case <-ctx.Done():
			result := StepResult{
				Step:      stepNum,
				Tool:      action.Tool,
				Success:   false,
				Error:     "execution cancelled",
				Timestamp: time.Now(),
			}
			results = append(results, result)
			if cb != nil {
				cb(stepNum, action, result)
			}
			return results
		default:
		}

		result := e.executeStep(ctx, stepNum, action)
		results = append(results, result)
		e.context.AddStepResult(result)

		if cb != nil {
			cb(stepNum, action, result)
		}

		// Stop on failure unless continueOnError
		if !result.Success && !continueOnError {
			log.Printf("[executor] step %d failed, stopping execution", stepNum)
			break
		}
	}

	log.Printf("[executor] plan execution complete: %d/%d steps succeeded",
		countSuccess(results), len(results))
	return results
}

// executeStep runs a single action and returns the result.
func (e *Executor) executeStep(ctx context.Context, stepNum int, action Action) StepResult {
	start := time.Now()

	tool, ok := e.tools.Get(action.Tool)
	if !ok {
		return StepResult{
			Step:      stepNum,
			Tool:      action.Tool,
			Success:   false,
			Error:     fmt.Sprintf("unknown tool: %s", action.Tool),
			Timestamp: start,
		}
	}

	toolResult, err := tool.Execute(ctx, action.Params)
	duration := time.Since(start)

	if err != nil {
		return StepResult{
			Step:      stepNum,
			Tool:      action.Tool,
			Success:   false,
			Error:     err.Error(),
			Timestamp: start,
		}
	}

	log.Printf("[executor] step %d: %s completed in %v (success=%v)", stepNum, action.Tool, duration, toolResult.Success)

	return StepResult{
		Step:      stepNum,
		Tool:      action.Tool,
		Success:   toolResult.Success,
		Data:      toolResult.Data,
		Error:     toolResult.Error,
		Timestamp: start,
	}
}

// countSuccess counts the number of successful steps.
func countSuccess(results []StepResult) int {
	count := 0
	for _, r := range results {
		if r.Success {
			count++
		}
	}
	return count
}