package workflow

import (
	"context"
	"fmt"
	"sync"
)

type Engine struct {
	registry   *NodeRegistry
	executor   *Executor
	taskStore  sync.Map
}

func NewEngine(registry *NodeRegistry) *Engine {
	return &Engine{
		registry:  registry,
		executor:  NewExecutor(registry),
		taskStore: sync.Map{},
	}
}

func NewDefaultEngine() *Engine {
	return NewEngine(DefaultRegistry)
}

func (e *Engine) Run(ctx context.Context, workflowJSON []byte) (*ExecutionResult, error) {
	wf, err := ParseAndValidate(workflowJSON)
	if err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	result, err := e.executor.Execute(ctx, wf)
	if err != nil {
		return nil, fmt.Errorf("workflow execution failed: %w", err)
	}

	e.taskStore.Store(result.TaskID, result)

	return result, nil
}

func (e *Engine) RunWithWorkflow(ctx context.Context, wf *Workflow) (*ExecutionResult, error) {
	result, err := e.executor.Execute(ctx, wf)
	if err != nil {
		return nil, fmt.Errorf("workflow execution failed: %w", err)
	}
	e.taskStore.Store(result.TaskID, result)
	return result, nil
}

func (e *Engine) GetTask(taskID string) (*ExecutionResult, bool) {
	if val, ok := e.taskStore.Load(taskID); ok {
		return val.(*ExecutionResult), true
	}
	return nil, false
}

func (e *Engine) Registry() *NodeRegistry {
	return e.registry
}
