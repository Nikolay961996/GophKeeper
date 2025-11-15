package main

import (
	"flag"
	"log"

	"gophkeeper/internal/server"
	"gophkeeper/internal/server/grpc"
	"gophkeeper/storage"
)

var (
	httpAddr  = flag.String("http-addr", ":8080", "HTTP server address")
	grpcAddr  = flag.String("grpc-addr", ":8081", "gRPC server address")
	jwtSecret = flag.String("jwt-secret", "default-secret-key-change-in-production", "JWT secret key")
)

func main() {
	flag.Parse()

	// Инициализируем хранилище
	storage := storage.NewMemoryStorage()

	// Запускаем HTTP сервер
	go func() {
		cfg := &server.Config{
			Addr:      *httpAddr,
			JWTSecret: *jwtSecret,
		}

		srv := server.NewServer(cfg, storage)
		log.Printf("Starting HTTP server on %s", *httpAddr)
		srv.Run()
	}()

	// Запускаем gRPC сервер
	grpcServer := grpc.NewGRPCServer(storage, *jwtSecret)
	log.Printf("Starting gRPC server on %s", *grpcAddr)
	if err := grpcServer.Start(*grpcAddr); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
