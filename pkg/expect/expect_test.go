package expect

import (
	"reflect"
	"testing"
)

func TestExpectBody_structured(t *testing.T) {
	tests := []struct {
		name string
		e    ExpectBody
		want map[string]any
		ok   bool
	}{
		{
			name: "nil body",
			e:    nil,
			want: nil,
			ok:   false,
		},
		{
			name: "text body",
			e:    ExpectBody([]byte("hello")),
			want: nil,
			ok:   false,
		},
		{
			name: "numeric body",
			e:    ExpectBody([]byte("12345")),
			want: nil,
			ok:   false,
		},
		{
			name: "boolean body",
			e:    ExpectBody([]byte("true")),
			want: nil,
			ok:   false,
		},
		{
			name: "JSON object body",
			e:    ExpectBody([]byte(`{"key":"value","number":42}`)),
			want: map[string]any{"key": "value", "number": float64(42)},
			ok:   true,
		},
		{
			name: "YAML body",
			e:    ExpectBody([]byte("key: value\nnumber: 42")),
			want: nil,
			ok:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.e.structured()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExpectBody.structured() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.ok {
				t.Errorf("ExpectBody.structured() got1 = %v, want %v", got1, tt.ok)
			}
		})
	}
}
