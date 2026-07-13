package executors

import (
	"context"
)

func LoopExecutor() func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		maxIter := 10
		if v, ok := config["maxIterations"].(float64); ok {
			maxIter = int(v)
		}

		currentIter := 0
		if v, ok := inputs["currentIteration"].(float64); ok {
			currentIter = int(v)
		}
		currentIter++

		shouldContinue := currentIter <= maxIter

		return map[string]interface{}{
			"continue":         shouldContinue,
			"currentIteration": currentIter,
		}, nil
	}
}
