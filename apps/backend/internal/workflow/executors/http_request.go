package executors

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func HTTPRequestExecutor() func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return func(ctx context.Context, inputs map[string]interface{}, config map[string]interface{}) (map[string]interface{}, error) {
		method, _ := config["method"].(string)
		if method == "" {
			method = "GET"
		}

		url, _ := config["url"].(string)
		if url == "" {
			return nil, fmt.Errorf("url is required")
		}

		headers, _ := config["headers"].(map[string]interface{})
		body, _ := config["body"].(string)

		var reqBody io.Reader
		if body != "" {
			reqBody = strings.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		for k, v := range headers {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("http request failed: %w", err)
		}
		defer resp.Body.Close()

		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		respHeaders := make(map[string]string)
		for k, v := range resp.Header {
			respHeaders[k] = strings.Join(v, ",")
		}

		return map[string]interface{}{
			"statusCode": resp.StatusCode,
			"body":       string(respBytes),
			"headers":    respHeaders,
			"status":     "completed",
		}, nil
	}
}
