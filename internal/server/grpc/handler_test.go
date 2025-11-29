package grpc

import (
	"google.golang.org/grpc/codes"
	"testing"

	"gophkeeper/internal/common"
	"gophkeeper/storage"

	"github.com/stretchr/testify/assert"
)

func TestCreateErrorResponse(t *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	resp, err := server.createErrorResponse(codes.Internal, "test error")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "test error", resp.Error)
}

func TestGetSecretByID(t *testing.T) {
	storage := storage.NewMemoryStorage()
	server := NewGRPCServer(storage, "test-secret")

	userID := common.MustParseUUID("12345678-1234-1234-1234-123456789012")
	secretID := common.MustParseUUID("12345678-1234-1234-1234-123456789013")

	// Secret not found
	secret, err := server.getSecretByID(userID, secretID)
	assert.Error(t, err)
	assert.Nil(t, secret)
}
