package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/automation-platform/control-plane/internal/auth"
	"github.com/automation-platform/control-plane/internal/store/mysql"
)

// AgentsHandler handles agent-related API requests
type AgentsHandler struct {
	store     *mysql.Store
	authorizer *auth.RBACAuthorizer
}

// NewAgentsHandler creates a new agents handler
func NewAgentsHandler(store *mysql.Store, authorizer *auth.RBACAuthorizer) *AgentsHandler {
	return &AgentsHandler{
		store:     store,
		authorizer: authorizer,
	}
}

// RegisterAgentRequest represents an agent registration request
type RegisterAgentRequest struct {
	AgentID   string          `json:"agent_id"`
	ProjectID string          `json:"project_id"`
	OS        string          `json:"os"`
	Labels    json.RawMessage `json:"labels"`
}

// RegisterAgent handles POST /agents/register
func (h *AgentsHandler) RegisterAgent(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "missing claims", http.StatusInternalServerError)
		return
	}

	if claims.AgentID == "" {
		http.Error(w, "only agents can register", http.StatusForbidden)
		return
	}

	var req RegisterAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	authorizedProjects, err := h.authorizer.GetAuthorizedProjects(r.Context(), claims)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	qb := mysql.NewQueryBuilder(claims.TenantID, authorizedProjects)

	agent := &mysql.Agent{
		AgentID:   claims.AgentID,
		TenantID:  claims.TenantID,
		ProjectID: req.ProjectID,
		OS:        &req.OS,
		Labels:    req.Labels,
	}

	if err := h.store.CreateOrUpdateAgent(r.Context(), qb, agent); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpgradeAgent handles POST /agents/{id}/upgrade
func (h *AgentsHandler) UpgradeAgent(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "missing claims", http.StatusInternalServerError)
		return
	}

	// Extract agent ID from path (Go 1.21 compatible)
	// Path format: /api/agents/{id}/upgrade
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var agentID string
	for i, part := range parts {
		if part == "agents" && i+1 < len(parts) {
			agentID = parts[i+1]
			break
		}
	}
	if agentID == "" {
		http.Error(w, "agent_id required", http.StatusBadRequest)
		return
	}

	// Authorize
	// For now, we'll need to get the agent's project_id to authorize
	// This is a simplified version - in production, you'd fetch the agent first
	if err := h.authorizer.Authorize(r.Context(), claims, "", auth.PermissionAgentUpgrade); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// TODO: Trigger upgrade via Centrifugo
	// This would publish an upgrade_available message to the agent's channel

	w.WriteHeader(http.StatusAccepted)
}
