package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/automation-platform/agent/internal/agent"
	"github.com/automation-platform/agent/internal/centrifugo"
	"github.com/automation-platform/agent/internal/controlplane"
	"github.com/yogzblr/probe"
)

func main() {
	// Load configuration from environment
	controlPlaneURL := getEnv("CONTROL_PLANE_URL", "http://localhost:8080")
	centrifugoURL := getEnv("CENTRIFUGO_URL", "ws://localhost:8000/connection/websocket")
	tenantID := getEnv("TENANT_ID", "")
	projectID := getEnv("PROJECT_ID", "")
	agentID := getEnv("AGENT_ID", generateAgentID())
	jwtToken := getEnv("JWT_TOKEN", "")
	
	if tenantID == "" || projectID == "" || jwtToken == "" {
		log.Fatal("TENANT_ID, PROJECT_ID, and JWT_TOKEN are required")
	}
	
	// Detect OS
	osName := "linux"
	if isWindows() {
		osName = "windows"
	}
	
	// Create agent
	ag := agent.NewAgent(agentID, tenantID, projectID, osName)
	
	// Initialize control plane client
	cpClient := controlplane.NewClient(controlplane.Config{
		BaseURL: controlPlaneURL,
		Token:   jwtToken,
		Timeout: 30 * time.Second,
	})
	
	// Initialize Centrifugo client
	centClient, err := centrifugo.NewClient(centrifugo.Config{
		URL:      centrifugoURL,
		APIKey:   jwtToken, // Simplified - in production, use separate API key
		TenantID: tenantID,
		AgentID:  agentID,
	})
	if err != nil {
		log.Fatalf("Failed to create Centrifugo client: %v", err)
	}
	
	// Initialize probe executor with all built-in tasks
	probeExecutor := probe.New()
	
	// Start agent
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	if err := ag.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}
	
	// Register with control plane
	if err := cpClient.RegisterAgent(ctx, controlplane.RegisterAgentRequest{
		AgentID:   agentID,
		ProjectID: projectID,
		OS:        osName,
		Labels:    make(map[string]interface{}),
	}); err != nil {
		log.Fatalf("Failed to register agent: %v", err)
	}
	
	// Connect to Centrifugo
	if err := centClient.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to Centrifugo: %v", err)
	}
	defer centClient.Disconnect()
	
	// Start message handler
	handler := &MessageHandler{
		agent:         ag,
		cpClient:      cpClient,
		probeExecutor: probeExecutor,
	}
	
	if err := centClient.StartMessageLoop(ctx, handler); err != nil {
		log.Fatalf("Failed to start message loop: %v", err)
	}
	
	// Start heartbeat loop
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				centClient.PublishHeartbeat(ctx, string(ag.State()), 0)
			}
		}
	}()
	
	// Wait for signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	<-sigint
	
	log.Println("Shutting down agent...")
	ag.Stop()
}

// MessageHandler handles Centrifugo messages
type MessageHandler struct {
	agent         *agent.Agent
	cpClient      *controlplane.Client
	probeExecutor *probe.Probe
}

func (h *MessageHandler) HandleJobAvailable(jobID string) {
	ctx := context.Background()
	
	// Transition to leasing state
	if err := h.agent.StateMachine.Transition(agent.StateLeasing); err != nil {
		log.Printf("Failed to transition to leasing: %v", err)
		return
	}
	
	// Try to lease the job
	job, err := h.cpClient.LeaseJob(ctx)
	if err != nil || job == nil {
		h.agent.StateMachine.Transition(agent.StateIdle)
		return
	}
	
	// Transition to executing
	if err := h.agent.StateMachine.Transition(agent.StateExecuting); err != nil {
		log.Printf("Failed to transition to executing: %v", err)
	}
	
	// Execute workflow using probe (job.Payload is now expected to be YAML)
	results, err := h.probeExecutor.ExecuteYAML(ctx, job.Payload)
	
	// Complete job
	success := err == nil && results.Success
	output := formatProbeResults(results, err)
	log.Printf("Job %s completed with output: %s", job.JobID, output)
	h.cpClient.CompleteJob(ctx, job.JobID, success)
	
	// Transition back to idle
	if err := h.agent.StateMachine.Transition(agent.StateIdle); err != nil {
		log.Printf("Failed to transition to idle: %v", err)
	}
}

// formatProbeResults formats probe execution results for API
func formatProbeResults(results *probe.WorkflowResult, err error) string {
	if err != nil {
		return fmt.Sprintf("Execution failed: %v", err)
	}
	
	// Marshal results to JSON for API
	data, marshalErr := json.Marshal(results)
	if marshalErr != nil {
		return fmt.Sprintf("Execution completed but failed to marshal results: %v", marshalErr)
	}
	
	return string(data)
}

func (h *MessageHandler) HandleCancelJob(jobID string) {
	// TODO: Implement job cancellation
	log.Printf("Job cancellation requested: %s", jobID)
}

func (h *MessageHandler) HandleUpgradeAvailable(version, url, sha256, signature, keyID string) {
	// TODO: Implement upgrade mechanism
	log.Printf("Upgrade available: version %s", version)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func generateAgentID() string {
	hostname, _ := os.Hostname()
	return fmt.Sprintf("%s-%d", hostname, os.Getpid())
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}
