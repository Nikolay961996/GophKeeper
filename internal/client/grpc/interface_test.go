package grpc

import (
	"context"
	"testing"

	"gophkeeper/internal/client/config"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestWithAuth(t *testing.T) {
	cfg := &config.Config{
		Token: "test-token",
	}

	// Создаем mock клиента
	client := &GRPCClient{
		cfg: cfg,
	}

	ctx := context.Background()
	authCtx := client.withAuth(ctx)

	// Проверяем что контекст содержит метаданные
	md, ok := metadata.FromOutgoingContext(authCtx)
	assert.True(t, ok)
	assert.Contains(t, md["authorization"], "Bearer test-token")
}

func TestWithAuth_NoToken(t *testing.T) {
	cfg := &config.Config{
		Token: "", // нет токена
	}

	client := &GRPCClient{
		cfg: cfg,
	}

	ctx := context.Background()
	authCtx := client.withAuth(ctx)

	// Должен вернуть тот же контекст
	assert.Equal(t, ctx, authCtx)
}

func TestClose(t *testing.T) {
	client := &GRPCClient{
		conn: nil, // nil connection
	}

	// Close не должен паниковать даже с nil connection
	err := client.Close()
	assert.NoError(t, err)
}
