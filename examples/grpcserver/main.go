package main

import (
	"context"
	"log"
	"net"
	"os"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/jesse0michael/go-expect/examples/grpcserver/proto"
)

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
	reflection.Register(s)
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
