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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestGRPCServer_Authentication(t *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	userID := uuid.New()
	token, err := common.GenerateJWTToken(userID, "testuser", "test-secret", time.Hour)
	require.NoError(t, err)

	ctx := context.Background()
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewIncomingContext(ctx, md) // Изменено на IncomingContext

	authenticatedUserID, err := server.authenticate(ctx)
	require.NoError(t, err)
	assert.Equal(t, userID, authenticatedUserID)
}

func TestGRPCServer_Authentication_InvalidToken(t *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	ctx := context.Background()
	md := metadata.New(map[string]string{"authorization": "Bearer invalid-token"})
	ctx = metadata.NewIncomingContext(ctx, md) // Изменено на IncomingContext

	_, err := server.authenticate(ctx)
	assert.Error(t, err)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, grpcStatus.Code())
}

func TestGRPCServer_Authentication_MissingToken(t *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	ctx := context.Background()
	md := metadata.New(map[string]string{})
	ctx = metadata.NewIncomingContext(ctx, md)

	_, err := server.authenticate(ctx)
	assert.Error(t, err)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, grpcStatus.Code())
}

func TestGRPCServer_Authentication_MissingMetadata(t *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	ctx := context.Background()

	_, err := server.authenticate(ctx)
	assert.Error(t, err)

	grpcStatus, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, grpcStatus.Code())
}

func TestGRPCServer_Authentication_BearerPrefix(t *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	userID := uuid.New()
	token, err := common.GenerateJWTToken(userID, "testuser", "test-secret", time.Hour)
	require.NoError(t, err)

	ctx := context.Background()
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx = metadata.NewIncomingContext(ctx, md)

	authenticatedUserID, err := server.authenticate(ctx)
	require.NoError(t, err)
	assert.Equal(t, userID, authenticatedUserID)

	ctx2 := context.Background()
	md2 := metadata.New(map[string]string{"authorization": token})
	ctx2 = metadata.NewIncomingContext(ctx2, md2)

	authenticatedUserID2, err := server.authenticate(ctx2)
	require.NoError(t, err)
	assert.Equal(t, userID, authenticatedUserID2)
}

// интеграционный тест с реальным gRPC сервером
func TestGRPCServer_RegisterAndLogin_Integration(t *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	api.RegisterGophKeeperServer(grpcServer, server)

	go func() {
		grpcServer.Serve(lis)
	}()
	defer grpcServer.GracefulStop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := api.NewGophKeeperClient(conn)

	registerOp := common.RegisterOp{
		Login:    "testuser",
		Password: "testpassword",
	}

	registerReq, err := common.CreateOperationRequest(common.OpRegister, registerOp)
	require.NoError(t, err)

	registerPayload, err := common.MarshalOperation(registerReq)
	require.NoError(t, err)

	registerResp, err := client.Execute(context.Background(), &api.CommandRequest{
		Payload: registerPayload,
	})
	require.NoError(t, err)
	assert.True(t, registerResp.Success)
	assert.Empty(t, registerResp.Error)

	loginOp := common.LoginOp{
		Login:    "testuser",
		Password: "testpassword",
	}

	loginReq, err := common.CreateOperationRequest(common.OpLogin, loginOp)
	require.NoError(t, err)

	loginPayload, err := common.MarshalOperation(loginReq)
	require.NoError(t, err)

	loginResp, err := client.Execute(context.Background(), &api.CommandRequest{
		Payload: loginPayload,
	})
	require.NoError(t, err)
	assert.True(t, loginResp.Success)
	assert.Empty(t, loginResp.Error)

	wrongLoginOp := common.LoginOp{
		Login:    "testuser",
		Password: "wrongpassword",
	}

	wrongLoginReq, err := common.CreateOperationRequest(common.OpLogin, wrongLoginOp)
	require.NoError(t, err)

	wrongLoginPayload, err := common.MarshalOperation(wrongLoginReq)
	require.NoError(t, err)

	wrongLoginResp, err := client.Execute(context.Background(), &api.CommandRequest{
		Payload: wrongLoginPayload,
	})
	require.NoError(t, err)
	assert.False(t, wrongLoginResp.Success)
	assert.NotEmpty(t, wrongLoginResp.Error)
}

