// Package executors — Math & Logic Node Executors
//
// Pure Go implementations for simple computation nodes that don't need
// any external engine. These run entirely in-process with zero dependencies.
//
// Nodes covered:
//   - math.add     — numeric addition (supports float and int)
//   - math.multiply — numeric multiplication
//   - logic.compare — equality and comparison of two values
//   - logic.merge   — merge multiple inputs into a single output
package executors

import (
	"context"
	"fmt"
	"strconv"
)

// ============================================================================
// MathAddExecutor — adds two numeric values
// ============================================================================

// MathAddExecutor returns an executor for the math.add node.
// Steps:
//   1. Read "a" and "b" from inputs
//   2. Convert both to float64
//   3. Return a + b as "result"
func MathAddExecutor() func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		a := toFloat64(inputs["a"])
		b := toFloat64(inputs["b"])
		return map[string]interface{}{
			"result":  a + b,
			"status":  "completed",
			"message": fmt.Sprintf("%v + %v = %v", a, b, a+b),
		}, nil
	}
}

// ============================================================================
// MathMultiplyExecutor — multiplies two numeric values
// ============================================================================

// MathMultiplyExecutor returns an executor for the math.multiply node.
// Steps:
//   1. Read "a" and "b" from inputs
//   2. Convert both to float64
//   3. Return a * b as "result"
func MathMultiplyExecutor() func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		a := toFloat64(inputs["a"])
		b := toFloat64(inputs["b"])
		return map[string]interface{}{
			"result":  a * b,
			"status":  "completed",
			"message": fmt.Sprintf("%v * %v = %v", a, b, a*b),
		}, nil
	}
}

// ============================================================================
// LogicCompareExecutor — compares two values
// ============================================================================

// LogicCompareExecutor returns an executor for the logic.compare node.
// Steps:
//   1. Read "a" and "b" from inputs
//   2. Compare as strings and as float64
//   3. Return "equal" (bool), "greater" (bool)
func LogicCompareExecutor() func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		a := inputs["a"]
		b := inputs["b"]

		equal := fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
		greater := toFloat64(a) > toFloat64(b)

		return map[string]interface{}{
			"equal":   equal,
			"greater": greater,
			"a":       a,
			"b":       b,
			"status":  "completed",
			"message": fmt.Sprintf("compare: equal=%v, greater=%v", equal, greater),
		}, nil
	}
}

// ============================================================================
// LogicMergeExecutor — merges multiple inputs into one output
// ============================================================================

// LogicMergeExecutor returns an executor for the logic.merge node.
// Steps:
//   1. Collect all non-nil inputs (input1, input2, input3)
//   2. Return them merged under "output"
func LogicMergeExecutor() func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		merged := make(map[string]interface{})
		for k, v := range inputs {
			if v != nil {
				merged[k] = v
			}
		}
		return map[string]interface{}{
			"output":  merged,
			"count":   len(merged),
			"status":  "completed",
			"message": fmt.Sprintf("merged %d inputs", len(merged)),
		}, nil
	}
}

// ============================================================================
// NLP Tokenizer — basic Go text tokenization
// ============================================================================

// TokenizerExecutor returns an executor for the nlp.tokenizer node.
// Steps:
//   1. Read "text" from inputs
//   2. Split by whitespace into tokens
//   3. Return tokens list and IDs
func TokenizerExecutor() func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		text, _ := inputs["text"].(string)
		if text == "" {
			text = "empty input"
		}

		// Simple whitespace tokenizer
		var tokens []string
		current := ""
		for _, ch := range text {
			if ch == ' ' || ch == '\t' || ch == '\n' {
				if current != "" {
					tokens = append(tokens, current)
					current = ""
				}
			} else {
				current += string(ch)
			}
		}
		if current != "" {
			tokens = append(tokens, current)
		}

		ids := make([]int, len(tokens))
		for i := range ids {
			ids[i] = i
		}

		return map[string]interface{}{
			"tokens":     tokens,
			"ids":        ids,
			"tokenCount": len(tokens),
			"status":     "completed",
			"message":    fmt.Sprintf("tokenized %d tokens", len(tokens)),
		}, nil
	}
}

// ============================================================================
// Helpers
// ============================================================================

// toFloat64 safely converts an interface{} to float64.
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}
