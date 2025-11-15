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

// Config представляет конфигурацию сервера
type Config struct {
	Addr      string
	JWTSecret string
}

// Server представляет HTTP сервер
type Server struct {
	httpServer *http.Server
	storage    storage.Storage
	config     *Config
}

// NewServer создает новый экземпляр Server
func NewServer(cfg *Config, storage storage.Storage) *Server {
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)

	// Создаем обработчики
	authHandler := handlers.NewAuthHandler(storage, cfg.JWTSecret)
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

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "ok", "timestamp": "%s"}`, time.Now().Format(time.RFC3339))
	})

	return &Server{
		httpServer: &http.Server{
			Addr:    cfg.Addr,
			Handler: mux,
		},
		storage: storage,
		config:  cfg,
	}
}

// Start запускает сервер
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// GracefulStop останавливает сервер
func (s *Server) GracefulStop(ctx context.Context) error {
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

	log.Printf("Server is running on %s", s.httpServer.Addr)
	log.Printf("Health check available at http://%s/health", s.httpServer.Addr)

	// Ждем сигнал остановки
	<-stop
	log.Println("Shutting down server...")

	// Даем серверу 5 секунд на завершение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.GracefulStop(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
