package engine

import "context"

type LogFunc func(format string, args ...interface{})

type EngineClientWithLogging struct {
	client EngineClient
	logFn  LogFunc
}

func NewClientWithLogging(config *Config, logFn LogFunc) EngineClient {
	return &EngineClientWithLogging{
		client: NewClient(config),
		logFn:  logFn,
	}
}

func (c *EngineClientWithLogging) Infer(ctx context.Context, req InferRequest) (*InferResponse, error) {
	if c.logFn != nil {
		c.logFn("[engine] Infer: model=%s", req.ModelName)
	}
	return c.client.Infer(ctx, req)
}

func (c *EngineClientWithLogging) Train(ctx context.Context, req TrainRequest) (*TrainResponse, error) {
	if c.logFn != nil {
		c.logFn("[engine] Train: model=%s dataset=%s", req.ModelName, req.Dataset)
	}
	return c.client.Train(ctx, req)
}

func (c *EngineClientWithLogging) Health(ctx context.Context) (*HealthResponse, error) {
	if c.logFn != nil {
		c.logFn("[engine] Health check")
	}
	return c.client.Health(ctx)
}

func (c *EngineClientWithLogging) LoadModel(ctx context.Context, req LoadModelRequest) (*LoadModelResponse, error) {
	if c.logFn != nil {
		c.logFn("[engine] LoadModel: model=%s path=%s", req.ModelName, req.ModelPath)
	}
	return c.client.LoadModel(ctx, req)
}
