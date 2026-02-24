package expect

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GRPCConnection is a connection to a gRPC service.
type GRPCConnection struct {
	Name string
	Addr string
	opts []grpc.DialOption
	conn *grpc.ClientConn
}

// GRPC creates a GRPCConnection. opts are appended after the default insecure credential.
// Use grpc.WithTransportCredentials(...) in opts to enable TLS.
func GRPC(name, addr string, opts ...grpc.DialOption) *GRPCConnection {
	return &GRPCConnection{Name: name, Addr: addr, opts: opts}
}

func (c *GRPCConnection) Type() string    { return "grpc" }
func (c *GRPCConnection) GetName() string { return c.Name }

// Dial opens the underlying gRPC client connection (lazy — called on first use).
func (c *GRPCConnection) Dial() error {
	if c.conn != nil {
		return nil
	}
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	opts = append(opts, c.opts...)
	conn, err := grpc.NewClient(c.Addr, opts...)
	if err != nil {
		return fmt.Errorf("grpc dial %q: %w", c.Addr, err)
	}
	c.conn = conn
	return nil
}

// ClientConn returns the raw *grpc.ClientConn, dialling if necessary.
func (c *GRPCConnection) ClientConn() (*grpc.ClientConn, error) {
	if err := c.Dial(); err != nil {
		return nil, err
	}
	return c.conn, nil
}

// Close tears down the gRPC connection.
func (c *GRPCConnection) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
