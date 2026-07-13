package diagnostic

import (
	"time"

	"github.com/aistudio/packages/workflow"
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

type AnalysisResult struct {
	ID              string        `json:"id"`
	Timestamp       time.Time     `json:"timestamp"`
	Severity        Severity      `json:"severity"`
	Summary         string        `json:"summary"`
	Detail          string        `json:"detail"`
	NodeID          string        `json:"node_id,omitempty"`
	NodeName        string        `json:"node_name,omitempty"`
	Suggestion      *FixSuggestion `json:"suggestion,omitempty"`
	OriginalLog     string        `json:"original_log"`
	Optimization    *OptimizationSuggestion `json:"optimization,omitempty"`
}

type LogTranslation struct {
	Original   string `json:"original"`
	Translated string `json:"translated"`
	Language   string `json:"language"`
}

type FixSuggestion struct {
	Description      string              `json:"description"`
	WorkflowChanges  []WorkflowChange    `json:"workflow_changes,omitempty"`
	Confidence       float64             `json:"confidence"`
	AutoFixable      bool                `json:"auto_fixable"`
}

type WorkflowChange struct {
	Action      string `json:"action"`
	NodeID      string `json:"node_id,omitempty"`
	Field       string `json:"field,omitempty"`
	OldVal      any    `json:"old_value,omitempty"`
	NewVal      any    `json:"new_value,omitempty"`
	Description string `json:"description,omitempty"`
}

type OptimizationSuggestion struct {
	Description        string           `json:"description"`
	ExpectedImprovement string          `json:"expected_improvement"`
	WorkflowChanges    []WorkflowChange `json:"workflow_changes,omitempty"`
	Confidence         float64          `json:"confidence"`
}

type LogEntry struct {
	Timestamp  time.Time      `json:"timestamp"`
	Level      string         `json:"level"`
	Source     string         `json:"source"`
	Message    string         `json:"message"`
	Raw        string         `json:"raw,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	NodeID     string         `json:"node_id,omitempty"`
	TaskID     string         `json:"task_id,omitempty"`
}

type AnalyzerConfig struct {
	EnableLLM    bool    `json:"enable_llm"`
	LLMProvider  string  `json:"llm_provider,omitempty"`
	LLMModel     string  `json:"llm_model,omitempty"`
	LLMAPIKey    string  `json:"-"`
	MaxResults   int     `json:"max_results"`
	MinConfidence float64 `json:"min_confidence"`
}

func DefaultConfig() AnalyzerConfig {
	return AnalyzerConfig{
		EnableLLM:     false,
		MaxResults:    50,
		MinConfidence: 0.3,
	}
}

type DiagnosticReport struct {
	TaskID       string            `json:"task_id"`
	WorkflowID   string            `json:"workflow_id,omitempty"`
	Results      []AnalysisResult  `json:"results"`
	Optimizations []OptimizationSuggestion `json:"optimizations,omitempty"`
	Summary      string            `json:"summary"`
	AnalyzedAt   time.Time         `json:"analyzed_at"`
	Duration     string            `json:"duration,omitempty"`

	results []AnalysisResult `json:"-"`
}

type LLMAnalyzer interface {
	AnalyzeLog(logContent string, workflowContext string) (*AnalysisResult, error)
	Translate(message string, targetLang string) (string, error)
	SuggestFix(result AnalysisResult, wf *workflow.Workflow) (*FixSuggestion, error)
}