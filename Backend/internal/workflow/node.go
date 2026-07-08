package workflow

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type ExecutableNode interface {
	Execute(ctx context.Context, inputs map[string]interface{}, params map[string]interface{}) (map[string]interface{}, error)
}

type DataSourceNode struct{}

func (n *DataSourceNode) Execute(ctx context.Context, inputs map[string]interface{}, params map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"out_image":  "/storage/datasets/sample.jpg",
		"out_text":   "Hello, AI Studio!",
		"out_tensor": []float64{1.0, 2.0, 3.0, 4.0, 5.0},
	}, nil
}

type YOLODetectorNode struct{}

func (n *YOLODetectorNode) Execute(ctx context.Context, inputs map[string]interface{}, params map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"out_detections": map[string]interface{}{
			"boxes":   [][]float64{{100, 200, 300, 400}, {50, 60, 120, 180}, {300, 400, 500, 600}},
			"scores":  []float64{0.95, 0.87, 0.76},
			"classes": []int{2, 7, 2},
			"count":   3,
		},
	}, nil
}

type PyTorchNode struct{}

func (n *PyTorchNode) Execute(ctx context.Context, inputs map[string]interface{}, params map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"out_model": map[string]interface{}{
			"path":       "model_epoch_50.pt",
			"accuracy":   0.9234,
			"loss":       0.1245,
			"epochs":     50,
			"batch_size": 32,
		},
	}, nil
}

type TransformerNode struct{}

func (n *TransformerNode) Execute(ctx context.Context, inputs map[string]interface{}, params map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"out_result": map[string]interface{}{
			"label": "positive",
			"score": 0.9876,
			"embeddings": []float64{
				0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8,
			},
			"tokens": 128,
		},
	}, nil
}

type LSTMNode struct{}

func (n *LSTMNode) Execute(ctx context.Context, inputs map[string]interface{}, params map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"out_prediction": map[string]interface{}{
			"values":      []float64{10.5, 11.2, 12.1, 13.0, 14.2, 15.1, 16.0, 17.3, 18.5, 19.2},
			"upper_bound": []float64{11.0, 11.8, 12.7, 13.6, 14.8, 15.7, 16.6, 17.9, 19.1, 19.8},
			"lower_bound": []float64{10.0, 10.6, 11.5, 12.4, 13.6, 14.5, 15.4, 16.7, 17.9, 18.6},
			"mse":         0.0234,
			"mae":         0.1123,
		},
	}, nil
}

type ExportNode struct{}

func (n *ExportNode) Execute(ctx context.Context, inputs map[string]interface{}, params map[string]interface{}) (map[string]interface{}, error) {
	format := "onnx"
	if v, ok := params["format"]; ok {
		format = fmt.Sprintf("%v", v)
	}
	return map[string]interface{}{
		"out_path": fmt.Sprintf("/storage/exports/model_%s.%s", time.Now().Format("20060102_150405"), format),
	}, nil
}

type MCPNode struct {
	MCPRuntime MCPRuntime
}

func (n *MCPNode) Execute(ctx context.Context, inputs map[string]interface{}, params map[string]interface{}) (map[string]interface{}, error) {
	server := "default"
	if v, ok := params["server"]; ok {
		server = fmt.Sprintf("%v", v)
	}

	toolName := "default"
	if v, ok := params["tool"]; ok {
		toolName = fmt.Sprintf("%v", v)
	}

	// Use real MCP runtime if available
	if n.MCPRuntime != nil {
		result, err := n.MCPRuntime.CallTool(ctx, server, toolName, params)
		if err != nil {
			return nil, fmt.Errorf("mcp call failed: %w", err)
		}
		if !result.Success {
			return nil, fmt.Errorf("mcp tool error: %s", result.Error)
		}
		return map[string]interface{}{
			"out_result": result.Output,
		}, nil
	}

	// Fallback to mock behavior
	return map[string]interface{}{
		"out_result": map[string]interface{}{
			"status":   "ok",
			"server":   server,
			"tool":     toolName,
			"output":   fmt.Sprintf("MCP call to %s.%s completed successfully", server, toolName),
			"duration": fmt.Sprintf("%dms", rand.Intn(1000)+100),
		},
	}, nil
}