package main

import (
	"embed"
	"net"
	"testing"

	"github.com/jesse0michael/go-expect/pkg/expect"
)

//go:embed testdata
var testdata embed.FS

func TestSuite(t *testing.T) {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := run()
	go srv.Serve(lis) //nolint:errcheck
	t.Cleanup(srv.Stop)

	suite, err := expect.LoadFS(testdata)
	if err != nil {
		t.Fatalf("load suite: %v", err)
	}

	suite.WithConnections(expect.GRPC("grpc", lis.Addr().String()))

	expect.NewTestSuite(suite).Run(t)
}
