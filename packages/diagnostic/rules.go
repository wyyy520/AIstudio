package diagnostic

import (
	"fmt"
	"strings"
	"time"
)

type Rule struct {
	ID        string
	Pattern   string
	Severity  Severity
	Summary   string
	Detail    string
	SuggestFn func(string) *FixSuggestion
}

func defaultRules() []Rule {
	return []Rule{
		{
			ID:       "cuda-oom",
			Pattern:  "CUDA out of memory",
			Severity: SeverityError,
			Summary:  "GPU out of memory (CUDA OOM)",
			Detail:   "The GPU ran out of memory during execution. This typically happens when batch_size is too large or the model is too big for the available GPU memory.",
			SuggestFn: func(log string) *FixSuggestion {
				return &FixSuggestion{
					Description: "Reduce batch size or use a smaller model to fit in GPU memory",
					Confidence:  0.9,
					AutoFixable: true,
					WorkflowChanges: []WorkflowChange{
						{Action: "update_config", Field: "batch_size", OldVal: nil, NewVal: "8"},
					},
				}
			},
		},
		{
			ID:       "module-not-found",
			Pattern:  "No module named",
			Severity: SeverityError,
			Summary:  "Missing Python dependency",
			Detail:   "A required Python package is not installed in the current environment.",
			SuggestFn: func(log string) *FixSuggestion {
				pkg := extractPackageName(log, "No module named")
				return &FixSuggestion{
					Description: fmt.Sprintf("Install the missing Python package: %s", pkg),
					Confidence:  0.95,
					AutoFixable: false,
				}
			},
		},
		{
			ID:       "file-not-found",
			Pattern:  "FileNotFoundError",
			Severity: SeverityError,
			Summary:  "File not found",
			Detail:   "The specified file path does not exist or cannot be accessed.",
			SuggestFn: func(log string) *FixSuggestion {
				return &FixSuggestion{
					Description: "Verify the file path exists and is accessible. Check dataset paths and configuration.",
					Confidence:  0.8,
					AutoFixable: false,
				}
			},
		},
		{
			ID:       "connection-refused",
			Pattern:  "Connection refused",
			Severity: SeverityError,
			Summary:  "Connection refused",
			Detail:   "A network connection was refused. The target service may not be running.",
			SuggestFn: func(log string) *FixSuggestion {
				return &FixSuggestion{
					Description: "Ensure the required service is running and the port is accessible",
					Confidence:  0.7,
					AutoFixable: false,
				}
			},
		},
		{
			ID:       "timeout",
			Pattern:  "Timeout",
			Severity: SeverityWarning,
			Summary:  "Operation timed out",
			Detail:   "An operation exceeded the configured timeout limit.",
			SuggestFn: func(log string) *FixSuggestion {
				return &FixSuggestion{
					Description: "Increase the timeout value or optimize the operation to complete faster",
					Confidence:  0.8,
					AutoFixable: true,
					WorkflowChanges: []WorkflowChange{
						{Action: "update_config", Field: "timeout", OldVal: nil, NewVal: "3600"},
					},
				}
			},
		},
		{
			ID:       "disk-full",
			Pattern:  "No space left on device",
			Severity: SeverityCritical,
			Summary:  "Disk space full",
			Detail:   "The disk has run out of space. Data cannot be written.",
			SuggestFn: func(log string) *FixSuggestion {
				return &FixSuggestion{
					Description: "Free up disk space by cleaning temporary files, old models, and logs",
					Confidence:  0.9,
					AutoFixable: false,
				}
			},
		},
		{
			ID:       "permission-denied",
			Pattern:  "Permission denied",
			Severity: SeverityError,
			Summary:  "Permission denied",
			Detail:   "The process does not have sufficient permissions to access the resource.",
			SuggestFn: func(log string) *FixSuggestion {
				return &FixSuggestion{
					Description: "Check file/directory permissions and ensure the user has appropriate access rights",
					Confidence:  0.8,
					AutoFixable: false,
				}
			},
		},
		{
			ID:       "training-divergence",
			Pattern:  "loss is NaN",
			Severity: SeverityCritical,
			Summary:  "Training loss diverged (NaN)",
			Detail:   "The training loss became NaN, indicating training instability or numerical issues.",
			SuggestFn: func(log string) *FixSuggestion {
				return &FixSuggestion{
					Description: "Reduce learning rate, add gradient clipping, or check for bad data in the dataset",
					Confidence:  0.85,
					AutoFixable: true,
					WorkflowChanges: []WorkflowChange{
						{Action: "update_config", Field: "lr", OldVal: nil, NewVal: "0.0001"},
						{Action: "update_config", Field: "gradient_clip", OldVal: nil, NewVal: 1.0},
					},
				}
			},
		},
		{
			ID:       "low-accuracy",
			Pattern:  "accuracy is low",
			Severity: SeverityWarning,
			Summary:  "Low model accuracy detected",
			Detail:   "The model evaluation shows lower than expected accuracy.",
			SuggestFn: func(log string) *FixSuggestion {
				return &FixSuggestion{
					Description: "Consider increasing epochs, adding data augmentation, or using a more powerful model architecture",
					Confidence:  0.6,
					AutoFixable: false,
				}
			},
		},
		{
			ID:       "cuda-version",
			Pattern:  "CUDA version mismatch",
			Severity: SeverityError,
			Summary:  "CUDA version mismatch",
			Detail:   "The installed CUDA version is incompatible with the required version.",
			SuggestFn: func(log string) *FixSuggestion {
				return &FixSuggestion{
					Description: "Install the correct CUDA version or use a compatible PyTorch/TensorFlow version",
					Confidence:  0.9,
					AutoFixable: false,
				}
			},
		},
		{
			ID:       "memory-leak",
			Pattern:  "memory leak",
			Severity: SeverityWarning,
			Summary:  "Possible memory leak detected",
			Detail:   "Memory usage is growing continuously, indicating a possible memory leak.",
			SuggestFn: func(log string) *FixSuggestion {
				return &FixSuggestion{
					Description: "Check for unclosed file handles, database connections, or large tensor accumulations in the training loop",
					Confidence:  0.5,
					AutoFixable: false,
				}
			},
		},
	}
}

