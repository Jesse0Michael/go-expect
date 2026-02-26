package expect

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/tidwall/gjson"
)

// ExpectBody holds the expected response body and validates it against actual bytes.
type ExpectBody []byte

func (e ExpectBody) structured() (map[string]any, bool) {
	var result map[string]any
	if err := json.Unmarshal(e, &result); err != nil {
		return nil, false
	}
	return result, true
}

// Validate checks that actual matches the expected body (partial JSON match or exact bytes).
func (e ExpectBody) Validate(actual []byte) error {
	if structuredExpected, ok := e.structured(); ok {
		if structuredActual, ok := ExpectBody(actual).structured(); ok {
			return partialMatch(structuredActual, structuredExpected)
		}
	}

	if string(actual) != string(e) {
		return fmt.Errorf("unexpected body: %s", string(actual))
	}

	return nil
}

// partialMatch recursively checks that actual satisfies expected.
// expected values may implement Matcher for custom assertions.
func partialMatch(actual, expected any) error {
	// If expected is a Matcher, delegate to it.
	if m, ok := expected.(Matcher); ok {
		return m.Match(actual)
	}

	switch exp := expected.(type) {
	case map[string]any:
		actMap, ok := actual.(map[string]any)
		if !ok {
			return fmt.Errorf("expected object, got %T", actual)
		}
		for key, expVal := range exp {
			actVal, exists := actMap[key]
			if !exists {
				return fmt.Errorf("missing field %q", key)
			}
			if err := partialMatch(actVal, expVal); err != nil {
				return fmt.Errorf("field %q: %w", key, err)
			}
		}
		return nil

	case []any:
		actSlice, ok := actual.([]any)
		if !ok {
			return fmt.Errorf("expected array, got %T", actual)
		}
		if len(exp) == 0 {
			return nil
		}
		for i, expElem := range exp {
			found := false
			for _, actElem := range actSlice {
				if partialMatch(actElem, expElem) == nil {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("array element [%d] not found in actual", i)
			}
		}
		return nil

	default:
		if !reflect.DeepEqual(actual, expected) {
			return fmt.Errorf("expected %v, got %v", expected, actual)
		}
		return nil
	}
}

// SaveEntry defines a field to extract from the response body into a variable.
// Field uses json path notation (e.g. "id", "user.name", "items.0.id").
type SaveEntry struct {
	Field string
	As    string
}

// saveFromJSON extracts fields from JSON bytes into vars using gjson paths.
func saveFromJSON(data []byte, entries []SaveEntry, vars VarStore) {
	for _, entry := range entries {
		if result := gjson.GetBytes(data, entry.Field); result.Exists() {
			vars[entry.As] = result.Value()
		}
	}
}

// HTTPExpect defines the expected HTTP response and optional variable extractions.
type HTTPExpect struct {
	Status    int
	StatusAny AnyOf // if set, status must be one of these codes
	Body      ExpectBody
	Header    map[string]string
	Save      []SaveEntry
}

// Validate checks the response against expectations, saving extracted values into vars.
func (e *HTTPExpect) Validate(resp *http.Response, vars VarStore) error {
	if len(e.StatusAny) > 0 {
		if err := e.StatusAny.MatchStatus(resp.StatusCode); err != nil {
			return err
		}
	} else if e.Status != 0 && resp.StatusCode != e.Status {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	for k, v := range e.Header {
		if resp.Header.Get(k) != v {
			return fmt.Errorf("unexpected header %s: %s", k, resp.Header.Get(k))
		}
	}

	var bodyBytes []byte
	if e.Body != nil || len(e.Save) > 0 {
		var err error
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
	}

	if e.Body != nil {
		if err := e.Body.Validate(bodyBytes); err != nil {
			return err
		}
	}

	if len(e.Save) > 0 && vars != nil {
		saveFromJSON(bodyBytes, e.Save, vars)
	}

	return nil
}
