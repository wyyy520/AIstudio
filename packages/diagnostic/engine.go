package diagnostic

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aistudio/packages/event"
	"github.com/aistudio/packages/workflow"
	"github.com/google/uuid"
)

type DiagnosticEngine struct {
	analyzer *Analyzer
	eventBus *event.EventBus
	config   AnalyzerConfig
}

func NewDiagnosticEngine(config AnalyzerConfig) *DiagnosticEngine {
	return &DiagnosticEngine{
		analyzer: NewAnalyzer(config),
		config:   config,
	}
}

func (e *DiagnosticEngine) WithLLM(llm LLMAnalyzer) *DiagnosticEngine {
	e.analyzer.WithLLM(llm)
	return e
}

func (e *DiagnosticEngine) WithEventBus(bus *event.EventBus) *DiagnosticEngine {
	e.eventBus = bus
	return e
}

func (e *DiagnosticEngine) Analyze(ctx context.Context, entries []LogEntry, wf *workflow.Workflow) (*DiagnosticReport, error) {
	start := time.Now()
	log.Printf("[diagnostic] analyzing %d log entries", len(entries))

	if len(entries) == 0 {
		return &DiagnosticReport{
			TaskID:     uuid.New().String(),
			Results:    make([]AnalysisResult, 0),
			Summary:    "No log entries to analyze",
			AnalyzedAt: time.Now(),
			Duration:   time.Since(start).String(),
		}, nil
	}

	results := e.analyzer.AnalyzeWithContext(entries, wf)

	if e.eventBus != nil {
		e.emitAnalysisEvents(results)
	}

	duration := time.Since(start)
	summary := buildSummary(results)

	log.Printf("[diagnostic] analysis complete: %d results in %s", len(results), duration)

	workflowID := ""
	if wf != nil {
		workflowID = wf.ID
	}

	return &DiagnosticReport{
		TaskID:     uuid.New().String(),
		WorkflowID: workflowID,
		Results:    results,
		Summary:    summary,
		AnalyzedAt: time.Now(),
		Duration:   duration.String(),
	}, nil
}

func (e *DiagnosticEngine) AnalyzeLog(ctx context.Context, entry LogEntry, wf *workflow.Workflow) (*AnalysisResult, error) {
	results := e.analyzer.AnalyzeWithContext([]LogEntry{entry}, wf)
	if len(results) == 0 {
		return nil, fmt.Errorf("no analysis results for log entry")
	}

	result := results[0]
	if e.eventBus != nil {
		e.eventBus.Publish(event.Topic("diagnostic:result"), result)
	}

	return &result, nil
}

func (e *DiagnosticEngine) Translate(ctx context.Context, entry LogEntry, targetLang string) (*LogTranslation, error) {
	translation := e.analyzer.TranslateLog(entry, targetLang)
	if translation == nil {
		return nil, fmt.Errorf("translation failed")
	}

	if e.eventBus != nil {
		e.eventBus.Publish(event.Topic("diagnostic:translation"), translation)
	}

	return translation, nil
}

func (e *DiagnosticEngine) SuggestFix(ctx context.Context, result AnalysisResult, wf *workflow.Workflow) (*FixSuggestion, error) {
	suggestion := e.analyzer.SuggestFix(result, wf)
	if suggestion == nil {
		return nil, fmt.Errorf("no fix suggestion available")
	}

	if e.eventBus != nil {
		e.eventBus.Publish(event.Topic("diagnostic:fix:suggested"), map[string]any{
			"result_id":  result.ID,
			"suggestion": suggestion,
		})
	}

	return suggestion, nil
}

func (e *DiagnosticEngine) SuggestOptimization(ctx context.Context, wf *workflow.Workflow, metrics map[string]any) ([]OptimizationSuggestion, error) {
	suggestions := e.analyzer.SuggestOptimization(wf, metrics)
	if len(suggestions) == 0 {
		return nil, fmt.Errorf("no optimization suggestions available")
	}

	if e.eventBus != nil {
		e.eventBus.Publish(event.Topic("diagnostic:optimization:suggested"), map[string]any{
			"workflow_id": wf.ID,
			"suggestions": suggestions,
		})
	}

	return suggestions, nil
}

func (e *DiagnosticEngine) RunDiagnosticPipeline(ctx context.Context, entries []LogEntry, wf *workflow.Workflow, metrics map[string]any) (*DiagnosticReport, error) {
	report, err := e.Analyze(ctx, entries, wf)
	if err != nil {
		return nil, err
	}

	if wf != nil && metrics != nil {
		optimizations, err := e.SuggestOptimization(ctx, wf, metrics)
		if err == nil {
			report.Optimizations = optimizations
		}
	}

	if e.eventBus != nil {
		e.eventBus.Publish(event.Topic("diagnostic:pipeline:completed"), report)
	}

	return report, nil
}

func (e *DiagnosticEngine) emitAnalysisEvents(results []AnalysisResult) {
	for _, result := range results {
		topic := event.Topic("diagnostic:analysis")
		switch result.Severity {
		case SeverityCritical:
			topic = "diagnostic:critical"
		case SeverityError:
			topic = "diagnostic:error"
		case SeverityWarning:
			topic = "diagnostic:warning"
		case SeverityInfo:
			topic = "diagnostic:info"
		}
		e.eventBus.Publish(topic, result)
	}
}

func buildSummary(results []AnalysisResult) string {
	if len(results) == 0 {
		return "No issues detected"
	}

	errCount := 0
	warnCount := 0
	infoCount := 0
	criticalCount := 0

	for _, r := range results {
		switch r.Severity {
		case SeverityCritical:
			criticalCount++
		case SeverityError:
			errCount++
		case SeverityWarning:
			warnCount++
		case SeverityInfo:
			infoCount++
		}
	}

	parts := make([]string, 0)
	if criticalCount > 0 {
		parts = append(parts, fmt.Sprintf("%d critical", criticalCount))
	}
	if errCount > 0 {
		parts = append(parts, fmt.Sprintf("%d errors", errCount))
	}
	if warnCount > 0 {
		parts = append(parts, fmt.Sprintf("%d warnings", warnCount))
	}
	if infoCount > 0 {
		parts = append(parts, fmt.Sprintf("%d info", infoCount))
	}

	return fmt.Sprintf("Found %s", joinParts(parts))
}

func joinParts(parts []string) string {
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	case 2:
		return parts[0] + " and " + parts[1]
	default:
		result := ""
		for i, p := range parts {
			if i == len(parts)-1 {
				result += "and " + p
			} else {
				result += p + ", "
			}
		}
		return result
	}
}

func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}