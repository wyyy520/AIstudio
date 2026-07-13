package agent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// LLMConfig holds configuration for an LLM provider.
type LLMConfig struct {
	Provider   string // "openai", "claude", "gemini", "qwen", "zhipu", "deepseek"
	APIKey     string
	BaseURL    string
	Model      string
	MaxTokens  int
	Temperature float64
	Timeout    time.Duration
}

// DefaultLLMConfig returns a default LLM configuration.
func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		Provider:    "openai",
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-4o-mini",
		MaxTokens:   4096,
		Temperature: 0.7,
		Timeout:     30 * time.Second,
	}
}

// NewLLMProvider creates an LLM provider based on config.
func NewLLMProvider(cfg LLMConfig) (LLMProvider, error) {
	switch strings.ToLower(cfg.Provider) {
	case "openai":
		return NewOpenAIProvider(cfg), nil
	case "claude":
		return NewClaudeProvider(cfg), nil
	case "gemini":
		return NewGeminiProvider(cfg), nil
	case "qwen", "dashscope":
		return NewQwenProvider(cfg), nil
	case "zhipu", "glm":
		return NewZhipuProvider(cfg), nil
	case "deepseek":
		return NewDeepSeekProvider(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}
}

// ---- OpenAI Provider ----

type OpenAIProvider struct {
	cfg LLMConfig
}

func NewOpenAIProvider(cfg LLMConfig) *OpenAIProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4o-mini"
	}
	return &OpenAIProvider{cfg: cfg}
}

func (p *OpenAIProvider) Chat(ctx ChatContext, messages []Message) (*LLMResponse, error) {
	reqBody := map[string]interface{}{
		"model":       p.cfg.Model,
		"messages":    messages,
		"max_tokens":  p.cfg.MaxTokens,
		"temperature": p.cfg.Temperature,
	}

	return p.doRequest(ctx, reqBody)
}

func (p *OpenAIProvider) StreamChat(ctx ChatContext, messages []Message, callback func(chunk string)) error {
	reqBody := map[string]interface{}{
		"model":       p.cfg.Model,
		"messages":    messages,
		"max_tokens":  p.cfg.MaxTokens,
		"temperature": p.cfg.Temperature,
		"stream":      true,
	}

	return p.doStreamRequest(ctx, reqBody, callback)
}

func (p *OpenAIProvider) GenerateJSON(ctx ChatContext, messages []Message) (map[string]interface{}, error) {
	messages = append(messages, Message{
		Role: "system",
		Content: "Respond ONLY with valid JSON. No other text, no markdown, no explanation.",
	})

	resp, err := p.Chat(ctx, messages)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return result, nil
}

func (p *OpenAIProvider) doRequest(ctx ChatContext, reqBody map[string]interface{}) (*LLMResponse, error) {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := p.cfg.BaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)

	client := &http.Client{Timeout: p.cfg.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := openAIResp.Choices[0].Message.Content
	log.Printf("[llm] OpenAI response: %d chars", len(content))

	return &LLMResponse{
		Content: content,
		Raw:     content,
	}, nil
}

func (p *OpenAIProvider) doStreamRequest(ctx ChatContext, reqBody map[string]interface{}, callback func(chunk string)) error {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := p.cfg.BaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: p.cfg.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Read SSE stream
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			chunk := string(buf[:n])
			lines := strings.Split(chunk, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "data: ") {
					data := strings.TrimPrefix(line, "data: ")
					if data == "[DONE]" {
						return nil
					}
					var sseData struct {
						Choices []struct {
							Delta struct {
								Content string `json:"content"`
							} `json:"delta"`
						} `json:"choices"`
					}
					if err := json.Unmarshal([]byte(data), &sseData); err == nil {
						if len(sseData.Choices) > 0 {
							content := sseData.Choices[0].Delta.Content
							if content != "" {
								callback(content)
							}
						}
					}
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("stream read error: %w", err)
		}
	}
}

// ---- Claude Provider ----

type ClaudeProvider struct {
	cfg LLMConfig
}

func NewClaudeProvider(cfg LLMConfig) *ClaudeProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.anthropic.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "claude-3-5-sonnet-20241022"
	}
	return &ClaudeProvider{cfg: cfg}
}