func extractPackageName(log, prefix string) string {
	idx := strings.Index(log, prefix)
	if idx < 0 {
		return "unknown"
	}
	rest := log[idx+len(prefix):]
	rest = strings.TrimSpace(rest)
	rest = strings.Trim(rest, "'\"")
	if i := strings.IndexAny(rest, " ,.;:!?"); i > 0 {
		rest = rest[:i]
	}
	return rest
}

type perfRule struct {
	ID          string
	Condition   func(log string) bool
	Severity    Severity
	Summary     string
	SuggestFn   func(log string) *OptimizationSuggestion
}

func performanceRules() []perfRule {
	return []perfRule{
		{
			ID: "slow-data-loading",
			Condition: func(log string) bool {
				return strings.Contains(log, "loading") && strings.Contains(log, "slow")
			},
			Severity: SeverityWarning,
			Summary:  "Slow data loading detected",
			SuggestFn: func(log string) *OptimizationSuggestion {
				return &OptimizationSuggestion{
					Description:         "Data loading is a bottleneck. Consider using multi-threaded loading or pre-caching.",
					ExpectedImprovement: "Up to 3x faster training iterations",
					Confidence:          0.75,
				}
			},
		},
		{
			ID: "low-gpu-utilization",
			Condition: func(log string) bool {
				return strings.Contains(log, "GPU utilization") && (strings.Contains(log, "low") || strings.Contains(log, "< 50%"))
			},
			Severity: SeverityWarning,
			Summary:  "Low GPU utilization",
			SuggestFn: func(log string) *OptimizationSuggestion {
				return &OptimizationSuggestion{
					Description:         "GPU is underutilized. Increase batch size or use gradient accumulation.",
					ExpectedImprovement: "Up to 2x faster training",
					Confidence:          0.8,
					WorkflowChanges: []WorkflowChange{
						{Action: "update_config", Field: "batch_size", OldVal: nil, NewVal: "increase"},
					},
				}
			},
		},
		{
			ID: "high-iowait",
			Condition: func(log string) bool {
				return strings.Contains(log, "iowait") || strings.Contains(log, "I/O wait")
			},
			Severity: SeverityWarning,
			Summary:  "High I/O wait detected",
			SuggestFn: func(log string) *OptimizationSuggestion {
				return &OptimizationSuggestion{
					Description:         "High I/O wait indicates disk bottleneck. Consider using SSD, RAM disk, or optimizing data pipeline.",
					ExpectedImprovement: "Up to 50% reduction in training time",
					Confidence:          0.7,
				}
			},
		},
		{
			ID: "unnecessary-copy",
			Condition: func(log string) bool {
				return strings.Contains(log, "copy") && strings.Contains(log, "CPU")
			},
			Severity: SeverityWarning,
			Summary:  "Unnecessary CPU-GPU data copy detected",
			SuggestFn: func(log string) *OptimizationSuggestion {
				return &OptimizationSuggestion{
					Description:         "Frequent CPU-GPU copies are slowing down execution. Pin memory and use async data loading.",
					ExpectedImprovement: "Up to 30% faster execution",
					Confidence:          0.65,
				}
			},
		},
	}
}

func resourceRules() []Rule {
	return []Rule{
		{
			ID:       "high-memory-usage",
			Pattern:  "memory usage:",
			Severity: SeverityWarning,
			Summary:  "High memory usage",
			Detail:   "System memory usage is approaching capacity limits.",
			SuggestFn: func(log string) *FixSuggestion {
				return &FixSuggestion{
					Description: "Reduce batch size or enable memory optimization techniques",
					Confidence:  0.7,
					AutoFixable: true,
					WorkflowChanges: []WorkflowChange{
						{Action: "update_config", Field: "batch_size", OldVal: nil, NewVal: "16"},
					},
				}
			},
		},
	}
}

func formatTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}
