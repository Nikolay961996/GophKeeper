package main

import (
	"flag"
	"gophkeeper/internal/server"
	"gophkeeper/storage"
)

var (
	addr      = flag.String("addr", ":8080", "Server address")
	jwtSecret = flag.String("jwt-secret", "default-secret-key", "JWT secret key")
)

func main() {
	flag.Parse()

	// Инициализируем хранилище
	storage := storage.NewMemoryStorage()

	// Создаем и запускаем сервер
	srv := server.NewServer(*addr, storage, *jwtSecret)
	srv.Run()
}
