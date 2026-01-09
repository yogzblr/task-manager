package api

import (
	"encoding/json"
	"net/http"

	"github.com/automation-platform/control-plane/internal/auth"
	"github.com/automation-platform/control-plane/internal/store/mysql"
)

// AuditHandler handles audit log API requests
type AuditHandler struct {
	store     *mysql.Store
	authorizer *auth.RBACAuthorizer
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(store *mysql.Store, authorizer *auth.RBACAuthorizer) *AuditHandler {
	return &AuditHandler{
		store:     store,
		authorizer: authorizer,
	}
}

// ListAuditLogs handles GET /audit/logs
func (h *AuditHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "missing claims", http.StatusInternalServerError)
		return
	}

	// Authorize
	if err := h.authorizer.Authorize(r.Context(), claims, "", auth.PermissionAuditRead); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	authorizedProjects, err := h.authorizer.GetAuthorizedProjects(r.Context(), claims)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	qb := mysql.NewQueryBuilder(claims.TenantID, authorizedProjects)

	filters := mysql.AuditLogFilters{
		ProjectID: stringPtr(r.URL.Query().Get("project_id")),
		ActorID:   stringPtr(r.URL.Query().Get("actor_id")),
		Action:    stringPtr(r.URL.Query().Get("action")),
	}

	limit := 50
	cursor := r.URL.Query().Get("cursor")

	logs, nextCursor, err := h.store.ListAuditLogs(r.Context(), qb, filters, limit, cursor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"logs":       logs,
		"next_cursor": nextCursor,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
