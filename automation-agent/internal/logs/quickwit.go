package logs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client provides Quickwit log streaming
type Client struct {
	url        string
	httpClient *http.Client
	index      string
}

// Config holds Quickwit configuration
type Config struct {
	URL   string
	Index string
}

// NewClient creates a new Quickwit client
func NewClient(cfg Config) *Client {
	return &Client{
		url:   cfg.URL,
		index: cfg.Index,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp string            `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	JobID     string            `json:"job_id,omitempty"`
	AgentID   string            `json:"agent_id,omitempty"`
	TenantID  string            `json:"tenant_id,omitempty"`
	ProjectID string            `json:"project_id,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
}

// StreamLog streams a log entry to Quickwit
func (c *Client) StreamLog(ctx context.Context, entry LogEntry) error {
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}
	
	url := fmt.Sprintf("%s/api/v1/%s/ingest", c.url, c.index)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	return nil
}

// StreamBatch streams multiple log entries in a batch
func (c *Client) StreamBatch(ctx context.Context, entries []LogEntry) error {
	if len(entries) == 0 {
		return nil
	}
	
	// Quickwit typically accepts NDJSON (newline-delimited JSON)
	var buf bytes.Buffer
	for _, entry := range entries {
		if entry.Timestamp == "" {
			entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
		}
		
		jsonData, err := json.Marshal(entry)
		if err != nil {
			continue
		}
		buf.Write(jsonData)
		buf.WriteString("\n")
	}
	
	url := fmt.Sprintf("%s/api/v1/%s/ingest", c.url, c.index)
	req, err := http.NewRequestWithContext(ctx, "POST", url, &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/x-ndjson")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	return nil
}
