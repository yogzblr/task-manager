package centrifugo

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"

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
	// Create client config
	opts := centrifuge.Config{
		Token: cfg.APIKey,
	}

	// Note: Proxy configuration would need to be set via environment variables
	// or through the underlying HTTP client used by centrifuge-go
	// For now, we'll create the client without explicit proxy config

	client := centrifuge.NewJsonClient(cfg.URL, opts)

	c := &Client{
		url:      cfg.URL,
		apiKey:   cfg.APIKey,
		tenantID: cfg.TenantID,
		agentID:  cfg.AgentID,
		client:   client,
		proxyURL: cfg.ProxyURL,
	}

	// Set up connection lifecycle event handlers
	c.setupConnectionHandlers()

	return c, nil
}

// setupConnectionHandlers configures all connection lifecycle event handlers
func (c *Client) setupConnectionHandlers() {
	// Handle connecting event
	c.client.OnConnecting(func(e centrifuge.ConnectingEvent) {
		log.Printf("[Centrifugo] Connecting to %s (code: %d, reason: %s)", c.url, e.Code, e.Reason)
	})

	// Handle successful connection
	c.client.OnConnected(func(e centrifuge.ConnectedEvent) {
		log.Printf("[Centrifugo] Connected successfully (client_id: %s, version: %s)", e.ClientID, e.Version)
	})

	// Handle disconnection
	c.client.OnDisconnected(func(e centrifuge.DisconnectedEvent) {
		log.Printf("[Centrifugo] Disconnected (code: %d, reason: %s)", e.Code, e.Reason)
	})

	// Handle errors
	c.client.OnError(func(e centrifuge.ErrorEvent) {
		log.Printf("[Centrifugo] Error: %v", e.Error)
	})

	// Handle server-side subscription events
	c.client.OnSubscribed(func(e centrifuge.ServerSubscribedEvent) {
		log.Printf("[Centrifugo] Subscribed to server-side channel: %s (recoverable: %v, recovered: %v)",
			e.Channel, e.Recoverable, e.Recovered)
	})

	// Handle server-side unsubscription events
	c.client.OnUnsubscribed(func(e centrifuge.ServerUnsubscribedEvent) {
		log.Printf("[Centrifugo] Unsubscribed from server-side channel: %s (code: %d, reason: %s)",
			e.Channel, e.Code, e.Reason)
	})
}

// Connect connects to Centrifugo
func (c *Client) Connect(ctx context.Context) error {
	log.Printf("[Centrifugo] Initiating connection to %s", c.url)
	if err := c.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	log.Printf("[Centrifugo] Connection initiated successfully")
	return nil
}

// Disconnect disconnects from Centrifugo
func (c *Client) Disconnect() {
	log.Printf("[Centrifugo] Closing connection")
	c.client.Close()
}

// Subscribe subscribes to the agent's channel
// This method supports server-side subscriptions (via channels claim in JWT)
// Event handlers must be set up BEFORE calling Connect()
func (c *Client) Subscribe(handler func([]byte)) error {
	// Listen for server-side subscription publications
	// When channels claim is present in JWT, Centrifugo sets up server-side subscriptions
	log.Printf("[Centrifugo] Setting up publication handler for server-side subscriptions")
	c.client.OnPublication(func(e centrifuge.ServerPublicationEvent) {
		log.Printf("[Centrifugo] Received publication on channel %s (offset: %d)", e.Channel, e.Publication.Offset)
		// Call the handler in a goroutine to avoid blocking the read loop
		go handler(e.Publication.Data)
	})

	return nil
}

// PublishHeartbeat publishes a heartbeat message
func (c *Client) PublishHeartbeat(ctx context.Context, state string, activeJobs int) error {
	message := map[string]interface{}{
		"type":        "heartbeat",
		"state":       state,
		"active_jobs": activeJobs,
	}
	
	_, err := json.Marshal(message)
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
