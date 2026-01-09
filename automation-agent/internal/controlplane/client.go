package controlplane

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client provides HTTPS client for control plane API
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// Config holds control plane client configuration
type Config struct {
	BaseURL string
	Token   string
	Timeout time.Duration
}

// NewClient creates a new control plane client
func NewClient(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	
	return &Client{
		baseURL: cfg.BaseURL,
		token:   cfg.Token,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// RegisterAgentRequest represents agent registration request
type RegisterAgentRequest struct {
	AgentID   string                 `json:"agent_id"`
	ProjectID string                 `json:"project_id"`
	OS        string                 `json:"os"`
	Labels    map[string]interface{} `json:"labels"`
}

// RegisterAgent registers the agent with the control plane
func (c *Client) RegisterAgent(ctx context.Context, req RegisterAgentRequest) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/agents/register", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("registration failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// LeaseJobResponse represents a job lease response
type LeaseJobResponse struct {
	JobID     string          `json:"job_id"`
	TenantID  string          `json:"tenant_id"`
	ProjectID string          `json:"project_id"`
	State     string          `json:"state"`
	Payload   json.RawMessage `json:"payload"`
}

// LeaseJob attempts to lease a job
func (c *Client) LeaseJob(ctx context.Context) (*LeaseJobResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/jobs/lease", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		// No job available
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("lease failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var job LeaseJobResponse
	if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &job, nil
}

// CompleteJobRequest represents a job completion request
type CompleteJobRequest struct {
	Success bool `json:"success"`
}

// CompleteJob marks a job as completed
func (c *Client) CompleteJob(ctx context.Context, jobID string, success bool) error {
	req := CompleteJobRequest{Success: success}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/jobs/%s/complete", c.baseURL, jobID)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("completion failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
