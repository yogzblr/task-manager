package centrifugo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/centrifugal/centrifuge-go"
)

// Client provides proxy-aware Centrifugo WebSocket client
type Client struct {
	url      string
	apiKey   string
	tenantID string
	agentID  string
	client   *centrifuge.Client
	proxyURL *url.URL
}

// Config holds Centrifugo client configuration
type Config struct {
	URL      string
	APIKey   string
	TenantID string
	AgentID  string
	ProxyURL *url.URL // Optional proxy URL
}

// NewClient creates a new Centrifugo client
func NewClient(cfg Config) (*Client, error) {
	opts := centrifuge.DefaultConfig()
	opts.Token = cfg.APIKey
	
	// Configure proxy if provided
	if cfg.ProxyURL != nil {
		opts.HTTPClient = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(cfg.ProxyURL),
			},
		}
	}
	
	client := centrifuge.NewJsonClient(cfg.URL, opts)
	
	return &Client{
		url:      cfg.URL,
		apiKey:   cfg.APIKey,
		tenantID: cfg.TenantID,
		agentID:  cfg.AgentID,
		client:   client,
		proxyURL: cfg.ProxyURL,
	}, nil
}

// Connect connects to Centrifugo
func (c *Client) Connect(ctx context.Context) error {
	if err := c.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	return nil
}

// Disconnect disconnects from Centrifugo
func (c *Client) Disconnect() {
	c.client.Close()
}

// Subscribe subscribes to the agent's channel
func (c *Client) Subscribe(handler func([]byte)) error {
	channel := fmt.Sprintf("agents.%s.%s", c.tenantID, c.agentID)
	
	sub, err := c.client.NewSubscription(channel)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	
	sub.OnPublication(func(e centrifuge.PublicationEvent) {
		handler(e.Data)
	})
	
	if err := sub.Subscribe(); err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	
	return nil
}

// PublishHeartbeat publishes a heartbeat message
func (c *Client) PublishHeartbeat(ctx context.Context, state string, activeJobs int) error {
	message := map[string]interface{}{
		"type":        "heartbeat",
		"state":       state,
		"active_jobs": activeJobs,
	}
	
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat: %w", err)
	}
	
	// Note: In production, heartbeats might be sent via HTTP API instead
	// This is a simplified version
	return nil
}

// MessageHandler handles incoming messages
type MessageHandler interface {
	HandleJobAvailable(jobID string)
	HandleCancelJob(jobID string)
	HandleUpgradeAvailable(version, url, sha256, signature, keyID string)
}

// StartMessageLoop starts the message handling loop
func (c *Client) StartMessageLoop(ctx context.Context, handler MessageHandler) error {
	return c.Subscribe(func(data []byte) {
		var msg map[string]interface{}
		if err := json.Unmarshal(data, &msg); err != nil {
			return
		}
		
		msgType, ok := msg["type"].(string)
		if !ok {
			return
		}
		
		switch msgType {
		case "job_available":
			if jobID, ok := msg["job_id"].(string); ok {
				handler.HandleJobAvailable(jobID)
			}
		case "cancel_job":
			if jobID, ok := msg["job_id"].(string); ok {
				handler.HandleCancelJob(jobID)
			}
		case "upgrade_available":
			version, _ := msg["version"].(string)
			url, _ := msg["url"].(string)
			sha256, _ := msg["sha256"].(string)
			signature, _ := msg["signature"].(string)
			keyID, _ := msg["key_id"].(string)
			handler.HandleUpgradeAvailable(version, url, sha256, signature, keyID)
		}
	})
}
