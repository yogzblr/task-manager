package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/automation-platform/control-plane/internal/auth"
	"github.com/automation-platform/control-plane/internal/centrifugo"
	"github.com/automation-platform/control-plane/internal/store/mysql"
	"github.com/google/uuid"
)

// JobsHandler handles job-related API requests
type JobsHandler struct {
	store            *mysql.Store
	authorizer       *auth.RBACAuthorizer
	centrifugoClient *centrifugo.Client
}

// NewJobsHandler creates a new jobs handler
func NewJobsHandler(store *mysql.Store, authorizer *auth.RBACAuthorizer, centrifugoClient *centrifugo.Client) *JobsHandler {
	return &JobsHandler{
		store:            store,
		authorizer:       authorizer,
		centrifugoClient: centrifugoClient,
	}
}

// CreateJobRequest represents a job creation request
type CreateJobRequest struct {
	TenantID  string          `json:"tenant_id"`
	ProjectID string          `json:"project_id"`
	AgentID   string          `json:"agent_id"`
	Workflow  json.RawMessage `json:"workflow"`
}

// CreateJobResponse represents a job creation response
type CreateJobResponse struct {
	JobID     string `json:"job_id"`
	TenantID  string `json:"tenant_id"`
	ProjectID string `json:"project_id"`
	State     string `json:"state"`
}

// CreateJob handles POST /jobs
func (h *JobsHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "missing claims", http.StatusInternalServerError)
		return
	}

	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Authorize
	if err := h.authorizer.Authorize(r.Context(), claims, req.ProjectID, auth.PermissionJobRun); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Get authorized projects for query builder
	authorizedProjects, err := h.authorizer.GetAuthorizedProjects(r.Context(), claims)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	qb := mysql.NewQueryBuilder(claims.TenantID, authorizedProjects)

	// Use tenant_id from request if provided, otherwise use from claims
	tenantID := req.TenantID
	if tenantID == "" {
		tenantID = claims.TenantID
	}

	// Create job
	jobID := uuid.New().String()
	job := &mysql.Job{
		JobID:     jobID,
		TenantID:  tenantID,
		ProjectID: req.ProjectID,
		State:     "pending",
		Payload:   req.Workflow,
	}

	if err := h.store.CreateJob(r.Context(), qb, job); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Publish job notification to Centrifugo if agent_id is specified
	if req.AgentID != "" {
		channel := "agents." + tenantID + "." + req.AgentID
		message := centrifugo.JobAvailableMessage{
			Type:  "job_available",
			JobID: jobID,
		}

		log.Printf("[ControlPlane] Publishing job_available to channel %s for job %s", channel, jobID)
		if err := h.centrifugoClient.Publish(r.Context(), channel, message); err != nil {
			log.Printf("[ControlPlane] Failed to publish to Centrifugo: %v", err)
			// Don't fail the request if Centrifugo publish fails
			// The agent can still poll for jobs
		} else {
			log.Printf("[ControlPlane] Successfully published job notification to %s", channel)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateJobResponse{
		JobID:     jobID,
		TenantID:  tenantID,
		ProjectID: req.ProjectID,
		State:     "pending",
	})
}

// LeaseJobRequest represents a job lease request
type LeaseJobRequest struct {
	AgentID string `json:"agent_id"`
}

// LeaseJob handles POST /jobs/{id}/lease
func (h *JobsHandler) LeaseJob(w http.ResponseWriter, r *http.Request) {
	log.Printf("[LeaseJob] Handler called for path: %s", r.URL.Path)

	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		log.Printf("[LeaseJob] Missing claims")
		http.Error(w, "missing claims", http.StatusInternalServerError)
		return
	}

	log.Printf("[LeaseJob] Claims: AgentID=%s, TenantID=%s, ProjectID=%s", claims.AgentID, claims.TenantID, claims.ProjectID)

	// Extract job ID from path (Go 1.21 compatible)
	// Path format: /api/jobs/{id}/lease
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var jobID string
	for i, part := range parts {
		if part == "jobs" && i+1 < len(parts) {
			jobID = parts[i+1]
			break
		}
	}
	if jobID == "" {
		log.Printf("[LeaseJob] No job_id in path")
		http.Error(w, "job_id required", http.StatusBadRequest)
		return
	}

	log.Printf("[LeaseJob] Extracted job_id: %s", jobID)

	// Agents can only lease jobs
	if claims.AgentID == "" {
		log.Printf("[LeaseJob] Not an agent")
		http.Error(w, "only agents can lease jobs", http.StatusForbidden)
		return
	}

	authorizedProjects, err := h.authorizer.GetAuthorizedProjects(r.Context(), claims)
	if err != nil {
		log.Printf("[LeaseJob] Failed to get authorized projects: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("[LeaseJob] Agent %s authorized for projects: %v (tenant: %s)", claims.AgentID, authorizedProjects, claims.TenantID)
	qb := mysql.NewQueryBuilder(claims.TenantID, authorizedProjects)

	// Lease job (30 minute lease)
	leaseDuration := 30 * time.Minute
	log.Printf("[LeaseJob] Attempting to lease job %s for agent %s", jobID, claims.AgentID)
	if err := h.store.LeaseJob(r.Context(), qb, jobID, claims.AgentID, leaseDuration); err != nil {
		log.Printf("[LeaseJob] Failed to lease job: %v", err)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	log.Printf("[LeaseJob] Successfully leased job %s", jobID)

	// Get the leased job
	job, err := h.store.GetJob(r.Context(), qb, jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// CompleteJobRequest represents a job completion request
type CompleteJobRequest struct {
	Success bool `json:"success"`
}

// CompleteJob handles POST /jobs/{id}/complete
func (h *JobsHandler) CompleteJob(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "missing claims", http.StatusInternalServerError)
		return
	}

	// Extract job ID from path (Go 1.21 compatible)
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var jobID string
	for i, part := range parts {
		if part == "jobs" && i+1 < len(parts) {
			jobID = parts[i+1]
			break
		}
	}
	if jobID == "" {
		http.Error(w, "job_id required", http.StatusBadRequest)
		return
	}

	if claims.AgentID == "" {
		http.Error(w, "only agents can complete jobs", http.StatusForbidden)
		return
	}

	var req CompleteJobRequest
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

	if err := h.store.CompleteJob(r.Context(), qb, jobID, claims.AgentID, req.Success); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListJobs handles GET /jobs
func (h *JobsHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "missing claims", http.StatusInternalServerError)
		return
	}

	// Authorize
	if err := h.authorizer.Authorize(r.Context(), claims, "", auth.PermissionJobRead); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
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

	jobs, nextCursor, err := h.store.ListJobs(r.Context(), qb, limit, cursor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"jobs":        jobs,
		"next_cursor": nextCursor,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
