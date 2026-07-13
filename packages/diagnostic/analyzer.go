package diagnostic

import (
	"fmt"
	"time"

	"github.com/aistudio/packages/workflow"
	"github.com/google/uuid"
)

type Analyzer struct {
	rules     []Rule
	perfRules []perfRule
	resourceRules []Rule
	llm       LLMAnalyzer
	config    AnalyzerConfig
	nodeMap   map[string]string
}

func NewAnalyzer(config AnalyzerConfig) *Analyzer {
	a := &Analyzer{
		rules:         defaultRules(),
		perfRules:     performanceRules(),
		resourceRules: resourceRules(),
		config:        config,
		nodeMap:       make(map[string]string),
	}
	return a
}

func (a *Analyzer) WithLLM(llm LLMAnalyzer) *Analyzer {
	a.llm = llm
	return a
}

func (a *Analyzer) Analyze(entries []LogEntry) []AnalysisResult {
	return a.AnalyzeWithContext(entries, nil)
}

func (a *Analyzer) AnalyzeWithContext(entries []LogEntry, wf *workflow.Workflow) []AnalysisResult {
	start := time.Now()
	results := make([]AnalysisResult, 0)

	a.buildNodeMap(wf)

	for _, entry := range entries {
		result := a.analyzeEntry(entry, wf)
		if result != nil {
			if a.config.MinConfidence > 0 {
				if result.Suggestion != nil && result.Suggestion.Confidence < a.config.MinConfidence {
					continue
				}
			}
			results = append(results, *result)

			if a.config.MaxResults > 0 && len(results) >= a.config.MaxResults {
				break
			}
		}
	}

	optimizations := a.analyzePerformance(entries)
	for _, opt := range optimizations {
		if len(results) >= a.config.MaxResults {
			break
		}
		results = append(results, AnalysisResult{
			ID:            uuid.New().String(),
			Timestamp:     time.Now(),
			Severity:      SeverityInfo,
			Summary:       opt.Description,
			Detail:        fmt.Sprintf("Expected improvement: %s", opt.ExpectedImprovement),
			Optimization:  &opt,
			OriginalLog:   "",
		})
	}

	_ = time.Since(start)
	return results
}

func (a *Analyzer) analyzeEntry(entry LogEntry, wf *workflow.Workflow) *AnalysisResult {
	if a.llm != nil && a.config.EnableLLM {
		ctx := getWorkflowContextStr(wf)
		result, err := a.llm.AnalyzeLog(entry.Message+"; "+entry.Raw, ctx)
		if err == nil && result != nil {
			result.Timestamp = entry.Timestamp
			result.OriginalLog = entry.Raw
			if result.NodeID == "" {
				result.NodeID = a.mapToNode(entry.Message, wf)
				result.NodeName = a.nodeMap[result.NodeID]
			}
			return result
		}
	}

	msg := entry.Message
	raw := entry.Raw

	for _, rule := range a.rules {
		if containsSubstring(msg, rule.Pattern) || containsSubstring(raw, rule.Pattern) {
			nodeID := a.mapToNode(msg, wf)
			suggestion := rule.SuggestFn(raw)

			return &AnalysisResult{
				ID:          uuid.New().String(),
				Timestamp:   entry.Timestamp,
				Severity:    rule.Severity,
				Summary:     rule.Summary,
				Detail:      rule.Detail,
				NodeID:      nodeID,
				NodeName:    a.nodeMap[nodeID],
				Suggestion:  suggestion,
				OriginalLog: raw,
			}
		}
	}

	return nil
}

func (a *Analyzer) analyzePerformance(entries []LogEntry) []OptimizationSuggestion {
	var suggestions []OptimizationSuggestion

	for _, entry := range entries {
		for _, rule := range a.perfRules {
			if rule.Condition(entry.Message) || rule.Condition(entry.Raw) {
				suggestion := rule.SuggestFn(entry.Message)
				suggestions = append(suggestions, *suggestion)
			}
		}
	}

	return suggestions
}

func (a *Analyzer) TranslateLog(entry LogEntry, targetLang string) *LogTranslation {
	if a.llm != nil && a.config.EnableLLM {
		translated, err := a.llm.Translate(entry.Message, targetLang)
		if err == nil && translated != "" {
			return &LogTranslation{
				Original:   entry.Message,
				Translated: translated,
				Language:   targetLang,
			}
		}
	}

	return a.translateWithRules(entry, targetLang)
}

