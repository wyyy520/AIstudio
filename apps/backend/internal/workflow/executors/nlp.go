package executors

import (
	"context"
	"fmt"

	"github.com/aistudio/backend/internal/engine"
)

func NLPExecutor(client engine.EngineClient) func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		text, _ := inputs["text"].(string)
		if text == "" {
			text, _ = inputs["prompt"].(string)
		}

		task, _ := config["task"].(string)
		if task == "" {
			task, _ = config["model"].(string)
		}
		if task == "" {
			task = "text-generation"
		}

		resp, err := client.Infer(ctx, engine.InferRequest{
			ModelName: task,
			Input: map[string]interface{}{
				"text": text,
			},
			Params: config,
		})
		if err != nil {
			return nil, fmt.Errorf("nlp inference failed: %w", err)
		}
		if resp.Error != "" {
			return nil, fmt.Errorf("engine error: %s", resp.Error)
		}

		result := resp.Result
		if result == "" {
			if resp.Output != nil {
				if r, ok := resp.Output["text"]; ok {
					result, _ = r.(string)
				}
			}
		}

		return map[string]interface{}{
			"result":     result,
			"confidence": resp.Confidence,
			"status":     "completed",
		}, nil
	}
}
