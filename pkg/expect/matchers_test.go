package expect

import "testing"

func TestContains(t *testing.T) {
	if err := Contains("ello").Match("hello world"); err != nil {
		t.Error(err)
	}
	if err := Contains("xyz").Match("hello world"); err == nil {
		t.Error("expected error for missing substring")
	}
	if err := Contains("x").Match(42); err == nil {
		t.Error("expected error for non-string actual")
	}
}

func TestMatches(t *testing.T) {
	if err := Matches(`^\d+$`).Match("12345"); err != nil {
		t.Error(err)
	}
	if err := Matches(`^\d+$`).Match("abc"); err == nil {
		t.Error("expected error for non-matching string")
	}
}

func TestNotEmpty(t *testing.T) {
	if err := (NotEmpty{}).Match("hello"); err != nil {
		t.Error(err)
	}
	if err := (NotEmpty{}).Match(""); err == nil {
		t.Error("expected error for empty string")
	}
	if err := (NotEmpty{}).Match(nil); err == nil {
		t.Error("expected error for nil")
	}
}

func TestNumericMatchers(t *testing.T) {
	if err := Gt(1).Match(float64(2)); err != nil {
		t.Error(err)
	}
	if err := Gt(5).Match(float64(2)); err == nil {
		t.Error("expected error: 2 not > 5")
	}
	if err := Gte(2).Match(float64(2)); err != nil {
		t.Error(err)
	}
	if err := Lt(5).Match(float64(2)); err != nil {
		t.Error(err)
	}
	if err := Lte(2).Match(float64(2)); err != nil {
		t.Error(err)
	}
}

func TestLength(t *testing.T) {
	if err := Length(3).Match([]any{1, 2, 3}); err != nil {
		t.Error(err)
	}
	if err := Length(2).Match([]any{1, 2, 3}); err == nil {
		t.Error("expected error for wrong length")
	}
	if err := Length(5).Match("hello"); err != nil {
		t.Error(err)
	}
}

func TestAnyOf(t *testing.T) {
	if err := (AnyOf{200, 201}).MatchStatus(201); err != nil {
		t.Error(err)
	}
	if err := (AnyOf{200, 201}).MatchStatus(404); err == nil {
		t.Error("expected error for unmatched status")
	}
}

func TestPartialMatch_withMatcher(t *testing.T) {
	actual := map[string]any{
		"count": float64(5),
		"name":  "alice",
	}
	expected := map[string]any{
		"count": Gte(3),
		"name":  Contains("ali"),
	}
	if err := partialMatch(actual, expected); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPartialMatch_withMatcherFail(t *testing.T) {
	actual := map[string]any{"count": float64(1)}
	expected := map[string]any{"count": Gt(10)}
	if err := partialMatch(actual, expected); err == nil {
		t.Error("expected error: 1 not > 10")
	}
}
