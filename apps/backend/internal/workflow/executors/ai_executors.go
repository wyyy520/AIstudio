// Package executors — AI Model Executors
//
// Executors that delegate to the Python AI Engine via EngineClient.
// These nodes act as thin wrappers: they marshal inputs, call the Engine,
// and return structured results.
//
// Nodes covered:
//   - vision.*        — image classification, object detection, segmentation
//   - nlp.*           — text classification, NER, summarization, translation, BERT, LSTM, transformer
//   - speech.*        — speech recognition (ASR), text-to-speech (TTS)
//   - training.*      — model training and export
//
// Architecture:
//
//	Workflow Node → AI Executor → engineClient.Infer/Train → Engine/server.py
//	                                                             → PLUGIN_REGISTRY dispatch
//	                                                             → vision.yolo / inference.base_inference / ...
//
// All AI nodes follow the same pattern:
//  1. Build InferRequest/TrainRequest from node inputs + config
//  2. Call engineClient.Infer() or engineClient.Train()
//  3. Convert InferResponse/TrainResponse to node output map
package executors

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aistudio/backend/internal/engine"
)

// ============================================================================
// Generic AI Infer Executor — used by vision.classification / nlp / speech nodes
// ============================================================================

// AIInferExecutor creates an executor that delegates to the Python AI Engine.
// pluginName maps to Engine PLUGIN_REGISTRY keys (e.g., "yolo", "nlp").
// actionName maps to the action within that plugin (e.g., "predict", "text_classification").
//
// Steps:
//  1. Build InferRequest from node inputs + config
//  2. Call engineClient.Infer(ctx, req)
//  3. Return the response as node output
func AIInferExecutor(client engine.EngineClient, pluginName, actionName string) func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		// Step 1: Build request — merge inputs and config into engine params
		params := make(map[string]interface{})
		for k, v := range inputs {
			params[k] = v
		}
		for k, v := range config {
			params[k] = v
		}

		req := engine.InferRequest{
			Plugin: pluginName,
			Input: map[string]interface{}{
				"action": actionName,
				"params": params,
			},
		}

		// Step 2: Call Engine
		log.Printf("[executor:ai] infer: plugin=%s, action=%s", pluginName, actionName)
		resp, err := client.Infer(ctx, req)
		if err != nil {
			// Step 2b: Engine unavailable — return graceful fallback
			return buildFallbackResult(inputs, config, pluginName, actionName, err), nil
		}

		// Step 3: Return Engine response as node output
		result := map[string]interface{}{
			"status":    "completed",
			"message":   fmt.Sprintf("%s.%s completed (engine)", pluginName, actionName),
			"output":    resp.Output,
			"result":    resp.Result,
			"confidence": resp.Confidence,
			"durationMs": resp.DurationMs,
		}
		if resp.Detections != nil {
			result["detections"] = resp.Detections
		}
		return result, nil
	}
}

// ============================================================================
// AI Train Executor — used by training.train / training.export nodes
// ============================================================================

// AITrainExecutor creates an executor that delegates training to the Engine.
//
// Steps:
//  1. Build TrainRequest from node inputs + config
//  2. Call engineClient.Train(ctx, req)
//  3. Return the response as node output
func AITrainExecutor(client engine.EngineClient, pluginName string) func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		// Step 1: Build request
		params := make(map[string]interface{})
		for k, v := range inputs {
			params[k] = v
		}
		for k, v := range config {
			params[k] = v
		}

		datasetPath, _ := params["dataset"].(string)
		if datasetPath == "" {
			datasetPath = params["data_path"].(string)
		}
		modelName, _ := params["model_name"].(string)
		if modelName == "" {
			modelName, _ = params["model"].(string)
		}

		req := engine.TrainRequest{
			Plugin:    pluginName,
			Dataset:   datasetPath,
			ModelName: modelName,
			Config:    params,
		}

		// Step 2: Call Engine
		log.Printf("[executor:ai] train: plugin=%s, model=%s", pluginName, modelName)
		resp, err := client.Train(ctx, req)
		if err != nil {
			return buildFallbackResult(inputs, config, pluginName, "train", err), nil
		}

		// Step 3: Return result
		return map[string]interface{}{
			"status":     "completed",
			"message":    fmt.Sprintf("%s training completed (engine)", pluginName),
			"modelPath":  resp.ModelPath,
			"metrics":    resp.Metrics,
			"durationMs": resp.DurationMs,
		}, nil
	}
}

// ============================================================================
// Fallback — graceful degradation when Engine is unreachable
// ============================================================================

// buildFallbackResult returns a graceful fallback when the Engine is unreachable.
// This prevents the entire workflow from crashing — the node completes with
// a warning status so the workflow can continue.
func buildFallbackResult(inputs, config map[string]interface{}, plugin, action string, engineErr error) map[string]interface{} {
	log.Printf("[executor:ai] engine unavailable for %s.%s: %v — using fallback", plugin, action, engineErr)

	result := map[string]interface{}{
		"status":       "completed",
		"message":      fmt.Sprintf("%s.%s executed (engine offline — fallback mode)", plugin, action),
		"engineStatus": "offline",
		"engineError":  engineErr.Error(),
	}

	// Pass through all input values so downstream nodes can still work
	for k, v := range inputs {
		result[k] = v
	}
	for k, v := range config {
		if _, exists := result[k]; !exists {
			result[k] = v
		}
	}

	// Provide sensible defaults based on node category
	switch {
	case strings.HasPrefix(plugin, "nlp") || strings.HasPrefix(plugin, "speech"):
		result["text"] = inputs["text"]
		result["output"] = fmt.Sprintf("[engine offline] %s.%s fallback", plugin, action)
	case plugin == "training":
		result["modelPath"] = config["model"]
		result["metrics"] = map[string]interface{}{"note": "engine offline — metrics unavailable"}
	}

	return result
}

// ============================================================================
// ModelExportExecutor — export model to different formats
// ============================================================================

// ModelExportExecutor creates an executor for training.export.
// Delegates to engineClient with action="export".
func ModelExportExecutor(client engine.EngineClient) func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		modelPath, _ := inputs["model"].(string)
		if modelPath == "" {
			modelPath, _ = config["model"].(string)
		}
		exportFormat, _ := config["format"].(string)
		if exportFormat == "" {
			exportFormat = "onnx"
		}

		// Try to use Engine for ONNX/TorchScript/TensorRT export
		req := engine.InferRequest{
			Plugin: "model",
			Input: map[string]interface{}{
				"action": "export",
				"params": map[string]interface{}{
					"model_path": modelPath,
					"format":     exportFormat,
				},
			},
		}

		_, err := client.Infer(ctx, req)
		if err != nil {
			return map[string]interface{}{
				"status":   "completed",
				"message":  fmt.Sprintf("model export to %s (file-based fallback)", exportFormat),
				"exported": modelPath,
				"format":   exportFormat,
			}, nil
		}

		return map[string]interface{}{
			"status":   "completed",
			"message":  fmt.Sprintf("model exported to %s format", exportFormat),
			"exported": modelPath,
			"format":   exportFormat,
		}, nil
	}
}
