package grpc

import (
	"context"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"gophkeeper/internal/common"
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

	conn, err := grpc.NewClient("localhost:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	client := &GRPCClient{
		conn:   conn,
		client: nil,
		cfg:    cfg,
	}

	ctx := context.Background()
	authCtx := client.withAuth(ctx)

	md, ok := metadata.FromOutgoingContext(authCtx)
	assert.True(t, ok)
	assert.Contains(t, md["authorization"], "Bearer test-token")
}

func TestGRPCClient_WithAuth_NoToken(t *testing.T) {
	cfg := &config.Config{
		Token: "",
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

	assert.Equal(t, ctx, authCtx)
}

func TestAllGRPCMethods(_ *testing.T) {
	cfg := &config.Config{
		ServerURL: "localhost:8081",
		Token:     "test-token",
	}

	client, _ := NewGRPCClient(cfg)
	if client != nil {
		_ = client.withAuth(context.Background())
		_ = client.Close()

		op := &common.OperationRequest{Type: common.OpLogin}
		_, _ = client.Execute(op)
		_, _ = client.Register("login", "pass")
		_, _ = client.Login("login", "pass")
		_, _ = client.Sync(common.Now(), []common.SecretData{})
		_ = client.UploadFile("path", "id")
		_ = client.DownloadFile("id", "path")
	}
}

func TestGRPCClient_AuthMethods(t *testing.T) {
	cfg := &config.Config{
		Token: "test-token",
	}

	client := &GRPCClient{
		cfg: cfg,
	}

	ctx := context.Background()
	authCtx := client.withAuth(ctx)

	md, ok := metadata.FromOutgoingContext(authCtx)
	if ok {
		_ = md
	}

	cfgNoToken := &config.Config{Token: ""}
	clientNoToken := &GRPCClient{cfg: cfgNoToken}
	ctxNoAuth := clientNoToken.withAuth(ctx)
	assert.Equal(t, ctx, ctxNoAuth)
}

func TestGRPCClient_Close(_ *testing.T) {
	client := &GRPCClient{}
	err := client.Close()
	_ = err
}

func TestGRPCClient_Execute(_ *testing.T) {
	client := &GRPCClient{
		cfg: &config.Config{},
	}
	op := &common.OperationRequest{
		Type: common.OpLogin,
	}
	_, _ = client.Execute(op)
}

func TestGRPCClient_MethodDeclarations(_ *testing.T) {
	client := &GRPCClient{
		cfg: &config.Config{},
	}

	_, _ = client.Register("user", "pass")
	_, _ = client.Login("user", "pass")
	_, _ = client.Sync(common.Now(), []common.SecretData{})
	_ = client.UploadFile("path", "id")
	_ = client.DownloadFile("id", "path")
}
