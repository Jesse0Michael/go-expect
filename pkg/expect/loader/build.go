package loader

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/jesse0michael/go-expect/pkg/expect"
)

// buildSuite performs a two-pass build over a set of parsed files:
// first collecting all connections, then building scenarios with full connection context.
func buildSuite(files []expectFile) (*expect.Suite, error) {
	var allConns []expect.Connection
	for _, f := range files {
		conns, err := buildConnections(f)
		if err != nil {
			return nil, err
		}
		allConns = append(allConns, conns...)
	}
	connMap, defaultConn := connectionMap(allConns)

	var scenarios []*expect.Scenario
	for _, f := range files {
		ss, err := buildScenarios(f, connMap, defaultConn)
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, ss...)
	}

	return expect.NewSuite().
		WithConnections(slices.Collect(maps.Values(connMap))...).
		WithScenarios(scenarios...), nil
}

// connectionMap builds a name→Connection map and picks the first entry, or a connection with an empty name, as the default.
func connectionMap(conns []expect.Connection) (map[string]expect.Connection, expect.Connection) {
	m := make(map[string]expect.Connection, len(conns))
	var def expect.Connection
	for _, c := range conns {
		name := c.GetName()
		m[name] = c
		if def == nil || name == "" {
			def = c
		}
	}
	return m, def
}

func buildConnections(f expectFile) ([]expect.Connection, error) {
	var conns []expect.Connection
	for _, c := range f.Connections {
		conn, err := buildConnection(c)
		if err != nil {
			return nil, err
		}
		conns = append(conns, conn)
	}
	return conns, nil
}

func buildScenarios(f expectFile, connMap map[string]expect.Connection, defaultConn expect.Connection) ([]*expect.Scenario, error) {
	var scenarios []*expect.Scenario
	for _, s := range f.Scenarios {
		sc := expect.NewScenario(s.Name)
		for _, st := range s.Steps {
			if st.Request == nil {
				continue
			}
			step, err := buildStep(st, connMap, defaultConn)
			if err != nil {
				return nil, fmt.Errorf("scenario %q: %w", s.Name, err)
			}
			sc.AddStep(step)
		}
		scenarios = append(scenarios, sc)
	}
	return scenarios, nil
}

func buildConnection(c connection) (expect.Connection, error) {
	switch c.Type {
	case "http", "https", "":
		return expect.HTTP(c.Name, c.URL), nil
	case "grpc":
		return expect.GRPC(c.Name, c.URL), nil
	default:
		return nil, fmt.Errorf("go-expect: unknown connection type %q", c.Type)
	}
}

func buildStep(s step, connMap map[string]expect.Connection, defaultConn expect.Connection) (*expect.StepBuilder, error) {
	conn := defaultConn
	if c, ok := connMap[s.Request.Connection]; ok {
		conn = c
	}
	switch conn.(type) {
	case *expect.HTTPConnection:
		return buildHTTPStep(s)
	case *expect.GRPCConnection:
		return buildGRPCStep(s)
	default:
		return nil, fmt.Errorf("go-expect: unsupported connection type %T", conn)
	}
}

func buildHTTPStep(s step) (*expect.StepBuilder, error) {
	r := s.Request
	b := expect.HTTPStep(r.Method, r.Endpoint).WithConnection(r.Connection)
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

func buildGRPCStep(s step) (*expect.StepBuilder, error) {
	r := s.Request
	var body []byte
	if r.Body != nil {
		var err error
		body, err = json.Marshal(r.Body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
	}
	b := expect.GRPCRawCall(r.Connection, r.Endpoint, body)
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
