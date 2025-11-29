package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"gophkeeper/api"
	"gophkeeper/internal/common"
	"gophkeeper/storage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func TestGRPCServer_Sync_WithAuth(t *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	user := &common.User{
		ID:           uuid.New(),
		Login:        "testuser",
		PasswordHash: "hashedpassword",
		CreatedAt:    common.Now(),
	}
	err := storage.CreateUser(user)
	require.NoError(t, err)

	token, err := common.GenerateJWTToken(user.ID, user.Login, "test-secret", time.Hour)
	require.NoError(t, err)

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(server.unaryAuthInterceptor),
	)
	api.RegisterGophKeeperServer(grpcServer, server)

	go func() {
		grpcServer.Serve(lis)
	}()
	defer grpcServer.GracefulStop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := api.NewGophKeeperClient(conn)

	syncOp := common.SyncOp{
		LastSync: common.Now().Add(-time.Hour),
		Data:     []common.SecretData{},
	}

	syncReq, err := common.CreateOperationRequest(common.OpSync, syncOp)
	require.NoError(t, err)

	syncPayload, err := common.MarshalOperation(syncReq)
	require.NoError(t, err)

	ctx := context.Background()
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	syncResp, err := client.Execute(ctx, &api.CommandRequest{
		Payload: syncPayload,
	})
	require.NoError(t, err)
	assert.True(t, syncResp.Success)
}

func TestGRPCServer_Sync_WithoutAuth(t *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(server.unaryAuthInterceptor),
	)
	api.RegisterGophKeeperServer(grpcServer, server)

	go func() {
		grpcServer.Serve(lis)
	}()
	defer grpcServer.GracefulStop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := api.NewGophKeeperClient(conn)

	syncOp := common.SyncOp{
		LastSync: common.Now().Add(-time.Hour),
		Data:     []common.SecretData{},
	}

	syncReq, err := common.CreateOperationRequest(common.OpSync, syncOp)
	require.NoError(t, err)

	syncPayload, err := common.MarshalOperation(syncReq)
	require.NoError(t, err)

	syncResp, err := client.Execute(context.Background(), &api.CommandRequest{
		Payload: syncPayload,
	})
	require.NoError(t, err)
	assert.False(t, syncResp.Success)
	assert.Contains(t, syncResp.Error, "authentication required")
}
