package commands

import (
	"context"
	"testing"
	"time"

	"gophkeeper/internal/client/config"
	"gophkeeper/internal/common"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// MockGRPCClient реализует интерфейс grpc.GRPCClient для тестирования
type MockGRPCClient struct {
	RegisterFunc     func(login, password string) (*common.AuthResult, error)
	LoginFunc        func(login, password string) (*common.AuthResult, error)
	SyncFunc         func(lastSync time.Time, data []common.SecretData) (*common.SyncResult, error)
	ExecuteFunc      func(operation *common.OperationRequest) (*common.OperationResponse, error)
	UploadFileFunc   func(filePath, fileID string) error
	DownloadFileFunc func(fileID, outputPath string) error
	CloseFunc        func() error
}

func (m *MockGRPCClient) Register(login, password string) (*common.AuthResult, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(login, password)
	}
	return nil, nil
}

func (m *MockGRPCClient) Login(login, password string) (*common.AuthResult, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(login, password)
	}
	return nil, nil
}

func (m *MockGRPCClient) Sync(lastSync time.Time, data []common.SecretData) (*common.SyncResult, error) {
	if m.SyncFunc != nil {
		return m.SyncFunc(lastSync, data)
	}
	return nil, nil
}

func (m *MockGRPCClient) Execute(operation *common.OperationRequest) (*common.OperationResponse, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(operation)
	}
	return nil, nil
}

func (m *MockGRPCClient) UploadFile(filePath, fileID string) error {
	if m.UploadFileFunc != nil {
		return m.UploadFileFunc(filePath, fileID)
	}
	return nil
}

func (m *MockGRPCClient) DownloadFile(fileID, outputPath string) error {
	if m.DownloadFileFunc != nil {
		return m.DownloadFileFunc(fileID, outputPath)
	}
	return nil
}

func (m *MockGRPCClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockGRPCClient) withAuth(ctx context.Context) context.Context {
	return ctx
}

func (m *MockGRPCClient) GetConn() *grpc.ClientConn {
	return nil
}

func TestAuthCommands_Register(t *testing.T) {
	cfg := &config.Config{}

	expectedUser := &common.User{
		ID:    uuid.New(),
		Login: "testuser",
	}

	mockClient := &MockGRPCClient{
		RegisterFunc: func(login, password string) (*common.AuthResult, error) {
			assert.Equal(t, "testuser", login)
			assert.Equal(t, "testpassword", password)
			return &common.AuthResult{
				Token: "test-token",
				User:  *expectedUser,
			}, nil
		},
	}

	authCommands := NewAuthCommands(cfg, mockClient)

	err := authCommands.Register("testuser", "testpassword")
	require.NoError(t, err)

	assert.Equal(t, "test-token", cfg.Token)
	assert.Equal(t, expectedUser.ID.String(), cfg.UserID)
}

func TestAuthCommands_Login(t *testing.T) {
	cfg := &config.Config{}

	expectedUser := &common.User{
		ID:    uuid.New(),
		Login: "testuser",
	}

	mockClient := &MockGRPCClient{
		LoginFunc: func(login, password string) (*common.AuthResult, error) {
			assert.Equal(t, "testuser", login)
			assert.Equal(t, "testpassword", password)
			return &common.AuthResult{
				Token: "test-token",
				User:  *expectedUser,
			}, nil
		},
	}

	authCommands := NewAuthCommands(cfg, mockClient)

	err := authCommands.Login("testuser", "testpassword")
	require.NoError(t, err)

	assert.Equal(t, "test-token", cfg.Token)
	assert.Equal(t, expectedUser.ID.String(), cfg.UserID)
}

func TestAuthCommands_Register_Error(t *testing.T) {
	cfg := &config.Config{}

	mockClient := &MockGRPCClient{
		RegisterFunc: func(login, password string) (*common.AuthResult, error) {
			return nil, assert.AnError
		},
	}

	authCommands := NewAuthCommands(cfg, mockClient)

	err := authCommands.Register("testuser", "testpassword")
	assert.Error(t, err)
	assert.Equal(t, "", cfg.Token)
	assert.Equal(t, "", cfg.UserID)
}

func TestAuthCommands_Login_Error(t *testing.T) {
	cfg := &config.Config{}

	mockClient := &MockGRPCClient{
		LoginFunc: func(login, password string) (*common.AuthResult, error) {
			return nil, assert.AnError
		},
	}

	authCommands := NewAuthCommands(cfg, mockClient)

	err := authCommands.Login("testuser", "testpassword")
	assert.Error(t, err)
	assert.Equal(t, "", cfg.Token)
	assert.Equal(t, "", cfg.UserID)
}
