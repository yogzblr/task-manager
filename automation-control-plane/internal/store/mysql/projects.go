package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Project represents a project in the database
type Project struct {
	ProjectID string
	TenantID  string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateProject creates a new project
func (s *Store) CreateProject(ctx context.Context, qb *QueryBuilder, project *Project) error {
	if err := qb.ValidateTenantProject(project.ProjectID); err != nil {
		return err
	}

	query := `INSERT INTO projects (project_id, tenant_id, name, created_at, updated_at)
	          VALUES (?, ?, ?, NOW(), NOW())`
	
	_, err := s.db.ExecContext(ctx, query, project.ProjectID, project.TenantID, project.Name)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}
	
	return nil
}

// GetProject retrieves a project by ID
func (s *Store) GetProject(ctx context.Context, qb *QueryBuilder, projectID string) (*Project, error) {
	where, args := qb.BuildWhereClause("project_id = ?")
	args = append([]interface{}{projectID}, args...)
	
	query := fmt.Sprintf(`SELECT project_id, tenant_id, name, created_at, updated_at
	                     FROM projects WHERE %s`, where)
	
	var project Project
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&project.ProjectID, &project.TenantID, &project.Name,
		&project.CreatedAt, &project.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	
	return &project, nil
}

// ListProjects lists projects with pagination
func (s *Store) ListProjects(ctx context.Context, qb *QueryBuilder, limit int, cursor string) ([]*Project, string, error) {
	where, args := qb.BuildWhereClause("")
	if cursor != "" {
		where += " AND project_id > ?"
		args = append(args, cursor)
	}
	
	query := fmt.Sprintf(`SELECT project_id, tenant_id, name, created_at, updated_at
	                     FROM projects WHERE %s ORDER BY project_id LIMIT ?`, where)
	args = append(args, limit+1)
	
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()
	
	var projects []*Project
	var nextCursor string
	
	for rows.Next() {
		var project Project
		err := rows.Scan(
			&project.ProjectID, &project.TenantID, &project.Name,
			&project.CreatedAt, &project.UpdatedAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan project: %w", err)
		}
		
		if len(projects) < limit {
			projects = append(projects, &project)
		} else {
			nextCursor = project.ProjectID
			break
		}
	}
	
	return projects, nextCursor, nil
}
