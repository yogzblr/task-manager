package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Job represents a job in the database
type Job struct {
	JobID         string
	TenantID      string
	ProjectID     string
	State         string
	LeaseOwner    *string
	LeaseExpiresAt *time.Time
	Payload       json.RawMessage
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CompletedAt   *time.Time
}

// CreateJob creates a new job
func (s *Store) CreateJob(ctx context.Context, qb *QueryBuilder, job *Job) error {
	if err := qb.ValidateTenantProject(job.ProjectID); err != nil {
		return err
	}

	query := `INSERT INTO jobs (job_id, tenant_id, project_id, state, payload, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, NOW(), NOW())`
	
	_, err := s.db.ExecContext(ctx, query, job.JobID, job.TenantID, job.ProjectID, job.State, job.Payload)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}
	
	return nil
}

// GetJob retrieves a job by ID with security filtering
func (s *Store) GetJob(ctx context.Context, qb *QueryBuilder, jobID string) (*Job, error) {
	where, args := qb.BuildWhereClause("job_id = ?")
	args = append([]interface{}{jobID}, args...)
	
	query := fmt.Sprintf(`SELECT job_id, tenant_id, project_id, state, lease_owner, 
	                     lease_expires_at, payload, created_at, updated_at, completed_at
	                     FROM jobs WHERE %s`, where)
	
	var job Job
	var leaseOwner, completedAt sql.NullString
	var leaseExpiresAt sql.NullTime
	
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&job.JobID, &job.TenantID, &job.ProjectID, &job.State,
		&leaseOwner, &leaseExpiresAt, &job.Payload,
		&job.CreatedAt, &job.UpdatedAt, &completedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	
	if leaseOwner.Valid {
		job.LeaseOwner = &leaseOwner.String
	}
	if leaseExpiresAt.Valid {
		job.LeaseExpiresAt = &leaseExpiresAt.Time
	}
	if completedAt.Valid {
		t, _ := time.Parse(time.RFC3339, completedAt.String)
		job.CompletedAt = &t
	}
	
	return &job, nil
}

// LeaseJob atomically leases a job (optimistic locking)
func (s *Store) LeaseJob(ctx context.Context, qb *QueryBuilder, jobID, agentID string, leaseDuration time.Duration) error {
	if err := qb.ValidateTenantProject(""); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check current state and lease
	where, args := qb.BuildWhereClause("job_id = ? AND state = 'pending'")
	args = append([]interface{}{jobID}, args...)
	
	var currentLeaseOwner sql.NullString
	var currentLeaseExpires sql.NullTime
	
	checkQuery := fmt.Sprintf(`SELECT lease_owner, lease_expires_at FROM jobs WHERE %s`, where)
	err = tx.QueryRowContext(ctx, checkQuery, args...).Scan(&currentLeaseOwner, &currentLeaseExpires)
	if err == sql.ErrNoRows {
		return fmt.Errorf("job not available for leasing")
	}
	if err != nil {
		return fmt.Errorf("failed to check job lease: %w", err)
	}

	// Check if lease is still valid
	if currentLeaseOwner.Valid && currentLeaseExpires.Valid {
		if time.Now().Before(currentLeaseExpires.Time) {
			return fmt.Errorf("job already leased")
		}
	}

	// Acquire lease
	expiresAt := time.Now().Add(leaseDuration)
	updateWhere, updateArgs := qb.BuildWhereClause("job_id = ? AND state = 'pending'")
	updateArgs = append([]interface{}{agentID, expiresAt, jobID}, updateArgs...)
	
	updateQuery := fmt.Sprintf(`UPDATE jobs SET state = 'leased', lease_owner = ?, 
	                           lease_expires_at = ?, updated_at = NOW() WHERE %s`, updateWhere)
	
	result, err := tx.ExecContext(ctx, updateQuery, updateArgs...)
	if err != nil {
		return fmt.Errorf("failed to lease job: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("job lease failed - job may have been leased by another agent")
	}
	
	return tx.Commit()
}

// CompleteJob marks a job as completed
func (s *Store) CompleteJob(ctx context.Context, qb *QueryBuilder, jobID, agentID string, success bool) error {
	if err := qb.ValidateTenantProject(""); err != nil {
		return err
	}

	state := "completed"
	if !success {
		state = "failed"
	}

	where, args := qb.BuildWhereClause("job_id = ? AND lease_owner = ?")
	args = append([]interface{}{jobID, agentID}, args...)
	
	query := fmt.Sprintf(`UPDATE jobs SET state = ?, completed_at = NOW(), updated_at = NOW() 
	                     WHERE %s`, where)
	
	result, err := s.db.ExecContext(ctx, query, append([]interface{}{state}, args...)...)
	if err != nil {
		return fmt.Errorf("failed to complete job: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("job not found or not leased by this agent")
	}
	
	return nil
}

// ListJobs lists jobs with pagination and filtering
func (s *Store) ListJobs(ctx context.Context, qb *QueryBuilder, limit int, cursor string) ([]*Job, string, error) {
	where, args := qb.BuildWhereClause("")
	if cursor != "" {
		where += " AND job_id > ?"
		args = append(args, cursor)
	}
	
	query := fmt.Sprintf(`SELECT job_id, tenant_id, project_id, state, lease_owner, 
	                     lease_expires_at, payload, created_at, updated_at, completed_at
	                     FROM jobs WHERE %s ORDER BY job_id LIMIT ?`, where)
	args = append(args, limit+1) // Fetch one extra to determine if there's a next page
	
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list jobs: %w", err)
	}
	defer rows.Close()
	
	var jobs []*Job
	var nextCursor string
	
	for rows.Next() {
		var job Job
		var leaseOwner, completedAt sql.NullString
		var leaseExpiresAt sql.NullTime
		
		err := rows.Scan(
			&job.JobID, &job.TenantID, &job.ProjectID, &job.State,
			&leaseOwner, &leaseExpiresAt, &job.Payload,
			&job.CreatedAt, &job.UpdatedAt, &completedAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan job: %w", err)
		}
		
		if leaseOwner.Valid {
			job.LeaseOwner = &leaseOwner.String
		}
		if leaseExpiresAt.Valid {
			job.LeaseExpiresAt = &leaseExpiresAt.Time
		}
		if completedAt.Valid {
			t, _ := time.Parse(time.RFC3339, completedAt.String)
			job.CompletedAt = &t
		}
		
		if len(jobs) < limit {
			jobs = append(jobs, &job)
		} else {
			nextCursor = job.JobID
			break
		}
	}
	
	return jobs, nextCursor, nil
}
