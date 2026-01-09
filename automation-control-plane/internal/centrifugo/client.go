package centrifugo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client provides Centrifugo API access
type Client struct {
	url     string
	apiKey  string
	httpClient *http.Client
}

// Config holds Centrifugo configuration
type Config struct {
	URL    string
	APIKey string
}

// NewClient creates a new Centrifugo client
func NewClient(cfg Config) *Client {
	return &Client{
		url:    cfg.URL,
		apiKey: cfg.APIKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// PublishRequest represents a publish request
type PublishRequest struct {
	Channel string      `json:"channel"`
	Data    interface{} `json:"data"`
}

// PublishResponse represents a publish response
type PublishResponse struct {
	Error  string `json:"error,omitempty"`
	Result struct {
		Offset uint64 `json:"offset"`
		Epoch  string `json:"epoch"`
	} `json:"result,omitempty"`
}

// Publish publishes a message to a channel
func (c *Client) Publish(ctx context.Context, channel string, data interface{}) error {
	reqBody := PublishRequest{
		Channel: channel,
		Data:    data,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.url+"/api/publish", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "apikey "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var publishResp PublishResponse
	if err := json.NewDecoder(resp.Body).Decode(&publishResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if publishResp.Error != "" {
		return fmt.Errorf("centrifugo error: %s", publishResp.Error)
	}

	return nil
}

// JobAvailableMessage represents a job_available message
type JobAvailableMessage struct {
	Type  string `json:"type"`
	JobID string `json:"job_id"`
}

// CancelJobMessage represents a cancel_job message
type CancelJobMessage struct {
	Type  string `json:"type"`
	JobID string `json:"job_id"`
}

// UpgradeAvailableMessage represents an upgrade_available message
type UpgradeAvailableMessage struct {
	Type     string `json:"type"`
	Version  string `json:"version"`
	URL      string `json:"url"`
	SHA256   string `json:"sha256"`
	Signature string `json:"signature"`
	KeyID    string `json:"key_id"`
}
