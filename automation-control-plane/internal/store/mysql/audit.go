package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	AuditID     string
	TenantID    string
	ProjectID   *string
	ActorType   string
	ActorID     string
	Action      string
	ResourceType *string
	ResourceID  *string
	Metadata    json.RawMessage
	CreatedAt   time.Time
}

// CreateAuditLog creates a new audit log entry
func (s *Store) CreateAuditLog(ctx context.Context, qb *QueryBuilder, log *AuditLog) error {
	query := `INSERT INTO audit_logs (audit_id, tenant_id, project_id, actor_type, actor_id, 
	          action, resource_type, resource_id, metadata, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())`
	
	_, err := s.db.ExecContext(ctx, query,
		log.AuditID, log.TenantID, log.ProjectID, log.ActorType, log.ActorID,
		log.Action, log.ResourceType, log.ResourceID, log.Metadata,
	)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	
	return nil
}

// ListAuditLogs lists audit logs with filtering and pagination
func (s *Store) ListAuditLogs(ctx context.Context, qb *QueryBuilder, filters AuditLogFilters, limit int, cursor string) ([]*AuditLog, string, error) {
	where, args := qb.BuildWhereClause("")
	
	if filters.ProjectID != nil {
		where += " AND project_id = ?"
		args = append(args, *filters.ProjectID)
	}
	if filters.ActorID != nil {
		where += " AND actor_id = ?"
		args = append(args, *filters.ActorID)
	}
	if filters.Action != nil {
		where += " AND action = ?"
		args = append(args, *filters.Action)
	}
	if cursor != "" {
		where += " AND audit_id > ?"
		args = append(args, cursor)
	}
	
	query := fmt.Sprintf(`SELECT audit_id, tenant_id, project_id, actor_type, actor_id, 
	                     action, resource_type, resource_id, metadata, created_at
	                     FROM audit_logs WHERE %s ORDER BY audit_id LIMIT ?`, where)
	args = append(args, limit+1)
	
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()
	
	var logs []*AuditLog
	var nextCursor string
	
	for rows.Next() {
		var log AuditLog
		var projectID, resourceType, resourceID sql.NullString
		
		err := rows.Scan(
			&log.AuditID, &log.TenantID, &projectID, &log.ActorType, &log.ActorID,
			&log.Action, &resourceType, &resourceID, &log.Metadata, &log.CreatedAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan audit log: %w", err)
		}
		
		if projectID.Valid {
			log.ProjectID = &projectID.String
		}
		if resourceType.Valid {
			log.ResourceType = &resourceType.String
		}
		if resourceID.Valid {
			log.ResourceID = &resourceID.String
		}
		
		if len(logs) < limit {
			logs = append(logs, &log)
		} else {
			nextCursor = log.AuditID
			break
		}
	}
	
	return logs, nextCursor, nil
}

// AuditLogFilters provides filtering options for audit logs
type AuditLogFilters struct {
	ProjectID *string
	ActorID   *string
	Action    *string
}
