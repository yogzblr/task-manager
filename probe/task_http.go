package probe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPTask performs HTTP health checks
type HTTPTask struct {
	URL            string
	Method         string
	ExpectedStatus []int
	Timeout        time.Duration
	Headers        map[string]string
}

// Configure sets up the HTTP task
func (t *HTTPTask) Configure(config map[string]interface{}) error {
	// URL is required
	url, ok := config["url"].(string)
	if !ok || url == "" {
		return fmt.Errorf("url is required")
	}
	t.URL = url
	
	// Method defaults to GET
	if method, ok := config["method"].(string); ok {
		t.Method = method
	} else {
		t.Method = "GET"
	}
	
	// Expected status codes (default: 200)
	if expectedStatus, ok := config["expected_status"].([]interface{}); ok {
		t.ExpectedStatus = make([]int, len(expectedStatus))
		for i, s := range expectedStatus {
			if statusCode, ok := s.(int); ok {
				t.ExpectedStatus[i] = statusCode
			}
		}
	} else {
		t.ExpectedStatus = []int{200}
	}
	
	// Timeout (default: 30s)
	if timeoutStr, ok := config["timeout"].(string); ok {
		duration, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout: %w", err)
		}
		t.Timeout = duration
	} else {
		t.Timeout = 30 * time.Second
	}
	
	// Headers
	if headers, ok := config["headers"].(map[string]interface{}); ok {
		t.Headers = make(map[string]string)
		for k, v := range headers {
			if strVal, ok := v.(string); ok {
				t.Headers[k] = strVal
			}
		}
	}
	
	return nil
}

// Execute performs the HTTP request
func (t *HTTPTask) Execute(ctx context.Context) (interface{}, error) {
	client := &http.Client{
		Timeout: t.Timeout,
	}
	
	req, err := http.NewRequestWithContext(ctx, t.Method, t.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add headers
	for k, v := range t.Headers {
		req.Header.Set(k, v)
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Read body (limit to 1MB)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	// Check if status code is expected
	statusOK := false
	for _, expected := range t.ExpectedStatus {
		if resp.StatusCode == expected {
			statusOK = true
			break
		}
	}
	
	result := map[string]interface{}{
		"status_code": resp.StatusCode,
		"body":        string(body),
		"headers":     resp.Header,
	}
	
	if !statusOK {
		return result, fmt.Errorf("unexpected status code: %d (expected one of %v)", resp.StatusCode, t.ExpectedStatus)
	}
	
	return result, nil
}
