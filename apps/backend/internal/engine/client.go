// Package engine provides the Go ↔ Python AI Engine bridge.
//
// The EngineClient is the only way for the Go backend to communicate with
// the Python Engine (Engine/server.py). It sends HTTP requests for AI tasks:
// inference, training, model loading, and health checks.
//
// Flow:
//
//	Go Backend (Workflow Runtime)
//	  → EngineClient.Infer/Train()
//	  → HTTP POST to Engine/server.py:/task
//	  → Engine dispatches to plugin handler (YOLO, NLP, data, model)
//	  → Returns InferResponse/TrainResponse
//
// The Engine runs as a separate process (Python HTTP server on :8082),
// which allows the Go backend to remain lightweight and fast.
//
// EngStudio.md §2 — Engine Integration
package engine

import "context"

// ============================================================================
// EngineClient Interface — the bridge to Python AI Engine
// ============================================================================

// EngineClient is the interface for communicating with the Python AI Engine.
// It sends AI task requests (inference, training, model operations) and
// receives structured responses.
//
// Implementation: engine.HTTPClient (sends HTTP POST to Engine/server.py)
type EngineClient interface {
	// Infer runs model inference on the Engine.
	// The Plugin field determines which AI model to use (e.g., "yolo", "nlp").
	// Input contains the data to process (images, text, etc.).
	Infer(ctx context.Context, req InferRequest) (*InferResponse, error)

	// Train starts a training job on the Engine.
	// Dataset is a path to training data, Config contains hyperparameters.
	Train(ctx context.Context, req TrainRequest) (*TrainResponse, error)

	// Health checks if the Engine is running and responsive.
	Health(ctx context.Context) (*HealthResponse, error)

	// LoadModel preloads a model into the Engine's memory for faster inference.
	LoadModel(ctx context.Context, req LoadModelRequest) (*LoadModelResponse, error)
}

// ============================================================================
// Request/Response Types
// ============================================================================

type InferRequest struct {
	TaskID    string                 `json:"task_id"`
	Plugin    string                 `json:"plugin"`
	ModelName string                 `json:"model_name"`
	Input     map[string]interface{} `json:"input"`
	Params    map[string]interface{} `json:"params,omitempty"`
}

type InferResponse struct {
	Output     map[string]interface{} `json:"output,omitempty"`
	Result     string                 `json:"result,omitempty"`
	Confidence float64                `json:"confidence,omitempty"`
	Detections []interface{}          `json:"detections,omitempty"`
	DurationMs int64                  `json:"duration_ms"`
	Error      string                 `json:"error,omitempty"`
}

type TrainRequest struct {
	TaskID    string                 `json:"task_id"`
	Plugin    string                 `json:"plugin"`
	Dataset   string                 `json:"dataset"`
	Config    map[string]interface{} `json:"config"`
	ModelName string                 `json:"model_name"`
}

type TrainResponse struct {
	ModelPath  string                 `json:"model_path"`
	Metrics    map[string]interface{} `json:"metrics,omitempty"`
	DurationMs int64                  `json:"duration_ms"`
	Error      string                 `json:"error,omitempty"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  int64  `json:"uptime"`
}

type LoadModelRequest struct {
	TaskID    string `json:"task_id"`
	Plugin    string `json:"plugin"`
	ModelName string `json:"model_name"`
	ModelPath string `json:"model_path"`
}

type LoadModelResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