func (a *Analyzer) translateWithRules(entry LogEntry, targetLang string) *LogTranslation {
	translations := map[string]map[string]string{
		"zh": {
			"CUDA out of memory": "GPU 显存不足，请减小 batch_size 或使用更小的模型",
			"No module named":    "缺少 Python 包，请使用 pip install 安装",
			"FileNotFoundError":  "文件未找到，请检查路径是否正确",
			"Connection refused": "连接被拒绝，请检查服务是否正在运行",
			"Permission denied":  "权限不足，请检查文件权限",
			"Timeout":            "操作超时，请检查网络连接或增加超时时间",
			"SyntaxError":        "Python 语法错误，请检查代码",
			"KeyError":           "字典键不存在，请检查键名",
			"IndexError":         "列表索引越界，请检查索引值",
			"ValueError":         "数值错误，请检查参数值",
			"TypeError":          "类型错误，请检查参数类型",
		},
		"ja": {
			"CUDA out of memory": "GPUメモリ不足です。batch_sizeを減らすか、より小さいモデルを使用してください",
			"No module named":    "Pythonパッケージが不足しています。pip installを実行してください",
			"FileNotFoundError":  "ファイルが見つかりません。パスを確認してください",
		},
	}

	if trans, ok := translations[targetLang]; ok {
		for pattern, translation := range trans {
			if containsSubstring(entry.Message, pattern) {
				return &LogTranslation{
					Original:   entry.Message,
					Translated: translation,
					Language:   targetLang,
				}
			}
		}
	}

	return &LogTranslation{
		Original:   entry.Message,
		Translated: entry.Message,
		Language:   targetLang,
	}
}

func (a *Analyzer) SuggestFix(result AnalysisResult, wf *workflow.Workflow) *FixSuggestion {
	if a.llm != nil && a.config.EnableLLM {
		suggestion, err := a.llm.SuggestFix(result, wf)
		if err == nil && suggestion != nil {
			return suggestion
		}
	}

	return result.Suggestion
}

func (a *Analyzer) SuggestOptimization(wf *workflow.Workflow, metrics map[string]any) []OptimizationSuggestion {
	suggestions := make([]OptimizationSuggestion, 0)

	if numNodes, ok := metrics["node_count"].(int); ok && numNodes > 20 {
		suggestions = append(suggestions, OptimizationSuggestion{
			Description:         "Large number of nodes detected. Consider merging related processing steps.",
			ExpectedImprovement: "Simpler workflow, easier maintenance",
			Confidence:          0.6,
		})
	}

	if depth, ok := metrics["max_depth"].(int); ok && depth > 10 {
		suggestions = append(suggestions, OptimizationSuggestion{
			Description:         "Workflow pipeline is very deep (>10 levels). Consider parallelizing independent branches.",
			ExpectedImprovement: "Faster execution through parallelization",
			Confidence:          0.7,
			WorkflowChanges: []WorkflowChange{
				{Action: "restructure", Description: "Parallelize independent branches"},
			},
		})
	}

	if nodeCount, ok := metrics["node_count"].(int); ok && nodeCount > 0 {
		duplicateTypes := a.findDuplicateNodeTypes(wf)
		for nodeType, count := range duplicateTypes {
			if count > 3 {
				suggestions = append(suggestions, OptimizationSuggestion{
					Description:         fmt.Sprintf("Multiple %s nodes found (%d). Consider consolidating.", nodeType, count),
					ExpectedImprovement: "Reduced complexity, easier configuration",
					Confidence:          0.5,
				})
			}
		}
	}

	return suggestions
}

func (a *Analyzer) buildNodeMap(wf *workflow.Workflow) {
	a.nodeMap = make(map[string]string)
	if wf == nil {
		return
	}
	for _, node := range wf.Nodes {
		a.nodeMap[node.ID] = node.Name
	}
}

func (a *Analyzer) mapToNode(message string, wf *workflow.Workflow) string {
	if wf == nil {
		return ""
	}

	for _, node := range wf.Nodes {
		if containsSubstring(message, node.Name) || containsSubstring(message, node.ID) {
			return node.ID
		}
		for _, val := range node.Config {
			if str, ok := val.(string); ok {
				if containsSubstring(message, str) {
					return node.ID
				}
			}
		}
	}
	return ""
}

func (a *Analyzer) findDuplicateNodeTypes(wf *workflow.Workflow) map[string]int {
	counts := make(map[string]int)
	if wf == nil {
		return counts
	}
	for _, node := range wf.Nodes {
		key := string(node.Type)
		counts[key]++
	}
	return counts
}

func getWorkflowContextStr(wf *workflow.Workflow) string {
	if wf == nil {
		return ""
	}
	ctx := fmt.Sprintf("Workflow: %s (%s)\nTarget: %s\nNodes: %d\n",
		wf.Name, wf.ID, wf.Target, len(wf.Nodes))
	for _, node := range wf.Nodes {
		ctx += fmt.Sprintf("  - %s (%s): %s\n", node.Name, node.Type, node.Description)
	}
	return ctx
}