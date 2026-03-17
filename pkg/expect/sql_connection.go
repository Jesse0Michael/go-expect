package expect

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DefaultSQLTimeout is applied to every SQL query unless overridden on the connection.
const DefaultSQLTimeout = 30 * time.Second

// SQLConnection is a connection to a SQL database.
type SQLConnection struct {
	Name    string
	DSN     string
	Driver  string
	Timeout time.Duration // per-query timeout; 0 means use DefaultSQLTimeout
	DB      *sql.DB
}

// SQL creates a SQLConnection with an explicit driver.
func SQL(name, driver, dsn string) *SQLConnection {
	return &SQLConnection{Name: name, DSN: dsn, Driver: driver}
}

func (c *SQLConnection) Type() string    { return c.Driver }
func (c *SQLConnection) GetName() string { return c.Name }

// Dial opens the underlying database connection, dialling if necessary.
func (c *SQLConnection) Dial() error {
	if c.DB != nil {
		return nil
	}
	db, err := sql.Open(c.Driver, c.DSN)
	if err != nil {
		return fmt.Errorf("sql open %q: %w", c.DSN, err)
	}
	c.DB = db
	return nil
}

// Close closes the database connection.
func (c *SQLConnection) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

// QueryContext executes a query and returns rows as []map[string]any.
func (c *SQLConnection) QueryContext(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	if err := c.Dial(); err != nil {
		return nil, err
	}

	rows, err := c.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("sql query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("sql columns: %w", err)
	}

	var results []map[string]any
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("sql scan: %w", err)
		}

		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = normalizeValue(values[i])
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

// ExecContext executes a statement and returns the number of rows affected.
func (c *SQLConnection) ExecContext(ctx context.Context, query string, args ...any) (int64, error) {
	if err := c.Dial(); err != nil {
		return 0, err
	}

	result, err := c.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("sql exec: %w", err)
	}
	return result.RowsAffected()
}

// normalizeValue converts driver types to JSON-friendly Go types.
func normalizeValue(v any) any {
	switch val := v.(type) {
	case []byte:
		return string(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	default:
		return val
	}
}
