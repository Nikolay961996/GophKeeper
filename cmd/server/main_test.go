package main

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainFunction(t *testing.T) {
	// Сохраняем оригинальные флаги и аргументы
	oldArgs := os.Args
	oldGrpcAddr := *grpcAddr
	oldJwtSecret := *jwtSecret
	oldDbConnStr := *dbConnStr
	oldUsePostgres := *usePostgres

	defer func() {
		os.Args = oldArgs
		*grpcAddr = oldGrpcAddr
		*jwtSecret = oldJwtSecret
		*dbConnStr = oldDbConnStr
		*usePostgres = oldUsePostgres
	}()

	// Тестируем парсинг флагов
	os.Args = []string{"test", "-grpc-addr", ":9090", "-jwt-secret", "test-secret"}
	flag.Parse()

	assert.Equal(t, ":9090", *grpcAddr)
	assert.Equal(t, "test-secret", *jwtSecret)
}

func TestEnvironmentVariables(t *testing.T) {
	// Тестируем чтение переменных окружения
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb")

	// Сбрасываем флаги
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// Инициализируем флаги
	grpcAddr := flag.String("grpc-addr", ":8081", "gRPC server address")
	jwtSecret := flag.String("jwt-secret", "default-secret", "JWT secret key")
	dbConnStr := flag.String("db-conn-str", "postgres://test:test@localhost:5432/testdb", "PostgreSQL connection string")
	usePostgres := flag.Bool("use-postgres", false, "Use PostgreSQL storage")

	// Парсим пустые аргументы
	flag.Parse()

	// Переменная окружения должна переопределить значение по умолчанию
	assert.Equal(t, "postgres://test:test@localhost:5432/testdb", *dbConnStr)
	assert.Equal(t, ":8081", *grpcAddr)
	assert.Equal(t, "default-secret", *jwtSecret)
	assert.False(t, *usePostgres)
}

func TestFlagDefaults(t *testing.T) {
	// Сбрасываем флаги
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// Инициализируем заново
	grpcAddr := flag.String("grpc-addr", ":8081", "gRPC server address")
	jwtSecret := flag.String("jwt-secret", "default-secret-key-change-in-production", "JWT secret key")
	dbConnStr := flag.String("db-conn-str", "postgres://user:password@localhost:5433/gophkeeper?sslmode=disable", "PostgreSQL connection string")
	usePostgres := flag.Bool("use-postgres", false, "Use PostgreSQL storage")

	flag.Parse()

	// Проверяем значения по умолчанию
	assert.Equal(t, ":8081", *grpcAddr)
	assert.Equal(t, "default-secret-key-change-in-production", *jwtSecret)
	assert.Equal(t, "postgres://user:password@localhost:5433/gophkeeper?sslmode=disable", *dbConnStr)
	assert.False(t, *usePostgres)
}
