package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Store provides MySQL database access with row-level security
type Store struct {
	db *sql.DB
}

// Config holds MySQL connection configuration
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// NewStore creates a new MySQL store with connection pooling
func NewStore(ctx context.Context, cfg Config) (*Store, error) {
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Verify connection
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// DB returns the underlying database connection (for transactions)
func (s *Store) DB() *sql.DB {
	return s.db
}

// BeginTx starts a new transaction
func (s *Store) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return s.db.BeginTx(ctx, opts)
}

// QueryBuilder helps build queries with mandatory tenant_id and project_id filters
type QueryBuilder struct {
	tenantID           string
	authorizedProjects []string
}

// NewQueryBuilder creates a new query builder with security context
func NewQueryBuilder(tenantID string, authorizedProjects []string) *QueryBuilder {
	return &QueryBuilder{
		tenantID:           tenantID,
		authorizedProjects: authorizedProjects,
	}
}

// BuildWhereClause builds a WHERE clause with mandatory security filters
func (qb *QueryBuilder) BuildWhereClause(baseWhere string) (string, []interface{}) {
	args := []interface{}{qb.tenantID}
	
	where := baseWhere
	if where != "" {
		where += " AND "
	}
	where += "tenant_id = ?"
	
	if len(qb.authorizedProjects) > 0 {
		where += " AND project_id IN ("
		placeholders := ""
		for i, pid := range qb.authorizedProjects {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
			args = append(args, pid)
		}
		where += placeholders + ")"
	}
	
	return where, args
}

// ValidateTenantProject ensures tenant_id and project_id are valid for the context
func (qb *QueryBuilder) ValidateTenantProject(projectID string) error {
	if qb.tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}
	
	// If projectID is empty and we have authorized projects, allow it
	// The WHERE clause will filter by authorized projects
	if projectID == "" && len(qb.authorizedProjects) > 0 {
		return nil
	}
	
	if len(qb.authorizedProjects) > 0 {
		authorized := false
		for _, pid := range qb.authorizedProjects {
			if pid == projectID {
				authorized = true
				break
			}
		}
		if !authorized {
			return fmt.Errorf("project_id %s not authorized", projectID)
		}
	}
	
	return nil
}
