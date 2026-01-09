package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Agent represents an agent in the database
type Agent struct {
	AgentID   string
	TenantID  string
	ProjectID string
	OS        *string
	Labels    json.RawMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateOrUpdateAgent creates or updates an agent
func (s *Store) CreateOrUpdateAgent(ctx context.Context, qb *QueryBuilder, agent *Agent) error {
	if err := qb.ValidateTenantProject(agent.ProjectID); err != nil {
		return err
	}

	query := `INSERT INTO agents (agent_id, tenant_id, project_id, os, labels, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	          ON DUPLICATE KEY UPDATE
	          os = VALUES(os),
	          labels = VALUES(labels),
	          updated_at = NOW()`
	
	_, err := s.db.ExecContext(ctx, query, agent.AgentID, agent.TenantID, agent.ProjectID, agent.OS, agent.Labels)
	if err != nil {
		return fmt.Errorf("failed to create/update agent: %w", err)
	}
	
	return nil
}

// GetAgent retrieves an agent by ID
func (s *Store) GetAgent(ctx context.Context, qb *QueryBuilder, agentID string) (*Agent, error) {
	where, args := qb.BuildWhereClause("agent_id = ?")
	args = append([]interface{}{agentID}, args...)
	
	query := fmt.Sprintf(`SELECT agent_id, tenant_id, project_id, os, labels, created_at, updated_at
	                     FROM agents WHERE %s`, where)
	
	var agent Agent
	var os sql.NullString
	
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&agent.AgentID, &agent.TenantID, &agent.ProjectID,
		&os, &agent.Labels, &agent.CreatedAt, &agent.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("agent not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	
	if os.Valid {
		agent.OS = &os.String
	}
	
	return &agent, nil
}

// ListAgents lists agents for a project
func (s *Store) ListAgents(ctx context.Context, qb *QueryBuilder, projectID string, limit int, cursor string) ([]*Agent, string, error) {
	where, args := qb.BuildWhereClause("project_id = ?")
	args = append([]interface{}{projectID}, args...)
	
	if cursor != "" {
		where += " AND agent_id > ?"
		args = append(args, cursor)
	}
	
	query := fmt.Sprintf(`SELECT agent_id, tenant_id, project_id, os, labels, created_at, updated_at
	                     FROM agents WHERE %s ORDER BY agent_id LIMIT ?`, where)
	args = append(args, limit+1)
	
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()
	
	var agents []*Agent
	var nextCursor string
	
	for rows.Next() {
		var agent Agent
		var os sql.NullString
		
		err := rows.Scan(
			&agent.AgentID, &agent.TenantID, &agent.ProjectID,
			&os, &agent.Labels, &agent.CreatedAt, &agent.UpdatedAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan agent: %w", err)
		}
		
		if os.Valid {
			agent.OS = &os.String
		}
		
		if len(agents) < limit {
			agents = append(agents, &agent)
		} else {
			nextCursor = agent.AgentID
			break
		}
	}
	
	return agents, nextCursor, nil
}