func (p *ClaudeProvider) Chat(ctx ChatContext, messages []Message) (*LLMResponse, error) {
	// Convert messages to Claude format
	var systemMsg string
	var userMsgs []map[string]string

	for _, m := range messages {
		if m.Role == "system" {
			systemMsg = m.Content
		} else {
			role := "user"
			if m.Role == "assistant" {
				role = "assistant"
			}
			userMsgs = append(userMsgs, map[string]string{
				"role": role,
				"content": m.Content,
			})
		}
	}

	reqBody := map[string]interface{}{
		"model":         p.cfg.Model,
		"messages":      userMsgs,
		"max_tokens":    p.cfg.MaxTokens,
		"temperature":   p.cfg.Temperature,
		"system":        systemMsg,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := p.cfg.BaseURL + "/messages"
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: p.cfg.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var claudeResp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	content := claudeResp.Content[0].Text
	log.Printf("[llm] Claude response: %d chars", len(content))

	return &LLMResponse{
		Content: content,
		Raw:     content,
	}, nil
}

func (p *ClaudeProvider) StreamChat(ctx ChatContext, messages []Message, callback func(chunk string)) error {
	var systemMsg string
	var filteredMessages []Message
	for _, m := range messages {
		if m.Role == "system" {
			systemMsg = m.Content
		} else {
			filteredMessages = append(filteredMessages, m)
		}
	}

	reqBody := map[string]interface{}{
		"model":       p.cfg.Model,
		"messages":    convertToAnthropicMessages(filteredMessages),
		"max_tokens":  p.cfg.MaxTokens,
		"temperature": p.cfg.Temperature,
		"stream":      true,
	}

	if systemMsg != "" {
		reqBody["system"] = systemMsg
	}

	jsonData, _ := json.Marshal(reqBody)

	url := p.cfg.BaseURL + "/messages"
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: p.cfg.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				continue
			}

			var event struct {
				Type  string `json:"type"`
				Delta struct {
					Text string `json:"text"`
				} `json:"delta"`
			}
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			if event.Type == "content_block_delta" && event.Delta.Text != "" {
				callback(event.Delta.Text)
			}
		}
	}

	return scanner.Err()
}

func convertToAnthropicMessages(messages []Message) []map[string]interface{} {
	var result []map[string]interface{}
	for _, m := range messages {
		role := m.Role
		if role == "assistant" {
			role = "assistant"
		} else {
			role = "user"
		}
		result = append(result, map[string]interface{}{
			"role": role,
			"content": []map[string]string{
				{"type": "text", "text": m.Content},
			},
		})
	}
	return result
}

func (p *ClaudeProvider) GenerateJSON(ctx ChatContext, messages []Message) (map[string]interface{}, error) {
	messages = append(messages, Message{
		Role: "system",
		Content: "Respond ONLY with valid JSON. No other text, no markdown, no explanation.",
	})

	resp, err := p.Chat(ctx, messages)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return result, nil
}

// ---- Gemini Provider ----

type GeminiProvider struct {
	cfg LLMConfig
}

func NewGeminiProvider(cfg LLMConfig) *GeminiProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	if cfg.Model == "" {
		cfg.Model = "gemini-pro"
	}
	return &GeminiProvider{cfg: cfg}
}

func (p *GeminiProvider) Chat(ctx ChatContext, messages []Message) (*LLMResponse, error) {
	// Convert messages to Gemini format
	var contents []map[string]interface{}
	for _, m := range messages {
		if m.Role == "system" {
			continue // Gemini doesn't have system messages
		}
		role := "user"
		if m.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, map[string]interface{}{
			"role": role,
			"parts": []map[string]string{
				{"text": m.Content},
			},
		})
	}

	reqBody := map[string]interface{}{
		"contents": contents,
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": p.cfg.MaxTokens,
			"temperature":     p.cfg.Temperature,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.cfg.BaseURL, p.cfg.Model, p.cfg.APIKey)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: p.cfg.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	content := geminiResp.Candidates[0].Content.Parts[0].Text
	log.Printf("[llm] Gemini response: %d chars", len(content))

	return &LLMResponse{
		Content: content,
		Raw:     content,
	}, nil
}

