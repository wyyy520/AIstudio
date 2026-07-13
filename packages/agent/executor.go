package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aistudio/packages/event"
)

type Executor struct {
	tools      *ToolRegistry
	eventBus   *event.EventBus
	timeout    time.Duration
	maxRetries int
}

func NewExecutor(tools *ToolRegistry) *Executor {
	return &Executor{
		tools:      tools,
		timeout:    30 * time.Second,
		maxRetries: 3,
	}
}

func (e *Executor) WithEventBus(bus *event.EventBus) *Executor {
	e.eventBus = bus
	return e
}

func (e *Executor) WithTimeout(timeout time.Duration) *Executor {
	e.timeout = timeout
	return e
}

func (e *Executor) WithMaxRetries(retries int) *Executor {
	e.maxRetries = retries
	return e
}

func (e *Executor) Execute(ctx context.Context, plan *Plan, toolCtx *ToolContext) []StepResult {
	results := make([]StepResult, 0)

	for i, step := range plan.Steps {
		stepNum := i + 1
		log.Printf("[executor] step %d/%d: %s", stepNum, len(plan.Steps), step.Action)

		select {
		case <-ctx.Done():
			results = append(results, StepResult{
				Step:      stepNum,
				Action:    step.Action,
				Success:   false,
				Error:     "execution cancelled",
				Timestamp: time.Now(),
			})
			return results
		default:
		}

		tool, ok := e.tools.Get(step.Action)
		if !ok {
			results = append(results, StepResult{
				Step:    stepNum,
				Action:  step.Action,
				Success: false,
				Error:   fmt.Sprintf("unknown tool: %s", step.Action),
				Timestamp: time.Now(),
			})
			continue
		}

		result := e.executeWithRetry(ctx, stepNum, step.Action, step.Params, tool, toolCtx)
		results = append(results, result)

		if e.eventBus != nil {
			e.emitStepEvent(result)
		}

		if !result.Success {
			log.Printf("[executor] step %d failed: %s", stepNum, result.Error)
		}
	}

	return results
}

func (e *Executor) executeWithRetry(ctx context.Context, stepNum int, action string, params map[string]any, tool ToolDef, toolCtx *ToolContext) StepResult {
	var lastErr error

	for attempt := 0; attempt <= e.maxRetries; attempt++ {
		start := time.Now()

		execCtx, cancel := context.WithTimeout(ctx, e.timeout)
		defer cancel()

		if attempt > 0 {
			log.Printf("[executor] retry %d/%d for step %d: %s", attempt, e.maxRetries, stepNum, action)
		}

		data, err := tool.Execute(execCtx, toolCtx, params)
		duration := time.Since(start)

		if err == nil {
			return StepResult{
				Step:      stepNum,
				Action:    action,
				Success:   true,
				Data:      data,
				Duration:  duration.String(),
				Timestamp: start,
			}
		}

		lastErr = err
		cancel()
	}

	return StepResult{
		Step:      stepNum,
		Action:    action,
		Success:   false,
		Error:     fmt.Sprintf("failed after %d retries: %v", e.maxRetries, lastErr),
		Timestamp: time.Now(),
	}
}

func (e *Executor) Rollback(ctx context.Context, results []StepResult) error {
	log.Printf("[executor] rolling back %d steps", len(results))

	for i := len(results) - 1; i >= 0; i-- {
		result := results[i]
		if result.Success {
			log.Printf("[executor] rollback step %d: %s", result.Step, result.Action)
		}
	}

	return nil
}

func (e *Executor) emitStepEvent(result StepResult) {
	topic := event.Topic("agent:step:completed")
	if !result.Success {
		topic = "agent:step:failed"
	}
	e.eventBus.Publish(topic, result)
}