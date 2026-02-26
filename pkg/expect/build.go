package expect

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
)

// buildSuite performs a two-pass build over a set of parsed files:
// first collecting all connections, then building scenarios with full connection context.
func buildSuite(files []expectFile) (*Suite, error) {
	var allConns []Connection
	for _, f := range files {
		conns, err := buildFileConnections(f)
		if err != nil {
			return nil, err
		}
		allConns = append(allConns, conns...)
	}
	connMap, defaultConn := buildConnMap(allConns)

	var scenarios []*Scenario
	for _, f := range files {
		ss, err := buildFileScenarios(f, connMap, defaultConn)
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, ss...)
	}

	return NewSuite().
		WithConnections(slices.Collect(maps.Values(connMap))...).
		WithScenarios(scenarios...), nil
}

func buildConnMap(conns []Connection) (map[string]Connection, Connection) {
	m := make(map[string]Connection, len(conns))
	var def Connection
	for _, c := range conns {
		name := c.GetName()
		m[name] = c
		if def == nil || name == "" {
			def = c
		}
	}
	return m, def
}

func buildFileConnections(f expectFile) ([]Connection, error) {
	var conns []Connection
	for _, c := range f.Connections {
		conn, err := buildFileConnection(c)
		if err != nil {
			return nil, err
		}
		conns = append(conns, conn)
	}
	return conns, nil
}

func buildFileScenarios(f expectFile, connMap map[string]Connection, defaultConn Connection) ([]*Scenario, error) {
	var scenarios []*Scenario
	for _, s := range f.Scenarios {
		sc := NewScenario(s.Name)
		for _, st := range s.Steps {
			if st.Request == nil {
				continue
			}
			step, err := buildFileStep(st, connMap, defaultConn)
			if err != nil {
				return nil, fmt.Errorf("scenario %q: %w", s.Name, err)
			}
			sc.AddStep(step)
		}
		scenarios = append(scenarios, sc)
	}
	return scenarios, nil
}

func buildFileConnection(c fileConnection) (Connection, error) {
	switch c.Type {
	case "http", "https", "":
		return HTTP(c.Name, c.URL), nil
	case "grpc":
		return GRPC(c.Name, c.URL), nil
	default:
		return nil, fmt.Errorf("go-expect: unknown connection type %q", c.Type)
	}
}

func buildFileStep(s fileStep, connMap map[string]Connection, defaultConn Connection) (*StepBuilder, error) {
	conn := defaultConn
	if c, ok := connMap[s.Request.Connection]; ok {
		conn = c
	}
	switch conn.(type) {
	case *HTTPConnection:
		return buildFileHTTPStep(s)
	case *GRPCConnection:
		return buildFileGRPCStep(s)
	default:
		return nil, fmt.Errorf("go-expect: unsupported connection type %T", conn)
	}
}

func buildFileHTTPStep(s fileStep) (*StepBuilder, error) {
	r := s.Request
	b := HTTPStep(r.Method, r.Endpoint).WithConnection(r.Connection)
	for k, v := range r.Header {
		b.WithHeader(k, v)
	}
	for k, v := range r.Query {
		b.WithQuery(k, v)
	}
	if r.Body != nil {
		body, err := json.Marshal(r.Body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		b.WithBody(body).WithHeader("Content-Type", "application/json")
	}
	if s.Expect != nil {
		e := s.Expect
		b.ExpectStatus(e.Status)
		for k, v := range e.Header {
			b.ExpectHeader(k, v)
		}
		if e.Body != nil {
			body, err := json.Marshal(e.Body)
			if err != nil {
				return nil, fmt.Errorf("marshal expect body: %w", err)
			}
			b.ExpectBody(body)
		}
		for _, sv := range e.Save {
			b.Save(sv.Field, sv.As)
		}
	}
	return b, nil
}

func buildFileGRPCStep(s fileStep) (*StepBuilder, error) {
	r := s.Request
	var body []byte
	if r.Body != nil {
		var err error
		body, err = json.Marshal(r.Body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
	}
	b := GRPCRawCall(r.Connection, r.Endpoint, body)
	for k, v := range r.Header {
		b.WithHeader(k, v)
	}
	if s.Expect != nil {
		e := s.Expect
		b.ExpectGRPCCode(e.Code)
		if e.Body != nil {
			body, err := json.Marshal(e.Body)
			if err != nil {
				return nil, fmt.Errorf("marshal expect body: %w", err)
			}
			b.ExpectGRPCBody(body)
		}
		for _, sv := range e.Save {
			b.SaveGRPC(sv.Field, sv.As)
		}
	}
	return b, nil
}
