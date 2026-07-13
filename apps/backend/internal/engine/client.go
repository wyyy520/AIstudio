package engine

import "context"

type EngineClient interface {
	Infer(ctx context.Context, req InferRequest) (*InferResponse, error)
	Train(ctx context.Context, req TrainRequest) (*TrainResponse, error)
	Health(ctx context.Context) (*HealthResponse, error)
	LoadModel(ctx context.Context, req LoadModelRequest) (*LoadModelResponse, error)
}

type InferRequest struct {
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
	ModelName string `json:"model_name"`
	ModelPath string `json:"model_path"`
}

type LoadModelResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
