package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

	// Создаем контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Канал для ожидания завершения серверов
	var wg sync.WaitGroup

	// Запускаем HTTP сервер
	wg.Add(1)
	go func() {
		defer wg.Done()

		cfg := &server.Config{
			Addr:      *httpAddr,
			JWTSecret: *jwtSecret,
		}

		srv := server.NewServer(cfg, storage)
		log.Printf("Starting HTTP server on %s", *httpAddr)

		// Запускаем сервер в отдельной горутине
		go func() {
			if err := srv.Start(); err != nil {
				log.Printf("HTTP server error: %v", err)
			}
		}()

		// Ждем сигнала остановки
		<-ctx.Done()
		log.Println("Shutting down HTTP server...")

		// Останавливаем HTTP сервер
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := srv.GracefulStop(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		} else {
			log.Println("HTTP server stopped gracefully")
		}
	}()

	// Запускаем gRPC сервер
	wg.Add(1)
	go func() {
		defer wg.Done()

		grpcServer := grpc.NewGRPCServer(storage, *jwtSecret)
		log.Printf("Starting gRPC server on %s", *grpcAddr)

		// Запускаем gRPC сервер в отдельной горутине
		go func() {
			if err := grpcServer.Start(*grpcAddr); err != nil {
				log.Printf("gRPC server error: %v", err)
			}
		}()

		// Ждем сигнала остановки
		<-ctx.Done()
		log.Println("Shutting down gRPC server...")

		// Останавливаем gRPC сервер
		grpcServer.Stop()
		log.Println("gRPC server stopped gracefully")
	}()

	// Ждем сигналов завершения
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	log.Printf("Both servers are running. Press Ctrl+C to stop.")
	log.Printf("HTTP server: http://localhost%s", *httpAddr)
	log.Printf("gRPC server: localhost%s", *grpcAddr)

	// Блокируемся до получения сигнала
	sig := <-signalCh
	log.Printf("Received signal: %v", sig)

	// Инициируем graceful shutdown
	log.Println("Initiating graceful shutdown...")
	cancel()

	// Ждем завершения всех серверов
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Таймаут на завершение
	select {
	case <-done:
		log.Println("All servers stopped gracefully")
	case <-time.After(10 * time.Second):
		log.Println("Timeout exceeded, forcing shutdown")
	}
}
