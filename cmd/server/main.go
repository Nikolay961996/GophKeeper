package main

import (
	"flag"
	"gophkeeper/internal/server"
	"gophkeeper/storage"
)

var (
	addr      = flag.String("addr", ":8080", "Server address")
	jwtSecret = flag.String("jwt-secret", "default-secret-key-change-in-production", "JWT secret key")
)

func main() {
	flag.Parse()

	// Инициализируем хранилище
	storage := storage.NewMemoryStorage()

	// Конфигурация сервера
	cfg := &server.Config{
		Addr:      *addr,
		JWTSecret: *jwtSecret,
	}

	// Создаем и запускаем сервер
	srv := server.NewServer(cfg, storage)
	srv.Run()
}
