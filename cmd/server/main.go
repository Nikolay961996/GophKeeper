package main

import (
	"flag"
	"log"

	"gophkeeper/internal/server/grpc"
	"gophkeeper/storage"
)

var (
	grpcAddr  = flag.String("grpc-addr", ":8081", "gRPC server address")
	jwtSecret = flag.String("jwt-secret", "default-secret-key-change-in-production", "JWT secret key")
)

func main() {
	flag.Parse()

	storage := storage.NewMemoryStorage()

	grpcServer := grpc.NewGRPCServer(storage, *jwtSecret)
	log.Printf("Starting gRPC server on %s", *grpcAddr)

	if err := grpcServer.Start(*grpcAddr); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