func (p *GeminiProvider) StreamChat(ctx ChatContext, messages []Message, callback func(chunk string)) error {
	var contents []map[string]interface{}
	for _, m := range messages {
		if m.Role == "system" {
			continue
		}
		role := "user"
		if m.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, map[string]interface{}{
			"role": role,
			"parts": []map[string]string{
				{"text": m.Content},
			},
		})
	}

	reqBody := map[string]interface{}{
		"contents": contents,
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": p.cfg.MaxTokens,
			"temperature":     p.cfg.Temperature,
		},
	}

	jsonData, _ := json.Marshal(reqBody)

	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?key=%s&alt=sse", p.cfg.BaseURL, p.cfg.Model, p.cfg.APIKey)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: p.cfg.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			var geminiResp struct {
				Candidates []struct {
					Content struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
					} `json:"content"`
				} `json:"candidates"`
			}
			if err := json.Unmarshal([]byte(data), &geminiResp); err != nil {
				continue
			}

			if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
				text := geminiResp.Candidates[0].Content.Parts[0].Text
				if text != "" {
					callback(text)
				}
			}
		}
	}

	return scanner.Err()
}

func (p *GeminiProvider) GenerateJSON(ctx ChatContext, messages []Message) (map[string]interface{}, error) {
	messages = append(messages, Message{
		Role: "user",
		Content: "Respond ONLY with valid JSON. No other text, no markdown, no explanation.",
	})

	resp, err := p.Chat(ctx, messages)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return result, nil
}

// ---- Qwen Provider (DashScope) ----

type QwenProvider struct {
	cfg LLMConfig
}

func NewQwenProvider(cfg LLMConfig) *QwenProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "qwen-plus"
	}
	return &QwenProvider{cfg: cfg}
}

func (p *QwenProvider) Chat(ctx ChatContext, messages []Message) (*LLMResponse, error) {
	// Qwen uses OpenAI-compatible API
	openai := NewOpenAIProvider(p.cfg)
	return openai.Chat(ctx, messages)
}

func (p *QwenProvider) StreamChat(ctx ChatContext, messages []Message, callback func(chunk string)) error {
	openai := NewOpenAIProvider(p.cfg)
	return openai.StreamChat(ctx, messages, callback)
}

func (p *QwenProvider) GenerateJSON(ctx ChatContext, messages []Message) (map[string]interface{}, error) {
	openai := NewOpenAIProvider(p.cfg)
	return openai.GenerateJSON(ctx, messages)
}

// ---- Zhipu Provider (GLM) ----

type ZhipuProvider struct {
	cfg LLMConfig
}

func NewZhipuProvider(cfg LLMConfig) *ZhipuProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://open.bigmodel.cn/api/paas/v4"
	}
	if cfg.Model == "" {
		cfg.Model = "glm-4"
	}
	return &ZhipuProvider{cfg: cfg}
}

func (p *ZhipuProvider) Chat(ctx ChatContext, messages []Message) (*LLMResponse, error) {
	// Zhipu uses OpenAI-compatible API
	openai := NewOpenAIProvider(p.cfg)
	return openai.Chat(ctx, messages)
}

func (p *ZhipuProvider) StreamChat(ctx ChatContext, messages []Message, callback func(chunk string)) error {
	openai := NewOpenAIProvider(p.cfg)
	return openai.StreamChat(ctx, messages, callback)
}

func (p *ZhipuProvider) GenerateJSON(ctx ChatContext, messages []Message) (map[string]interface{}, error) {
	openai := NewOpenAIProvider(p.cfg)
	return openai.GenerateJSON(ctx, messages)
}

// ---- DeepSeek Provider ----

type DeepSeekProvider struct {
	cfg LLMConfig
}

func NewDeepSeekProvider(cfg LLMConfig) *DeepSeekProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.deepseek.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "deepseek-chat"
	}
	return &DeepSeekProvider{cfg: cfg}
}

func (p *DeepSeekProvider) Chat(ctx ChatContext, messages []Message) (*LLMResponse, error) {
	// DeepSeek uses OpenAI-compatible API
	openai := NewOpenAIProvider(p.cfg)
	return openai.Chat(ctx, messages)
}

func (p *DeepSeekProvider) StreamChat(ctx ChatContext, messages []Message, callback func(chunk string)) error {
	openai := NewOpenAIProvider(p.cfg)
	return openai.StreamChat(ctx, messages, callback)
}

func (p *DeepSeekProvider) GenerateJSON(ctx ChatContext, messages []Message) (map[string]interface{}, error) {
	openai := NewOpenAIProvider(p.cfg)
	return openai.GenerateJSON(ctx, messages)
}