// Тест для обработки регистрации существующего пользователя
func TestGRPCServer_RegisterExistingUser(t *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	user := &common.User{
		ID:           uuid.New(),
		Login:        "existinguser",
		PasswordHash: "hashedpassword",
		CreatedAt:    common.Now(),
	}
	err := storage.CreateUser(user)
	require.NoError(t, err)

	registerOp := common.RegisterOp{
		Login:    "existinguser",
		Password: "password",
	}

	registerReq, err := common.CreateOperationRequest(common.OpRegister, registerOp)
	require.NoError(t, err)

	registerPayload, err := common.MarshalOperation(registerReq)
	require.NoError(t, err)

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	api.RegisterGophKeeperServer(grpcServer, server)

	go func() {
		grpcServer.Serve(lis)
	}()
	defer grpcServer.GracefulStop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := api.NewGophKeeperClient(conn)

	registerResp, err := client.Execute(context.Background(), &api.CommandRequest{
		Payload: registerPayload,
	})
	require.NoError(t, err)
	assert.False(t, registerResp.Success)
	assert.Contains(t, registerResp.Error, "already exists")
}

// Тест для проверки обработки неверного формата операции
func TestGRPCServer_InvalidOperation(t *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	api.RegisterGophKeeperServer(grpcServer, server)

	go func() {
		grpcServer.Serve(lis)
	}()
	defer grpcServer.GracefulStop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := api.NewGophKeeperClient(conn)

	invalidResp, err := client.Execute(context.Background(), &api.CommandRequest{
		Payload: []byte("invalid json data"),
	})
	require.NoError(t, err)
	assert.False(t, invalidResp.Success)
	assert.NotEmpty(t, invalidResp.Error)

	unknownOp := common.OperationRequest{
		Type: "unknown_operation",
	}
	unknownPayload, err := common.MarshalOperation(unknownOp)
	require.NoError(t, err)

	unknownResp, err := client.Execute(context.Background(), &api.CommandRequest{
		Payload: unknownPayload,
	})
	require.NoError(t, err)
	assert.False(t, unknownResp.Success)
	assert.Contains(t, unknownResp.Error, "unknown operation")
}

func TestAllServerMethods(_ *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	req := &api.CommandRequest{
		Payload: []byte("invalid"),
	}
	_, _ = server.Execute(context.Background(), req)
	_, _ = server.handleRegister([]byte("invalid"))
	_, _ = server.handleLogin([]byte("invalid"))
	userID := uuid.New()
	_, _ = server.handleSync(userID, []byte("invalid"))
	ctx := context.Background()
	_, _ = server.authenticate(ctx)
	ctxWithMD := metadata.NewIncomingContext(ctx, metadata.New(map[string]string{}))
	_, _ = server.authenticate(ctxWithMD)
	ctxWithToken := metadata.NewIncomingContext(ctx, metadata.New(map[string]string{
		"authorization": "Bearer invalid",
	}))
	_, _ = server.authenticate(ctxWithToken)
	_, _ = server.getSecretByID(userID, uuid.New())
	_, _ = server.createErrorResponse(codes.Internal, "error")
}

func TestGRPCServer_Creation(t *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")
	assert.NotNil(t, server)
}

func TestGRPCServer_ErrorHandling(t *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	resp, err := server.createErrorResponse(codes.Internal, "test error")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "test error", resp.Error)
}

func TestGRPCServer_Authentication_EdgeCases(t *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	ctx := context.Background()
	_, err := server.authenticate(ctx)
	assert.Error(t, err)

	md := metadata.New(map[string]string{})
	ctxWithEmptyMD := metadata.NewIncomingContext(ctx, md)
	_, err = server.authenticate(ctxWithEmptyMD)
	assert.Error(t, err)

	mdWithInvalidToken := metadata.New(map[string]string{
		"authorization": "Bearer invalid-token",
	})
	ctxWithInvalidToken := metadata.NewIncomingContext(ctx, mdWithInvalidToken)
	_, err = server.authenticate(ctxWithInvalidToken)
	assert.Error(t, err)
}

func TestGRPCServer_GetSecretByID_NotFound(t *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	userID := uuid.New()
	secretID := uuid.New()

	secret, err := server.getSecretByID(userID, secretID)
	assert.Error(t, err)
	assert.Nil(t, secret)
}

func TestGRPCServer_HandleInvalidOperations(_ *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	_, _ = server.handleRegister([]byte("invalid json"))
	_, _ = server.handleLogin([]byte("invalid json"))
	userID := uuid.New()
	_, _ = server.handleSync(userID, []byte("invalid json"))
}

func TestWrappedStream(_ *testing.T) {
	stream := &wrappedStream{}
	ctx := stream.Context()
	_ = ctx
}
