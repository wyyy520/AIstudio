package executors

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"text/scanner"
)

func ConditionExecutor() func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		expression, _ := config["expression"].(string)
		if expression == "" {
			return map[string]interface{}{
				"result": true,
				"branch": "true",
			}, nil
		}

		result, err := evalExpression(expression, inputs)
		if err != nil {
			return nil, fmt.Errorf("condition eval error: %w", err)
		}

		branch := "false"
		if result {
			branch = "true"
		}
		return map[string]interface{}{
			"result": result,
			"branch": branch,
		}, nil
	}
}

type exprToken struct {
	typ rune
	val string
}

func tokenize(expr string) []exprToken {
	var tokens []exprToken
	var s scanner.Scanner
	s.Init(strings.NewReader(expr))
	s.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats | scanner.ScanStrings

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		text := s.TokenText()
		switch {
		case tok == scanner.Int || tok == scanner.Float:
			tokens = append(tokens, exprToken{typ: 'n', val: text})
		case tok == scanner.String:
			tokens = append(tokens, exprToken{typ: 's', val: text})
		case tok == scanner.Ident:
			tokens = append(tokens, exprToken{typ: 'v', val: text})
		default:
			switch text {
			case ">=", "<=", "==", "!=", "&&", "||":
				tokens = append(tokens, exprToken{typ: 'o', val: text})
			case ">", "<":
				tokens = append(tokens, exprToken{typ: 'o', val: text})
			default:
				if text != " " {
					tokens = append(tokens, exprToken{typ: 'x', val: text})
				}
			}
		}
	}
	return tokens
}

func resolveValue(token string, inputs map[string]interface{}) (float64, string, bool) {
	if n, err := strconv.ParseFloat(token, 64); err == nil {
		return n, "", true
	}
	t := strings.Trim(token, "\"")
	if t != token {
		return 0, t, true
	}
	if val, ok := inputs[token]; ok {
		switch v := val.(type) {
		case float64:
			return v, "", true
		case int:
			return float64(v), "", true
		case string:
			return 0, v, true
		case bool:
			if v {
				return 1, "", true
			}
			return 0, "", true
		}
	}
	return 0, "", false
}

func evalExpression(expr string, inputs map[string]interface{}) (bool, error) {
	expr = strings.TrimSpace(expr)

	parts := strings.SplitN(expr, "||", 2)
	if len(parts) == 2 {
		left, err := evalExpression(strings.TrimSpace(parts[0]), inputs)
		if err != nil {
			return false, err
		}
		if left {
			return true, nil
		}
		return evalExpression(strings.TrimSpace(parts[1]), inputs)
	}

	parts = strings.SplitN(expr, "&&", 2)
	if len(parts) == 2 {
		left, err := evalExpression(strings.TrimSpace(parts[0]), inputs)
		if err != nil {
			return false, err
		}
		if !left {
			return false, nil
		}
		return evalExpression(strings.TrimSpace(parts[1]), inputs)
	}

	tokens := tokenize(expr)
	if len(tokens) == 1 && tokens[0].typ == 'v' {
		_, strVal, ok := resolveValue(tokens[0].val, inputs)
		if !ok {
			return false, fmt.Errorf("unknown variable: %s", tokens[0].val)
		}
		if strVal == "" || strVal == "false" || strVal == "0" {
			return false, nil
		}
		return true, nil
	}

	if len(tokens) >= 3 {
		left := tokens[0]
		op := tokens[1]
		right := tokens[2]

		if op.typ != 'o' {
			return false, fmt.Errorf("expected operator, got %s", op.val)
		}

		leftNum, leftStr, leftOk := resolveValue(left.val, inputs)
		rightNum, rightStr, rightOk := resolveValue(right.val, inputs)

		if !leftOk || !rightOk {
			return false, fmt.Errorf("cannot resolve values: %s %s %s", left.val, op.val, right.val)
		}

		if leftStr != "" || rightStr != "" {
			switch op.val {
			case "==":
				return leftStr == rightStr, nil
			case "!=":
				return leftStr != rightStr, nil
			default:
				return false, fmt.Errorf("operator %s not supported for string comparison", op.val)
			}
		}

		switch op.val {
		case ">":
			return leftNum > rightNum, nil
		case "<":
			return leftNum < rightNum, nil
		case "==":
			return leftNum == rightNum, nil
		case ">=":
			return leftNum >= rightNum, nil
		case "<=":
			return leftNum <= rightNum, nil
		case "!=":
			return leftNum != rightNum, nil
		default:
			return false, fmt.Errorf("unknown operator: %s", op.val)
		}
	}

	return false, fmt.Errorf("cannot parse expression: %s", expr)
}
