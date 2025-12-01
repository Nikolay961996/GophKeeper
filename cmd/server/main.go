package main

import (
	"flag"
	"log"
	"os"

	"gophkeeper/internal/server/grpc"
	"gophkeeper/storage"
)

var (
	grpcAddr    = flag.String("grpc-addr", ":8081", "gRPC server address")
	jwtSecret   = flag.String("jwt-secret", "default-secret-key-change-in-production", "JWT secret key")
	dbConnStr   = flag.String("db-conn-str", "postgres://user:password@localhost:5433/gophkeeper?sslmode=disable", "PostgreSQL connection string")
	usePostgres = flag.Bool("use-postgres", false, "Use PostgreSQL storage instead of in-memory")
)

func main() {
	flag.Parse()

	if connStr := os.Getenv("DATABASE_URL"); connStr != "" {
		*dbConnStr = connStr
	}

	var s storage.Storage
	var err error

	if *usePostgres {
		log.Printf("Using PostgreSQL storage: %s", *dbConnStr)
		s, err = storage.NewPostgresStorage(*dbConnStr)
		if err != nil {
			log.Fatalf("Failed to create PostgreSQL storage: %v", err)
		}
		defer s.(*storage.PostgresStorage).Close()
		log.Println("File storage support: ENABLED")
	} else {
		log.Println("Using in-memory storage")
		s = storage.NewMemoryStorage()
		log.Println("File storage support: DISABLED (use --use-postgres for file support)")
	}

	grpcServer := grpc.NewGRPCServer(s, *jwtSecret)
	log.Printf("Starting gRPC server on %s", *grpcAddr)

	if err := grpcServer.Start(*grpcAddr); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
