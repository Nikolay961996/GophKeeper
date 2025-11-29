// Package grpc contains client implementation
package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"gophkeeper/api"
	"gophkeeper/internal/client/config"
	"gophkeeper/internal/common"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// GRPCClient представляет gRPC клиент
type GRPCClient struct {
	conn   *grpc.ClientConn
	client api.GophKeeperClient
	cfg    *config.Config
}

// NewGRPCClient создает новый gRPC клиент
func NewGRPCClient(cfg *config.Config) (*GRPCClient, error) {
	grpcAddr := "localhost:8081"

	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %v", err)
	}

	return &GRPCClient{
		conn:   conn,
		client: api.NewGophKeeperClient(conn),
		cfg:    cfg,
	}, nil
}

// Close закрывает соединение
func (c *GRPCClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// withAuth добавляет токен аутентификации в контекст
func (c *GRPCClient) withAuth(ctx context.Context) context.Context {
	if c.cfg.Token != "" {
		md := metadata.New(map[string]string{"authorization": "Bearer " + c.cfg.Token})
		return metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}

// Execute выполняет операцию через Opaque API
func (c *GRPCClient) Execute(operation *common.OperationRequest) (*common.OperationResponse, error) {
	// Маршалим операцию
	payload, err := common.MarshalOperation(operation)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal operation: %v", err)
	}

	req := &api.CommandRequest{
		Payload: payload,
	}

	// Отправляем запрос
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if c.client == nil {
		return nil, fmt.Errorf("no gRPC client")
	}

	resp, err := c.client.Execute(c.withAuth(ctx), req)
	if err != nil {
		return nil, fmt.Errorf("gRPC call failed: %v", err)
	}

	return &common.OperationResponse{
		Success: resp.Success,
		Error:   resp.Error,
		Payload: resp.Payload,
	}, nil
}

// Register регистрация пользователя
func (c *GRPCClient) Register(login, password string) (*common.AuthResult, error) {
	op := common.RegisterOp{
		Login:    login,
		Password: password,
	}

	req, err := common.CreateOperationRequest(common.OpRegister, op)
	if err != nil {
		return nil, err
	}

	resp, err := c.Execute(req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("registration failed: %s", resp.Error)
	}

	var authResult common.AuthResult
	if err := common.UnmarshalOperation(resp.Payload, &authResult); err != nil {
		return nil, fmt.Errorf("failed to parse auth result: %v", err)
	}

	return &authResult, nil
}

// Login вход пользователя
func (c *GRPCClient) Login(login, password string) (*common.AuthResult, error) {
	op := common.LoginOp{
		Login:    login,
		Password: password,
	}

	req, err := common.CreateOperationRequest(common.OpLogin, op)
	if err != nil {
		return nil, err
	}

	resp, err := c.Execute(req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("login failed: %s", resp.Error)
	}

	var authResult common.AuthResult
	if err := common.UnmarshalOperation(resp.Payload, &authResult); err != nil {
		return nil, fmt.Errorf("failed to parse auth result: %v", err)
	}

	return &authResult, nil
}

// Sync синхронизация данных
func (c *GRPCClient) Sync(lastSync time.Time, data []common.SecretData) (*common.SyncResult, error) {
	op := common.SyncOp{
		LastSync: lastSync,
		Data:     data,
	}

	req, err := common.CreateOperationRequest(common.OpSync, op)
	if err != nil {
		return nil, err
	}

	resp, err := c.Execute(req)
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf("sync failed: %s", resp.Error)
	}

	var syncResult common.SyncResult
	if err := common.UnmarshalOperation(resp.Payload, &syncResult); err != nil {
		return nil, fmt.Errorf("failed to parse sync result: %v", err)
	}

	return &syncResult, nil
}

// UploadFile загружает файл на сервер (без изменений)
func (c *GRPCClient) UploadFile(filePath, fileID string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer func(file *os.File) {
		e := file.Close()
		if e != nil {
			log.Fatalf("failed to close file: %v", e)
		}
	}(file)

	stream, err := c.client.UploadFile(c.withAuth(context.Background()))
	if err != nil {
		return fmt.Errorf("failed to create upload stream: %v", err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	chunkSize := 64 * 1024 // 64 KB
	buffer := make([]byte, chunkSize)

	for i := 0; ; i++ {
		n, e := file.Read(buffer)
		if e == io.EOF {
			break
		}
		if e != nil {
			return fmt.Errorf("failed to read file: %v", e)
		}

		chunk := &api.FileChunk{
			FileId:      fileID,
			FileName:    fileID,
			ChunkData:   buffer[:n],
			ChunkIndex:  int32(i),
			TotalChunks: int32((int(fileInfo.Size()) + chunkSize - 1) / chunkSize),
		}

		if err = stream.Send(chunk); err != nil {
			return fmt.Errorf("failed to send chunk: %v", err)
		}
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("upload failed: %v", err)
	}

	if response.Success {
		fmt.Printf("File uploaded successfully: %s (%d bytes)\n", fileID, response.FileSize)
	} else {
		return fmt.Errorf("upload failed")
	}

	return nil
}

// DownloadFile скачивает файл с сервера (без изменений)
func (c *GRPCClient) DownloadFile(fileID, outputPath string) error {
	req := &api.DownloadRequest{FileId: fileID}
	if c.client == nil {
		return fmt.Errorf("no gRPC client")
	}
	stream, err := c.client.DownloadFile(c.withAuth(context.Background()), req)
	if err != nil {
		return fmt.Errorf("failed to create download stream: %v", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("close failed: %v", err)
		}
	}(file)

	var fileName string
	var totalSize int

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive chunk: %v", err)
		}

		if fileName == "" {
			fileName = chunk.FileName
		}

		if _, err := file.Write(chunk.ChunkData); err != nil {
			return fmt.Errorf("failed to write chunk: %v", err)
		}

		totalSize += len(chunk.ChunkData)
		fmt.Printf("\rDownloading: %s - %d/%d chunks", fileName, chunk.ChunkIndex+1, chunk.TotalChunks)
	}

	fmt.Printf("\nFile downloaded successfully: %s (%d bytes)\n", outputPath, totalSize)
	return nil
}
