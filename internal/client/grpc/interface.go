package grpc

import (
	"time"

	"gophkeeper/internal/common"
)

// ClientInterface определяет контракт для gRPC клиента
type ClientInterface interface {
	// Auth methods
	Register(login, password string) (*common.AuthResult, error)
	Login(login, password string) (*common.AuthResult, error)

	// Data methods
	Sync(lastSync time.Time, data []common.SecretData) (*common.SyncResult, error)
	Execute(operation *common.OperationRequest) (*common.OperationResponse, error)

	// File methods
	UploadFile(filePath, fileID string) error
	DownloadFile(fileID, outputPath string) error

	// Connection management
	Close() error
	//withAuth(ctx context.Context) context.Context
	//GetConn() *grpc.ClientConn
}

// Убедимся, что GRPCClient реализует интерфейс
var _ ClientInterface = (*GRPCClient)(nil)
