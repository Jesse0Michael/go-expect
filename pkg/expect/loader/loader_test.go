package loader

import (
	"testing"

	expect "github.com/jesse0michael/go-expect/pkg/expect"
)

func TestUnmarshalYAML_connections(t *testing.T) {
	data := []byte(`
connections:
  - name: api
    type: http
    url: http://localhost:8080
scenarios: []
`)
	_, err := unmarshalYAML(data)
	if err != nil {
		t.Fatalf("unmarshalYAML error: %v", err)
	}
}

func TestLoadYAML_connectionOverride(t *testing.T) {
	data := []byte(`
connections:
  - name: api
    type: http
    url: http://localhost:8080
scenarios: []
`)
	suite, err := LoadYAML(data)
	if err != nil {
		t.Fatalf("LoadYAML error: %v", err)
	}
	// Connections from file are in the suite; caller can override by name.
	suite.WithConnections(expect.HTTP("api", "http://example.com"))
}

func TestBuildScenarios_steps(t *testing.T) {
	data := []byte(`
connections:
  - name: api
    type: http
    url: http://localhost:8080

scenarios:
  - name: counter test
    steps:
      - request:
          connection: api
          method: POST
          endpoint: /increment
        expect:
          status: 200
          body:
            count: 1
      - request:
          connection: api
          method: POST
          endpoint: /increment
        expect:
          status: 200
          body:
            count: 2
`)
	f, err := unmarshalYAML(data)
	if err != nil {
		t.Fatalf("unmarshalYAML error: %v", err)
	}
	conns, err := buildConnections(f)
	if err != nil {
		t.Fatalf("buildConnections error: %v", err)
	}
	connMap, defaultConn := connectionMap(conns)
	scenarios, err := buildScenarios(f, connMap, defaultConn)
	if err != nil {
		t.Fatalf("buildScenarios error: %v", err)
	}
	if len(scenarios) != 1 {
		t.Fatalf("expected 1 scenario, got %d", len(scenarios))
	}
}

func TestBuildScenarios_save(t *testing.T) {
	data := []byte(`
connections:
  - name: api
    type: http
    url: http://localhost:8080

scenarios:
  - name: save test
    steps:
      - request:
          connection: api
          method: POST
          endpoint: /users
        expect:
          status: 201
          save:
            - field: id
              as: user_id
`)
	f, err := unmarshalYAML(data)
	if err != nil {
		t.Fatalf("unmarshalYAML error: %v", err)
	}
	conns, err := buildConnections(f)
	if err != nil {
		t.Fatalf("buildConnections error: %v", err)
	}
	connMap, defaultConn := connectionMap(conns)
	_, err = buildScenarios(f, connMap, defaultConn)
	if err != nil {
		t.Fatalf("buildScenarios error: %v", err)
	}
}

func TestBuildConnections_unknownType(t *testing.T) {
	data := []byte(`
connections:
  - name: bad
    type: websocket
    url: ws://localhost:9090
scenarios: []
`)
	f, err := unmarshalYAML(data)
	if err != nil {
		t.Fatalf("unmarshalYAML error: %v", err)
	}
	_, err = buildConnections(f)
	if err == nil {
		t.Fatal("expected error for unknown connection type, got nil")
	}
}
