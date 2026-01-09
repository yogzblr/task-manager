package api

import (
	"encoding/json"
	"net/http"

	"github.com/automation-platform/control-plane/internal/auth"
	"github.com/automation-platform/control-plane/internal/store/mysql"
)

// ProjectsHandler handles project-related API requests
type ProjectsHandler struct {
	store     *mysql.Store
	authorizer *auth.RBACAuthorizer
}

// NewProjectsHandler creates a new projects handler
func NewProjectsHandler(store *mysql.Store, authorizer *auth.RBACAuthorizer) *ProjectsHandler {
	return &ProjectsHandler{
		store:     store,
		authorizer: authorizer,
	}
}

// ListProjects handles GET /projects
func (h *ProjectsHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "missing claims", http.StatusInternalServerError)
		return
	}

	authorizedProjects, err := h.authorizer.GetAuthorizedProjects(r.Context(), claims)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	qb := mysql.NewQueryBuilder(claims.TenantID, authorizedProjects)

	limit := 50
	cursor := r.URL.Query().Get("cursor")

	projects, nextCursor, err := h.store.ListProjects(r.Context(), qb, limit, cursor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"projects":    projects,
		"next_cursor": nextCursor,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
