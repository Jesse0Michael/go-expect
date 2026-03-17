package expect

import (
	"testing"
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
	suite.WithConnections(HTTP("api", "http://example.com"))
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
	conns, err := buildFileConnections(f)
	if err != nil {
		t.Fatalf("buildFileConnections error: %v", err)
	}
	connMap, defaultConn := buildConnMap(conns)
	scenarios, err := buildFileScenarios(f, connMap, defaultConn)
	if err != nil {
		t.Fatalf("buildFileScenarios error: %v", err)
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
	conns, err := buildFileConnections(f)
	if err != nil {
		t.Fatalf("buildFileConnections error: %v", err)
	}
	connMap, defaultConn := buildConnMap(conns)
	_, err = buildFileScenarios(f, connMap, defaultConn)
	if err != nil {
		t.Fatalf("buildFileScenarios error: %v", err)
	}
}

func TestBuildConnections_sqlTypes(t *testing.T) {
	data := []byte(`
connections:
  - name: mydb
    type: postgres
    url: postgres://localhost/testdb
  - name: mydb2
    type: mysql
    url: root@tcp(localhost)/testdb
scenarios: []
`)
	f, err := unmarshalYAML(data)
	if err != nil {
		t.Fatalf("unmarshalYAML error: %v", err)
	}
	conns, err := buildFileConnections(f)
	if err != nil {
		t.Fatalf("buildFileConnections error: %v", err)
	}
	if len(conns) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(conns))
	}
	pg, ok := conns[0].(*SQLConnection)
	if !ok {
		t.Fatalf("expected SQLConnection, got %T", conns[0])
	}
	if pg.Driver != "postgres" {
		t.Fatalf("expected driver postgres, got %s", pg.Driver)
	}
	my, ok := conns[1].(*SQLConnection)
	if !ok {
		t.Fatalf("expected SQLConnection, got %T", conns[1])
	}
	if my.Driver != "mysql" {
		t.Fatalf("expected driver mysql, got %s", my.Driver)
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
	_, err = buildFileConnections(f)
	if err == nil {
		t.Fatal("expected error for unknown connection type, got nil")
	}
}
