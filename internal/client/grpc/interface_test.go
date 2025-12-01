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

	client := &GRPCClient{
		cfg: cfg,
	}

	ctx := context.Background()
	authCtx := client.withAuth(ctx)

	md, ok := metadata.FromOutgoingContext(authCtx)
	assert.True(t, ok)
	assert.Contains(t, md["authorization"], "Bearer test-token")
}

func TestWithAuth_NoToken(t *testing.T) {
	cfg := &config.Config{
		Token: "",
	}

	client := &GRPCClient{
		cfg: cfg,
	}

	ctx := context.Background()
	authCtx := client.withAuth(ctx)

	assert.Equal(t, ctx, authCtx)
}

func TestClose(t *testing.T) {
	client := &GRPCClient{
		conn: nil,
	}

	err := client.Close()
	assert.NoError(t, err)
}
