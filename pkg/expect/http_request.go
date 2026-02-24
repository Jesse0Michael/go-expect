package expect

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"time"
)

// DefaultHTTPTimeout is applied to every HTTP request unless overridden on the connection.
const DefaultHTTPTimeout = 30 * time.Second

// HTTPRequest describes an outbound HTTP request.
type HTTPRequest struct {
	Method  string
	Path    string
	Body    []byte
	Header  map[string]string
	Query   map[string]string
	Timeout time.Duration // 0 means use DefaultHTTPTimeout
}

// Run executes the HTTP request against conn, interpolating variables from vars.
func (r *HTTPRequest) Run(conn *HTTPConnection, vars VarStore) (*http.Response, error) {
	path := vars.Interpolate(r.Path)
	url := strings.TrimRight(conn.URL, "/") + "/" + strings.TrimLeft(path, "/")

	body := vars.InterpolateBytes(r.Body)

	timeout := r.Timeout
	if timeout == 0 {
		timeout = conn.Timeout
	}
	if timeout == 0 {
		timeout = DefaultHTTPTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, r.Method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for k, v := range r.Query {
		q.Add(k, vars.Interpolate(v))
	}
	req.URL.RawQuery = q.Encode()

	for k, v := range r.Header {
		req.Header.Set(k, vars.Interpolate(v))
	}

	client := conn.Client
	if client == nil {
		client = http.DefaultClient
	}

	return client.Do(req)
}
