package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gophkeeper/internal/server/handlers"
	"gophkeeper/internal/server/middleware"
	"gophkeeper/storage"
)

// Server представляет HTTP сервер
type Server struct {
	httpServer *http.Server
	storage    storage.Storage
	jwtSecret  string
}

// NewServer создает новый экземпляр Server
func NewServer(addr string, storage storage.Storage, jwtSecret string) *Server {
	authMiddleware := middleware.NewAuthMiddleware(jwtSecret)

	// Создаем обработчики
	authHandler := handlers.NewAuthHandler(storage, jwtSecret)
	secretsHandler := handlers.NewSecretsHandler(storage)

	// Настраиваем маршруты
	mux := http.NewServeMux()

	// Публичные маршруты
	mux.HandleFunc("/api/register", authHandler.Register)
	mux.HandleFunc("/api/login", authHandler.Login)

	// Защищенные маршруты
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("/api/sync", secretsHandler.Sync)
	protectedMux.HandleFunc("/api/secrets", secretsHandler.GetSecrets)

	// Обертываем защищенные маршруты middleware аутентификации
	mux.Handle("/api/", authMiddleware.Middleware(protectedMux))

	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
		storage:   storage,
		jwtSecret: jwtSecret,
	}
}

// Start запускает сервер
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Stop останавливает сервер
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Run запускает сервер с обработкой сигналов
func (s *Server) Run() {
	// Канал для сигналов OS
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запускаем сервер в горутине
	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Ждем сигнал остановки
	<-stop
	log.Println("Shutting down server...")

	// Даем серверу 5 секунд на завершение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Stop(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
