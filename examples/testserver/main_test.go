package main

import (
	"net/http/httptest"
	"testing"

	"github.com/jesse0michael/go-expect/pkg/expect"
	"github.com/jesse0michael/go-expect/pkg/expect/loader"
)

// TestYAMLSuite loads testdata/expect.yaml and runs it against an in-process server.
func TestYAMLSuite(t *testing.T) {
	srv := httptest.NewServer(run().Handler)
	t.Cleanup(srv.Close)

	suite, err := loader.LoadFile("testdata/expect.yaml")
	if err != nil {
		t.Fatalf("load yaml: %v", err)
	}

	// Override the connection URL to point at the in-process test server.
	suite.WithConnections(expect.HTTP("http", srv.URL))

	expect.NewTestSuite(suite).Run(t)
}

// TestGoSuite demonstrates the fluent Go API against an in-process server.
func TestGoSuite(t *testing.T) {
	srv := httptest.NewServer(run().Handler)
	t.Cleanup(srv.Close)

	suite := expect.NewSuite().
		WithConnections(expect.HTTP("api", srv.URL)).
		WithScenarios(
			expect.NewScenario("counter flow").
				AddStep(expect.POST("/increment").ExpectStatus(200).ExpectBody(map[string]any{"count": float64(1)})).
				AddStep(expect.POST("/increment").ExpectStatus(200).ExpectBody(map[string]any{"count": float64(2)})).
				AddStep(expect.POST("/decrement").ExpectStatus(200).ExpectBody(map[string]any{"count": float64(1)})).
				AddStep(expect.POST("/zero").ExpectStatus(200).ExpectBody(map[string]any{"count": float64(0)})),
		)

	expect.NewTestSuite(suite).Run(t)
}
