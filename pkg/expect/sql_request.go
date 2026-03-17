package expect

import (
	"context"
	"encoding/json"
	"fmt"
)

// SQLRequest describes a SQL query or exec to run against a SQLConnection.
type SQLRequest struct {
	Statement string
	Params    []any
	Exec      bool // true for INSERT/UPDATE/DELETE, false for SELECT
}

// SQLResult holds the result of a SQL request.
type SQLResult struct {
	Rows         []map[string]any
	RowsAffected int64
}

// Run executes the SQL request against conn, interpolating variables from vars.
func (r *SQLRequest) Run(conn *SQLConnection, vars VarStore) (*SQLResult, error) {
	timeout := conn.Timeout
	if timeout == 0 {
		timeout = DefaultSQLTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	stmt := vars.Interpolate(r.Statement)

	params := make([]any, len(r.Params))
	for i, p := range r.Params {
		if s, ok := p.(string); ok {
			params[i] = vars.Interpolate(s)
		} else {
			params[i] = p
		}
	}

	if r.Exec {
		affected, err := conn.ExecContext(ctx, stmt, params...)
		if err != nil {
			return nil, err
		}
		return &SQLResult{RowsAffected: affected}, nil
	}

	rows, err := conn.QueryContext(ctx, stmt, params...)
	if err != nil {
		return nil, err
	}
	return &SQLResult{Rows: rows}, nil
}

// SQLExpect validates a SQL result.
type SQLExpect struct {
	RowCount     *int
	RowsAffected *int64
	Rows         []ExpectBody
	Save         []SaveEntry
}

// Validate checks the result against expectations, saving extracted values into vars.
func (e *SQLExpect) Validate(result *SQLResult, vars VarStore) error {
	if e.RowCount != nil {
		if len(result.Rows) != *e.RowCount {
			return fmt.Errorf("unexpected row count: got %d, want %d", len(result.Rows), *e.RowCount)
		}
	}

	if e.RowsAffected != nil {
		if result.RowsAffected != *e.RowsAffected {
			return fmt.Errorf("unexpected rows affected: got %d, want %d", result.RowsAffected, *e.RowsAffected)
		}
	}

	for i, expectedRow := range e.Rows {
		if i >= len(result.Rows) {
			return fmt.Errorf("expected row [%d] but only got %d rows", i, len(result.Rows))
		}
		actualJSON, err := json.Marshal(result.Rows[i])
		if err != nil {
			return fmt.Errorf("marshal actual row [%d]: %w", i, err)
		}
		if err := expectedRow.Validate(actualJSON); err != nil {
			return fmt.Errorf("row [%d]: %w", i, err)
		}
	}

	if len(e.Save) > 0 && vars != nil && len(result.Rows) > 0 {
		firstRow, err := json.Marshal(result.Rows[0])
		if err == nil {
			saveFromJSON(firstRow, e.Save, vars)
		}
	}

	return nil
}
