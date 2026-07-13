package sdk

import (
	"context"

	"github.com/aistudio/packages/agent"
	"github.com/aistudio/packages/workflow"
)

func GenerateWorkflow(description string) (*Workflow, error) {
	a := agent.NewAgent(nil)
	return a.GenerateWorkflow(context.Background(), description, workflow.TargetPython)
}

func Chat(message string, contextID string) (string, error) {
	a := agent.NewAgent(nil)
	msg, err := a.Chat(context.Background(), contextID, message)
	if err != nil {
		return "", err
	}
	return msg.Content, nil
}