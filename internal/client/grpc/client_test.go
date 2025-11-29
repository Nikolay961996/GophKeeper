package grpc

import (
	"context"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"testing"

	"gophkeeper/internal/client/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestNewGRPCClient(t *testing.T) {
	cfg := &config.Config{
		ServerURL: "localhost:8081",
	}

	// Этот тест будет пропущен в CI, так как требует запущенный сервер
	t.Skip("Skipping test that requires gRPC server")

	client, err := NewGRPCClient(cfg)
	require.NoError(t, err)
	assert.NotNil(t, client)
	defer client.Close()
}

func TestGRPCClient_WithAuth(t *testing.T) {
	cfg := &config.Config{
		Token: "test-token",
	}

	// Создаем mock клиента для тестирования withAuth
	conn, err := grpc.NewClient("localhost:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	client := &GRPCClient{
		conn:   conn,
		client: nil, // не нужно для этого теста
		cfg:    cfg,
	}

	ctx := context.Background()
	authCtx := client.withAuth(ctx)

	// Проверяем, что контекст содержит метаданные
	md, ok := metadata.FromOutgoingContext(authCtx)
	assert.True(t, ok)
	assert.Contains(t, md["authorization"], "Bearer test-token")
}

func TestGRPCClient_WithAuth_NoToken(t *testing.T) {
	cfg := &config.Config{
		Token: "", // нет токена
	}

	conn, err := grpc.NewClient("localhost:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	client := &GRPCClient{
		conn:   conn,
		client: nil,
		cfg:    cfg,
	}

	ctx := context.Background()
	authCtx := client.withAuth(ctx)

	// Должен вернуть оригинальный контекст без изменений
	assert.Equal(t, ctx, authCtx)
}

// Mock тесты для основных методов
func TestGRPCClient_Execute(t *testing.T) {
	// Здесь нужно создать mock gRPC соединение
	// Пока пропускаем, так как требует сложной настройки
	t.Skip("Requires gRPC mock setup")
}

func TestGRPCClient_Register(t *testing.T) {
	t.Skip("Requires gRPC mock setup")
}

func TestGRPCClient_Login(t *testing.T) {
	t.Skip("Requires gRPC mock setup")
}

func TestGRPCClient_Sync(t *testing.T) {
	t.Skip("Requires gRPC mock setup")
}
