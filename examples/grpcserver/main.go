package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	pb "github.com/jesse0michael/go-expect/examples/grpcserver/proto"
)

func init() {
	// Register a JSON codec so the server accepts JSON-encoded messages,
	// enabling GRPCRawCall from go-expect (and YAML-driven tests).
	encoding.RegisterCodec(jsonCodec{})
}

func main() {
	lis, err := net.Listen("tcp", ":"+port())
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	s := run()
	log.Printf("listening on %s", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func run() *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterCounterServiceServer(s, &counterServer{})
	return s
}

type counterServer struct {
	pb.UnimplementedCounterServiceServer
	count atomic.Int32
}

func (s *counterServer) Increment(_ context.Context, _ *pb.Empty) (*pb.CounterResponse, error) {
	return &pb.CounterResponse{Count: s.count.Add(1)}, nil
}

func (s *counterServer) Decrement(_ context.Context, _ *pb.Empty) (*pb.CounterResponse, error) {
	return &pb.CounterResponse{Count: s.count.Add(-1)}, nil
}

func (s *counterServer) Zero(_ context.Context, _ *pb.Empty) (*pb.CounterResponse, error) {
	s.count.Store(0)
	return &pb.CounterResponse{Count: s.count.Load()}, nil
}

func (s *counterServer) Add(_ context.Context, req *pb.AddRequest) (*pb.CounterResponse, error) {
	return &pb.CounterResponse{Count: s.count.Add(req.N)}, nil
}

func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "50051"
}

// jsonCodec lets the server accept JSON-encoded gRPC messages using protojson.
type jsonCodec struct{}

func (jsonCodec) Name() string { return "json" }

func (jsonCodec) Marshal(v any) ([]byte, error) {
	if m, ok := v.(proto.Message); ok {
		return protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(m)
	}
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v any) error {
	if len(data) == 0 {
		data = []byte("{}")
	}
	if m, ok := v.(proto.Message); ok {
		return protojson.Unmarshal(data, m)
	}
	return json.Unmarshal(data, v)
}
