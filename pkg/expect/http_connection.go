package expect

import (
	"net/http"
	"time"
)

// HTTPConnection is a connection to an HTTP/HTTPS service.
type HTTPConnection struct {
	Name    string
	URL     string
	Timeout time.Duration // per-request timeout; 0 means use DefaultHTTPTimeout
	Client  *http.Client  // nil means use http.DefaultClient
}

func (c *HTTPConnection) Type() string    { return "http" }
func (c *HTTPConnection) GetName() string { return c.Name }

// HTTP is a convenience constructor for an HTTPConnection.
func HTTP(name, url string) *HTTPConnection {
	return &HTTPConnection{Name: name, URL: url}
}
