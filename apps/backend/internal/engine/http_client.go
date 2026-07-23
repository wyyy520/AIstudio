package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"
)

type httpClient struct {
	config    *Config
	client    *http.Client
	taskPath  string
}

// taskPayload matches the Engine Server's expected format:
//   {"task_id": "...", "plugin": "yolo", "action": "train", "params": {...}}
type taskPayload struct {
	TaskID  string      `json:"task_id"`
	Plugin  string      `json:"plugin"`
	Action  string      `json:"action"`
	Params  interface{} `json:"params"`
}

func NewClient(config *Config) EngineClient {
	if config == nil {
		config = DefaultConfig()
	}
	return &httpClient{
		config:   config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		taskPath: "/task",
	}
}

func (c *httpClient) Health(ctx context.Context) (*HealthResponse, error) {
	url := c.config.BaseURL + "/health"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create health request: %w", err)
	}

	var resp *http.Response
	err = c.retry(ctx, "health", func() error {
		var innerErr error
		resp, innerErr = c.client.Do(req)
		if innerErr != nil {
			return innerErr
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return fmt.Errorf("health returned status %d: %s", resp.StatusCode, string(body))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("health request failed: %w", err)
	}
	defer resp.Body.Close()

	var health HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("decode health response: %w", err)
	}
	return &health, nil
}

func (c *httpClient) Infer(ctx context.Context, req InferRequest) (*InferResponse, error) {
	payload := taskPayload{
		TaskID: req.TaskID,
		Plugin: req.Plugin,
		Action: "predict",
		Params: req.Params,
	}
	var result InferResponse
	if err := c.doTask(ctx, payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *httpClient) Train(ctx context.Context, req TrainRequest) (*TrainResponse, error) {
	payload := taskPayload{
		TaskID: req.TaskID,
		Plugin: req.Plugin,
		Action: "train",
		Params: req.Config,
	}
	var result TrainResponse
	if err := c.doTask(ctx, payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *httpClient) LoadModel(ctx context.Context, req LoadModelRequest) (*LoadModelResponse, error) {
	payload := taskPayload{
		TaskID: req.TaskID,
		Plugin: req.Plugin,
		Action: "load_model",
		Params: map[string]string{
			"model_name": req.ModelName,
			"model_path": req.ModelPath,
		},
	}
	var result LoadModelResponse
	if err := c.doTask(ctx, payload, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *httpClient) doTask(ctx context.Context, payload interface{}, result interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal task payload: %w", err)
	}

	url := c.config.BaseURL + c.taskPath

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create task request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	err = c.retry(ctx, "task", func() error {
		var innerErr error
		resp, innerErr = c.client.Do(req)
		if innerErr != nil {
			return innerErr
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return fmt.Errorf("task returned status %d: %s", resp.StatusCode, string(bodyBytes))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("task request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode task response: %w", err)
	}
	return nil
}

func (c *httpClient) retry(ctx context.Context, name string, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= c.config.RetryCount; attempt++ {
		if attempt > 0 {
			delay := time.Duration(float64(c.config.RetryDelay) * math.Pow(2, float64(attempt-1)))
			log.Printf("[engine] %s attempt %d/%d failed, retrying in %v: %v", name, attempt, c.config.RetryCount, delay, lastErr)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	log.Printf("[engine] %s all %d retries exhausted: %v", name, c.config.RetryCount, lastErr)
	return fmt.Errorf("%s failed after %d retries: %w", name, c.config.RetryCount, lastErr)
}
