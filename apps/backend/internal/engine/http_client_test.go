package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("expected path /health, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HealthResponse{Status: "ok", Version: "1.0.0"})
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		RetryCount: 0,
		RetryDelay: 0,
	})

	resp, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() failed: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp.Status)
	}
	if resp.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", resp.Version)
	}
}

func TestHealthEndpointError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":"service unavailable"}`))
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		RetryCount: 0,
		RetryDelay: 0,
	})

	_, err := client.Health(context.Background())
	if err == nil {
		t.Fatal("expected error for non-200 health response")
	}
}

func TestTaskEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/task" {
			t.Errorf("expected path /task, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var payload taskPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if payload.Action != "infer" {
			t.Errorf("expected action 'infer', got '%s'", payload.Action)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(InferResponse{Result: "test-result", Confidence: 0.95})
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		RetryCount: 0,
		RetryDelay: 0,
	})

	resp, err := client.Infer(context.Background(), InferRequest{
		ModelName: "test-model",
		Input:     map[string]interface{}{"image": "data"},
	})
	if err != nil {
		t.Fatalf("Infer() failed: %v", err)
	}
	if resp.Result != "test-result" {
		t.Errorf("expected result 'test-result', got '%s'", resp.Result)
	}
	if resp.Confidence != 0.95 {
		t.Errorf("expected confidence 0.95, got %f", resp.Confidence)
	}
}

func TestTaskEndpointBadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid request"}`))
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		RetryCount: 0,
		RetryDelay: 0,
	})

	_, err := client.Infer(context.Background(), InferRequest{ModelName: "test"})
	if err == nil {
		t.Fatal("expected error for bad request")
	}
}

func TestTimeoutHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL:    server.URL,
		Timeout:    100 * time.Millisecond,
		RetryCount: 0,
		RetryDelay: 0,
	})

	_, err := client.Health(context.Background())
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRetryLogic(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(HealthResponse{Status: "ok", Version: "1.0"})
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		RetryCount: 3,
		RetryDelay: 10 * time.Millisecond,
	})

	resp, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() with retry failed: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp.Status)
	}
	if attempts != 3 {
		t.Errorf("expected 3 total attempts, got %d", attempts)
	}
}

func TestRetryExhaustion(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		RetryCount: 2,
		RetryDelay: 10 * time.Millisecond,
	})

	_, err := client.Health(context.Background())
	if err == nil {
		t.Fatal("expected error after retry exhaustion")
	}
	if attempts != 3 {
		t.Errorf("expected 3 total attempts (1 initial + 2 retries), got %d", attempts)
	}
}

func TestLoadModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(LoadModelResponse{Success: true})
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		RetryCount: 0,
		RetryDelay: 0,
	})

	resp, err := client.LoadModel(context.Background(), LoadModelRequest{
		ModelName: "yolo",
		ModelPath: "/models/test.pt",
	})
	if err != nil {
		t.Fatalf("LoadModel() failed: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success=true, got %v", resp.Success)
	}
}

func TestTrain(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(TrainResponse{
			ModelPath:  "/models/trained.pt",
			DurationMs: 1000,
		})
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		RetryCount: 0,
		RetryDelay: 0,
	})

	resp, err := client.Train(context.Background(), TrainRequest{
		Dataset:   "coco128",
		ModelName: "yolov8n",
	})
	if err != nil {
		t.Fatalf("Train() failed: %v", err)
	}
	if resp.ModelPath != "/models/trained.pt" {
		t.Errorf("expected model path '/models/trained.pt', got '%s'", resp.ModelPath)
	}
	if resp.DurationMs != 1000 {
		t.Errorf("expected duration 1000ms, got %d", resp.DurationMs)
	}
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(&Config{
		BaseURL:    server.URL,
		Timeout:    10 * time.Second,
		RetryCount: 0,
		RetryDelay: 0,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Health(ctx)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.BaseURL != "http://localhost:8082" {
		t.Errorf("expected BaseURL 'http://localhost:8082', got '%s'", cfg.BaseURL)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", cfg.Timeout)
	}
	if cfg.RetryCount != 3 {
		t.Errorf("expected retry count 3, got %d", cfg.RetryCount)
	}
	if cfg.RetryDelay != 1*time.Second {
		t.Errorf("expected retry delay 1s, got %v", cfg.RetryDelay)
	}
}
