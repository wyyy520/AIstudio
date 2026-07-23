package executors

import (
	"context"
	"fmt"

	"github.com/aistudio/backend/internal/engine"
)

func YOLOPredictExecutor(client engine.EngineClient) func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		image, _ := inputs["image"].(string)
		images, _ := inputs["images"].([]interface{})

		modelName, _ := config["model_name"].(string)
		if modelName == "" {
			modelName = "yolov8"
		}

		taskID, _ := config["task_id"].(string)
		if taskID == "" {
			taskID = "yolo-predict-" + modelName
		}

		inferInput := make(map[string]interface{})
		if image != "" {
			inferInput["image"] = image
		}
		if len(images) > 0 {
			inferInput["images"] = images
		}

		resp, err := client.Infer(ctx, engine.InferRequest{
			TaskID:    taskID,
			Plugin:    "yolo",
			ModelName: modelName,
			Input:     inferInput,
			Params:    config,
		})
		if err != nil {
			return nil, fmt.Errorf("yolo predict failed: %w", err)
		}
		if resp.Error != "" {
			return nil, fmt.Errorf("engine error: %s", resp.Error)
		}

		return map[string]interface{}{
			"detections": resp.Detections,
			"status":     "completed",
		}, nil
	}
}
