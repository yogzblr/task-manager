package probe

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DBTask performs database operations
type DBTask struct {
	Driver  string
	DSN     string
	Query   string
	Timeout time.Duration
}

// Configure sets up the database task
func (t *DBTask) Configure(config map[string]interface{}) error {
	// Driver is required
	driver, ok := config["driver"].(string)
	if !ok || driver == "" {
		return fmt.Errorf("driver is required")
	}
	t.Driver = driver
	
	// DSN is required
	dsn, ok := config["dsn"].(string)
	if !ok || dsn == "" {
		return fmt.Errorf("dsn is required")
	}
	t.DSN = dsn
	
	// Query is required
	query, ok := config["query"].(string)
	if !ok || query == "" {
		return fmt.Errorf("query is required")
	}
	t.Query = query
	
	// Timeout (default: 30s)
	if timeoutStr, ok := config["timeout"].(string); ok {
		duration, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout: %w", err)
		}
		t.Timeout = duration
	} else {
		t.Timeout = 30 * time.Second
	}
	
	return nil
}

// Execute performs the database query
func (t *DBTask) Execute(ctx context.Context) (interface{}, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, t.Timeout)
	defer cancel()
	
	// Connect to database
	db, err := sql.Open(t.Driver, t.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	
	// Test connection
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	// Execute query
	rows, err := db.QueryContext(ctx, t.Query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()
	
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}
	
	// Read all rows
	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	
	return map[string]interface{}{
		"rows":  results,
		"count": len(results),
	}, nil
}
