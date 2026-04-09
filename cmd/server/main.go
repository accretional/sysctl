//go:build darwin

package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "github.com/accretional/sysctl/proto/sysctlpb"
	"github.com/accretional/sysctl/internal/server"
)

func main() {
	port := flag.Int("port", 50051, "gRPC server port")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", addr, err)
	}

	s := grpc.NewServer()
	pb.RegisterSysctlServiceServer(s, server.New())

	log.Printf("sysctl gRPC server listening on %s", addr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
