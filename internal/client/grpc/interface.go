package grpc

import (
	"time"

	"gophkeeper/internal/common"
)

// ClientInterface определяет контракт для gRPC клиента
type ClientInterface interface {
	Register(login, password string) (*common.AuthResult, error)
	Login(login, password string) (*common.AuthResult, error)

	Sync(lastSync time.Time, data []common.SecretData) (*common.SyncResult, error)
	Execute(operation *common.OperationRequest) (*common.OperationResponse, error)

	UploadFile(filePath, fileID string) error
	DownloadFile(fileID, outputPath string) error

	Close() error
}

var _ ClientInterface = (*GRPCClient)(nil)
