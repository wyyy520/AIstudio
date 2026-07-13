package executors

import (
	"context"
	"fmt"

	"github.com/aistudio/backend/internal/engine"
)

func YOLOTrainExecutor(client engine.EngineClient) func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		dataset, _ := inputs["dataset"].(string)
		modelName, _ := config["model_name"].(string)
		if modelName == "" {
			modelName = "yolov8"
		}

		resp, err := client.Train(ctx, engine.TrainRequest{
			Dataset:   dataset,
			Config:    config,
			ModelName: modelName,
		})
		if err != nil {
			return nil, fmt.Errorf("yolo train failed: %w", err)
		}
		if resp.Error != "" {
			return nil, fmt.Errorf("engine error: %s", resp.Error)
		}

		return map[string]interface{}{
			"model_path":  resp.ModelPath,
			"metrics":     resp.Metrics,
			"duration_ms": resp.DurationMs,
			"status":      "completed",
		}, nil
	}
}
