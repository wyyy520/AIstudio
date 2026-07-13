package agent

import (
	"context"
	"encoding/json"
	"testing"
)

// MockLLMProvider is a mock LLM provider for testing.
type MockLLMProvider struct {
	ChatResponse     string
	GenerateJSONResp map[string]interface{}
}

func (m *MockLLMProvider) Chat(ctx ChatContext, messages []Message) (*LLMResponse, error) {
	return &LLMResponse{
		Content: m.ChatResponse,
		Raw:     m.ChatResponse,
	}, nil
}

func (m *MockLLMProvider) StreamChat(ctx ChatContext, messages []Message, callback func(chunk string)) error {
	callback(m.ChatResponse)
	return nil
}

func (m *MockLLMProvider) GenerateJSON(ctx ChatContext, messages []Message) (map[string]interface{}, error) {
	if m.GenerateJSONResp != nil {
		return m.GenerateJSONResp, nil
	}
	return map[string]interface{}{
		"goal":        "test goal",
		"explanation": "test explanation",
	}, nil
}

func TestAgent_Process(t *testing.T) {
	mockLLM := &MockLLMProvider{
		ChatResponse: "I'll help you train a vehicle detection model.",
	}

	mem, err := NewMemoryInMemory()
	if err != nil {
		t.Fatalf("failed to create memory: %v", err)
	}

	ag := NewAgent(mockLLM, mem)

	// Register test tools
	ag.tools.Register(&CheckEnvironmentTool{
		CheckFn: func(ctx context.Context) (map[string]interface{}, error) {
			return map[string]interface{}{
				"python": true,
				"cuda":   true,
			}, nil
		},
	})

	ag.tools.Register(&CreateWorkflowTool{
		CreateFn: func(ctx context.Context, name string, workflowJSON json.RawMessage) (string, error) {
			return "wf_test", nil
		},
	})

	ag.tools.Register(&RunWorkflowTool{
		RunFn: func(ctx context.Context, workflowID string, params map[string]interface{}) (string, error) {
			return "task_test", nil
		},
	})

	ctx := context.Background()
	resp, err := ag.Process(ctx, "帮我训练一个车辆检测模型", "proj_1", "user_1", []string{"yolo"}, "{}")
	if err != nil {
		t.Fatalf("process failed: %v", err)
	}

	t.Logf("Goal: %s", resp.Goal)
	t.Logf("Explanation: %s", resp.Explanation)
	t.Logf("Status: %s", resp.Status)
	t.Logf("Steps: %d", len(resp.Steps))

	if resp.Goal == "" {
		t.Error("goal should not be empty")
	}

	if len(resp.Steps) == 0 {
		t.Error("should have executed some steps")
	}
}

func TestAgent_PlanOnly(t *testing.T) {
	mockLLM := &MockLLMProvider{
		ChatResponse: "I'll help you.",
	}

	mem, err := NewMemoryInMemory()
	if err != nil {
		t.Fatalf("failed to create memory: %v", err)
	}

	ag := NewAgent(mockLLM, mem)

	ctx := context.Background()
	plan, err := ag.PlanOnly(ctx, "帮我训练一个车辆检测模型", "proj_1", "user_1", []string{"yolo"}, "{}")
	if err != nil {
		t.Fatalf("plan only failed: %v", err)
	}

	t.Logf("Goal: %s", plan.Goal)
	t.Logf("Steps: %d", len(plan.Steps))

	if plan.Goal == "" {
		t.Error("goal should not be empty")
	}
}

func TestLLMProvider_OpenAI(t *testing.T) {
	// This test requires a valid API key, so skip by default
	t.Skip("requires API key")

	cfg := LLMConfig{
		Provider: "openai",
		APIKey:   "sk-test",
		Model:    "gpt-4o-mini",
	}

	provider := NewOpenAIProvider(cfg)
	if provider == nil {
		t.Fatal("failed to create OpenAI provider")
	}
}

func TestLLMProvider_Claude(t *testing.T) {
	cfg := LLMConfig{
		Provider: "claude",
		APIKey:   "test-key",
		Model:    "claude-3-5-sonnet-20241022",
	}

	provider := NewClaudeProvider(cfg)
	if provider == nil {
		t.Fatal("failed to create Claude provider")
	}
}

func TestLLMProvider_Gemini(t *testing.T) {
	cfg := LLMConfig{
		Provider: "gemini",
		APIKey:   "test-key",
		Model:    "gemini-pro",
	}

	provider := NewGeminiProvider(cfg)
	if provider == nil {
		t.Fatal("failed to create Gemini provider")
	}
}

func TestLLMProvider_Qwen(t *testing.T) {
	cfg := LLMConfig{
		Provider: "qwen",
		APIKey:   "test-key",
		Model:    "qwen-plus",
	}

	provider := NewQwenProvider(cfg)
	if provider == nil {
		t.Fatal("failed to create Qwen provider")
	}
}

func TestLLMProvider_Zhipu(t *testing.T) {
	cfg := LLMConfig{
		Provider: "zhipu",
		APIKey:   "test-key",
		Model:    "glm-4",
	}

	provider := NewZhipuProvider(cfg)
	if provider == nil {
		t.Fatal("failed to create Zhipu provider")
	}
}

func TestLLMProvider_DeepSeek(t *testing.T) {
	cfg := LLMConfig{
		Provider: "deepseek",
		APIKey:   "test-key",
		Model:    "deepseek-chat",
	}

	provider := NewDeepSeekProvider(cfg)
	if provider == nil {
		t.Fatal("failed to create DeepSeek provider")
	}
}

func TestNewLLMProvider(t *testing.T) {
	tests := []struct {
		provider string
		wantErr  bool
	}{
		{"openai", false},
		{"claude", false},
		{"gemini", false},
		{"qwen", false},
		{"zhipu", false},
		{"deepseek", false},
		{"unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			cfg := LLMConfig{
				Provider: tt.provider,
				APIKey:   "test-key",
			}

			_, err := NewLLMProvider(cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLLMProvider() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}