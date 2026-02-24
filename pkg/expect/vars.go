package expect

import (
	"fmt"
	"strings"
)

// VarStore holds variables that can be set by one step and consumed by later steps.
type VarStore map[string]any

// Interpolate replaces all {key} placeholders in s with values from the store.
// Unknown keys are left as-is.
func (v VarStore) Interpolate(s string) string {
	for key, val := range v {
		s = strings.ReplaceAll(s, "{"+key+"}", fmt.Sprintf("%v", val))
	}
	return s
}

// InterpolateBytes replaces {key} placeholders in a byte slice.
func (v VarStore) InterpolateBytes(b []byte) []byte {
	return []byte(v.Interpolate(string(b)))
}
