package expect

import (
	"encoding/json"
	"testing"
)

func TestSQLExpect_Validate(t *testing.T) {
	t.Run("row count match", func(t *testing.T) {
		count := 2
		exp := &SQLExpect{RowCount: &count}
		result := &SQLResult{Rows: []map[string]any{{"id": float64(1)}, {"id": float64(2)}}}
		if err := exp.Validate(result, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("row count mismatch", func(t *testing.T) {
		count := 3
		exp := &SQLExpect{RowCount: &count}
		result := &SQLResult{Rows: []map[string]any{{"id": float64(1)}}}
		if err := exp.Validate(result, nil); err == nil {
			t.Fatal("expected error for row count mismatch")
		}
	})

	t.Run("rows affected match", func(t *testing.T) {
		affected := int64(1)
		exp := &SQLExpect{RowsAffected: &affected}
		result := &SQLResult{RowsAffected: 1}
		if err := exp.Validate(result, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rows affected mismatch", func(t *testing.T) {
		affected := int64(2)
		exp := &SQLExpect{RowsAffected: &affected}
		result := &SQLResult{RowsAffected: 1}
		if err := exp.Validate(result, nil); err == nil {
			t.Fatal("expected error for rows affected mismatch")
		}
	})

	t.Run("row partial match", func(t *testing.T) {
		exp := &SQLExpect{
			Rows: []ExpectBody{
				ExpectBody(`{"name":"alice"}`),
			},
		}
		result := &SQLResult{Rows: []map[string]any{
			{"id": float64(1), "name": "alice", "email": "alice@example.com"},
		}}
		if err := exp.Validate(result, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("row partial mismatch", func(t *testing.T) {
		exp := &SQLExpect{
			Rows: []ExpectBody{
				ExpectBody(`{"name":"bob"}`),
			},
		}
		result := &SQLResult{Rows: []map[string]any{
			{"id": float64(1), "name": "alice"},
		}}
		if err := exp.Validate(result, nil); err == nil {
			t.Fatal("expected error for row mismatch")
		}
	})

	t.Run("save from first row", func(t *testing.T) {
		exp := &SQLExpect{
			Save: []SaveEntry{{Field: "id", As: "user_id"}},
		}
		result := &SQLResult{Rows: []map[string]any{
			{"id": float64(42), "name": "alice"},
		}}
		vars := make(VarStore)
		if err := exp.Validate(result, vars); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if vars["user_id"] != float64(42) {
			t.Fatalf("expected user_id=42, got %v", vars["user_id"])
		}
	})
}

func TestNormalizeValue(t *testing.T) {
	tests := []struct {
		input any
		want  any
	}{
		{[]byte("hello"), "hello"},
		{int64(42), float64(42)},
		{int32(42), float64(42)},
		{"text", "text"},
		{float64(3.14), float64(3.14)},
		{nil, nil},
	}
	for _, tt := range tests {
		got := normalizeValue(tt.input)
		gotJSON, _ := json.Marshal(got)
		wantJSON, _ := json.Marshal(tt.want)
		if string(gotJSON) != string(wantJSON) {
			t.Errorf("normalizeValue(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestBuildFileSQLStep(t *testing.T) {
	t.Run("query with row expectations", func(t *testing.T) {
		rowCount := 1
		s := fileStep{
			Request: &fileRequest{
				Connection: "mydb",
				Statement:  "SELECT * FROM users WHERE id = $1",
				Params:     []any{float64(1)},
			},
			Expect: &fileExpectation{
				RowCount: &rowCount,
				Rows:     []any{map[string]any{"name": "alice"}},
				Save:     []fileSaveEntry{{Field: "id", As: "user_id"}},
			},
		}
		b, err := buildFileSQLStep(s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		step := b.Build()
		req, ok := step.Request.(*SQLRequest)
		if !ok {
			t.Fatal("expected SQLRequest")
		}
		if req.Statement != "SELECT * FROM users WHERE id = $1" {
			t.Fatalf("unexpected statement: %s", req.Statement)
		}
		if req.Exec {
			t.Fatal("expected query, not exec")
		}
		exp, ok := step.Expect.(*SQLExpect)
		if !ok {
			t.Fatal("expected SQLExpect")
		}
		if exp.RowCount == nil || *exp.RowCount != 1 {
			t.Fatal("expected row count of 1")
		}
		if len(exp.Rows) != 1 {
			t.Fatal("expected 1 expected row")
		}
		if len(exp.Save) != 1 || exp.Save[0].Field != "id" {
			t.Fatal("expected save entry")
		}
	})

	t.Run("exec step", func(t *testing.T) {
		s := fileStep{
			Request: &fileRequest{
				Connection: "mydb",
				Statement:  "INSERT INTO users (name) VALUES ($1)",
				Params:     []any{"alice"},
				Exec:       true,
			},
			Expect: &fileExpectation{},
		}
		b, err := buildFileSQLStep(s)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		step := b.Build()
		req := step.Request.(*SQLRequest)
		if !req.Exec {
			t.Fatal("expected exec")
		}
	})
}
