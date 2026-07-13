package executors

import (
	"context"
	"time"
)

func RetryExecutor(inner func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error)) func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	if inner == nil {
		inner = func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
			return map[string]interface{}{"status": "completed"}, nil
		}
	}

	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		maxRetries := 3
		if v, ok := config["maxRetries"].(float64); ok {
			maxRetries = int(v)
		}

		backoffMs := 1000
		if v, ok := config["backoffMs"].(float64); ok {
			backoffMs = int(v)
		}

		var lastErr error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			if attempt > 0 {
				delay := time.Duration(backoffMs*(1<<(attempt-1))) * time.Millisecond
				select {
				case <-ctx.Done():
					return map[string]interface{}{
						"success":   false,
						"attempts":  attempt,
						"lastError": ctx.Err().Error(),
					}, nil
				case <-time.After(delay):
				}
			}

			_, err := inner(ctx, inputs, config)
			if err == nil {
				return map[string]interface{}{
					"success":   true,
					"attempts":  attempt + 1,
					"lastError": "",
				}, nil
			}
			lastErr = err
		}

		errMsg := ""
		if lastErr != nil {
			errMsg = lastErr.Error()
		}
		return map[string]interface{}{
			"success":   false,
			"attempts":  maxRetries + 1,
			"lastError": errMsg,
		}, nil
	}
}
