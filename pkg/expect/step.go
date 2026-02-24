package expect

import (
	"fmt"
	"net/http"
)

// Step is a single request/response pair within a scenario.
type Step struct {
	Connection string
	Request    any
	Expect     any
}

// Run executes the step against the given connection, applying variable interpolation.
func (s *Step) Run(conn Connection, vars VarStore) error {
	if s.Request == nil {
		return nil
	}
	switch req := s.Request.(type) {
	case *HTTPRequest:
		httpConn, ok := conn.(*HTTPConnection)
		if !ok {
			return fmt.Errorf("mismatched connection type for HTTP request: %T", conn)
		}
		resp, err := req.Run(httpConn, vars)
		if err != nil {
			return err
		}
		return s.validateHTTP(resp, vars)

	case *GRPCRequest:
		grpcConn, ok := conn.(*GRPCConnection)
		if !ok {
			return fmt.Errorf("mismatched connection type for gRPC request: %T", conn)
		}
		respBytes, grpcErr := req.Run(grpcConn, vars)
		return s.validateGRPC(respBytes, grpcErr, vars)

	default:
		return fmt.Errorf("unsupported request type: %T", s.Request)
	}
}

func (s *Step) validateHTTP(resp *http.Response, vars VarStore) error {
	if s.Expect == nil {
		return nil
	}
	switch exp := s.Expect.(type) {
	case *HTTPExpect:
		return exp.Validate(resp, vars)
	case HTTPExpect:
		return exp.Validate(resp, vars)
	default:
		return fmt.Errorf("mismatched expect type for HTTP request: %T", s.Expect)
	}
}

func (s *Step) validateGRPC(respBytes []byte, grpcErr error, vars VarStore) error {
	if s.Expect == nil {
		return grpcErr
	}
	switch exp := s.Expect.(type) {
	case *GRPCExpect:
		return exp.Validate(respBytes, grpcErr, vars)
	case GRPCExpect:
		return exp.Validate(respBytes, grpcErr, vars)
	default:
		return fmt.Errorf("mismatched expect type for gRPC request: %T", s.Expect)
	}
}
