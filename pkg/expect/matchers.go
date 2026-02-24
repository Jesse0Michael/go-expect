package expect

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
)

// Matcher is implemented by any value that can assert itself against an actual value.
// When the JSON body is parsed, field values that implement Matcher are invoked
// instead of doing a DeepEqual comparison.
type Matcher interface {
	Match(actual any) error
}

// ---- String matchers ----

// Contains asserts the actual string contains the given substring.
type Contains string

func (m Contains) Match(actual any) error {
	s, ok := actual.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", actual)
	}
	re := regexp.MustCompile(regexp.QuoteMeta(string(m)))
	if !re.MatchString(s) {
		return fmt.Errorf("%q does not contain %q", s, string(m))
	}
	return nil
}

// Matches asserts the actual string matches the given regular expression.
type Matches string

func (m Matches) Match(actual any) error {
	s, ok := actual.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", actual)
	}
	re, err := regexp.Compile(string(m))
	if err != nil {
		return fmt.Errorf("invalid regex %q: %w", string(m), err)
	}
	if !re.MatchString(s) {
		return fmt.Errorf("%q does not match regex %q", s, string(m))
	}
	return nil
}

// NotEmpty asserts the actual value is non-nil and non-zero.
type NotEmpty struct{}

func (NotEmpty) Match(actual any) error {
	if actual == nil {
		return fmt.Errorf("expected non-empty value, got nil")
	}
	if reflect.DeepEqual(actual, reflect.Zero(reflect.TypeOf(actual)).Interface()) {
		return fmt.Errorf("expected non-empty value, got zero value")
	}
	return nil
}

// ---- Numeric matchers ----

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	}
	return 0, false
}

// Gt asserts actual > n.
type Gt float64

func (m Gt) Match(actual any) error {
	f, ok := toFloat(actual)
	if !ok {
		return fmt.Errorf("expected number, got %T", actual)
	}
	if !(f > float64(m)) {
		return fmt.Errorf("expected > %v, got %v", float64(m), f)
	}
	return nil
}

// Gte asserts actual >= n.
type Gte float64

func (m Gte) Match(actual any) error {
	f, ok := toFloat(actual)
	if !ok {
		return fmt.Errorf("expected number, got %T", actual)
	}
	if !(f >= float64(m)) {
		return fmt.Errorf("expected >= %v, got %v", float64(m), f)
	}
	return nil
}

// Lt asserts actual < n.
type Lt float64

func (m Lt) Match(actual any) error {
	f, ok := toFloat(actual)
	if !ok {
		return fmt.Errorf("expected number, got %T", actual)
	}
	if !(f < float64(m)) {
		return fmt.Errorf("expected < %v, got %v", float64(m), f)
	}
	return nil
}

// Lte asserts actual <= n.
type Lte float64

func (m Lte) Match(actual any) error {
	f, ok := toFloat(actual)
	if !ok {
		return fmt.Errorf("expected number, got %T", actual)
	}
	if !(f <= float64(m)) {
		return fmt.Errorf("expected <= %v, got %v", float64(m), f)
	}
	return nil
}

// ---- Array matchers ----

// Length asserts the actual array or string has exactly n elements.
type Length int

func (m Length) Match(actual any) error {
	v := reflect.ValueOf(actual)
	switch v.Kind() {
	case reflect.Slice, reflect.Array, reflect.String, reflect.Map:
		if v.Len() != int(m) {
			return fmt.Errorf("expected length %d, got %d", int(m), v.Len())
		}
		return nil
	}
	return fmt.Errorf("expected slice/string/map, got %T", actual)
}

// ---- Status code matchers ----

// AnyOf asserts the actual HTTP status code is one of the given codes.
type AnyOf []int

func (m AnyOf) MatchStatus(actual int) error {
	if slices.Contains([]int(m), actual) {
		return nil
	}
	return fmt.Errorf("expected status one of %v, got %d", []int(m), actual)
}
