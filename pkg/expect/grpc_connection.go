package expect

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpc_reflection_v1 "google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// GRPCConnection is a connection to a gRPC service.
type GRPCConnection struct {
	Name string
	Addr string
	opts []grpc.DialOption
	conn *grpc.ClientConn

	mu      sync.Mutex
	methods map[string]protoreflect.MethodDescriptor
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

// resolveMethod uses gRPC server reflection to look up the MethodDescriptor for fullMethod.
// Results are cached for the lifetime of the connection.
func (c *GRPCConnection) resolveMethod(ctx context.Context, fullMethod string) (protoreflect.MethodDescriptor, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.methods == nil {
		c.methods = make(map[string]protoreflect.MethodDescriptor)
	}
	if md, ok := c.methods[fullMethod]; ok {
		return md, nil
	}

	// fullMethod is "/<package.Service>/<Method>"
	trimmed := strings.TrimPrefix(fullMethod, "/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid full method %q", fullMethod)
	}
	serviceSymbol, methodName := parts[0], parts[1]

	refClient := grpc_reflection_v1.NewServerReflectionClient(c.conn)
	stream, err := refClient.ServerReflectionInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("reflection stream: %w", err)
	}
	defer stream.CloseSend() //nolint:errcheck

	if err := stream.Send(&grpc_reflection_v1.ServerReflectionRequest{
		MessageRequest: &grpc_reflection_v1.ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: serviceSymbol,
		},
	}); err != nil {
		return nil, fmt.Errorf("reflection send: %w", err)
	}

	resp, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("reflection recv: %w", err)
	}

	fdResp, ok := resp.MessageResponse.(*grpc_reflection_v1.ServerReflectionResponse_FileDescriptorResponse)
	if !ok {
		if errMsg, ok := resp.MessageResponse.(*grpc_reflection_v1.ServerReflectionResponse_ErrorResponse); ok {
			return nil, fmt.Errorf("reflection error: %s", errMsg.ErrorResponse.ErrorMessage)
		}
		return nil, fmt.Errorf("unexpected reflection response type")
	}

	// Parse all returned FileDescriptorProtos (includes transitive dependencies).
	// Deduplicate by name in case the server returns duplicates.
	fdSet := &descriptorpb.FileDescriptorSet{}
	seen := make(map[string]bool)
	for _, b := range fdResp.FileDescriptorResponse.FileDescriptorProto {
		fdp := &descriptorpb.FileDescriptorProto{}
		if err := proto.Unmarshal(b, fdp); err != nil {
			return nil, fmt.Errorf("unmarshal file descriptor: %w", err)
		}
		if !seen[fdp.GetName()] {
			seen[fdp.GetName()] = true
			fdSet.File = append(fdSet.File, fdp)
		}
	}

	files, err := protodesc.NewFiles(fdSet)
	if err != nil {
		return nil, fmt.Errorf("build file descriptors: %w", err)
	}

	desc, err := files.FindDescriptorByName(protoreflect.FullName(serviceSymbol))
	if err != nil {
		return nil, fmt.Errorf("find service %q: %w", serviceSymbol, err)
	}
	svcDesc, ok := desc.(protoreflect.ServiceDescriptor)
	if !ok {
		return nil, fmt.Errorf("%q is not a service descriptor", serviceSymbol)
	}

	methodDesc := svcDesc.Methods().ByName(protoreflect.Name(methodName))
	if methodDesc == nil {
		return nil, fmt.Errorf("method %q not found in service %q", methodName, serviceSymbol)
	}

	c.methods[fullMethod] = methodDesc
	return methodDesc, nil
}
