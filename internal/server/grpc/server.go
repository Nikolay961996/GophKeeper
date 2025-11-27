package grpc

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"gophkeeper/api"
	"gophkeeper/internal/common"
	"gophkeeper/storage"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Константа для ключа контекста (перенесем из middleware)
const userIDKey = "userID"

// GRPCServer представляет gRPC сервер
type GRPCServer struct {
	api.UnimplementedGophKeeperServer
	storage    storage.Storage
	jwtSecret  string
	grpcServer *grpc.Server
}

// NewGRPCServer создает новый gRPC сервер
func NewGRPCServer(storage storage.Storage, jwtSecret string) *GRPCServer {
	return &GRPCServer{
		storage:   storage,
		jwtSecret: jwtSecret,
	}
}

// Start запускает gRPC сервер
func (s *GRPCServer) Start(addr string) error {
	// Создаем gRPC сервер с интерцептором аутентификации
	s.grpcServer = grpc.NewServer(
		grpc.StreamInterceptor(s.streamAuthInterceptor),
		grpc.UnaryInterceptor(s.unaryAuthInterceptor),
	)

	api.RegisterGophKeeperServer(s.grpcServer, s)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	log.Printf("gRPC server listening on %s", addr)
	return s.grpcServer.Serve(lis)
}

// Stop останавливает gRPC сервер
func (s *GRPCServer) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

// Execute - универсальный Opaque API метод
func (s *GRPCServer) Execute(ctx context.Context, req *api.CommandRequest) (*api.CommandResponse, error) {
	log.Println("Execute")
	var opReq common.OperationRequest
	if err := common.UnmarshalOperation(req.Payload, &opReq); err != nil {
		return s.createErrorResponse(codes.InvalidArgument, "invalid operation format")
	}

	var result *common.OperationResponse
	var err error

	switch opReq.Type {
	case common.OpRegister:
		result, err = s.handleRegister(opReq.Payload)
	case common.OpLogin:
		result, err = s.handleLogin(opReq.Payload)
	case common.OpSync:
		userID, authErr := s.authenticate(ctx)
		if authErr != nil {
			return s.createErrorResponse(codes.Unauthenticated, "authentication required")
		}
		result, err = s.handleSync(userID, opReq.Payload)
	default:
		return s.createErrorResponse(codes.InvalidArgument, "unknown operation")
	}

	if err != nil {
		return s.createErrorResponse(codes.Internal, err.Error())
	}

	return &api.CommandResponse{
		Success: result.Success,
		Error:   result.Error,
		Payload: result.Payload,
	}, nil
}

// handleRegister обработчик регистрации
func (s *GRPCServer) handleRegister(payload []byte) (*common.OperationResponse, error) {
	log.Println("handleRegister")

	var registerOp common.RegisterOp
	if err := common.UnmarshalOperation(payload, &registerOp); err != nil {
		return common.CreateErrorResponse(fmt.Errorf("invalid register data")), nil
	}

	existingUser, _ := s.storage.GetUserByLogin(registerOp.Login)
	if existingUser != nil {
		return common.CreateErrorResponse(fmt.Errorf("user already exists")), nil
	}

	passwordHash, err := common.HashPassword(registerOp.Password)
	if err != nil {
		return common.CreateErrorResponse(fmt.Errorf("error creating user")), nil
	}

	user := &common.User{
		ID:           uuid.New(),
		Login:        registerOp.Login,
		PasswordHash: passwordHash,
		CreatedAt:    common.Now(),
	}

	if err := s.storage.CreateUser(user); err != nil {
		return common.CreateErrorResponse(fmt.Errorf("error creating user")), nil
	}

	token, err := common.GenerateJWTToken(user.ID, user.Login, s.jwtSecret, common.TokenExpiration)
	if err != nil {
		return common.CreateErrorResponse(fmt.Errorf("error generating token")), nil
	}

	authResult := common.AuthResult{
		Token: token,
		User:  *user,
	}

	return common.CreateSuccessResponse(authResult)
}

// handleLogin обработчик входа
func (s *GRPCServer) handleLogin(payload []byte) (*common.OperationResponse, error) {
	log.Println("handleLogin")

	var loginOp common.LoginOp
	if err := common.UnmarshalOperation(payload, &loginOp); err != nil {
		return common.CreateErrorResponse(fmt.Errorf("invalid login data")), nil
	}

	user, err := s.storage.GetUserByLogin(loginOp.Login)
	if err != nil {
		return common.CreateErrorResponse(fmt.Errorf("invalid credentials")), nil
	}

	if !common.CheckPasswordHash(loginOp.Password, user.PasswordHash) {
		return common.CreateErrorResponse(fmt.Errorf("invalid credentials")), nil
	}

	token, err := common.GenerateJWTToken(user.ID, user.Login, s.jwtSecret, common.TokenExpiration)
	if err != nil {
		return common.CreateErrorResponse(fmt.Errorf("error generating token")), nil
	}

	authResult := common.AuthResult{
		Token: token,
		User:  *user,
	}

	return common.CreateSuccessResponse(authResult)
}

// handleSync обработчик синхронизации
func (s *GRPCServer) handleSync(userID uuid.UUID, payload []byte) (*common.OperationResponse, error) {
	log.Println("handleSync")

	var syncOp common.SyncOp
	if err := common.UnmarshalOperation(payload, &syncOp); err != nil {
		return common.CreateErrorResponse(fmt.Errorf("invalid sync data")), nil
	}

	log.Println("1")

	// Получаем изменения с сервера
	serverSecretsPtr, err := s.storage.GetSecretsSince(userID, syncOp.LastSync)
	if err != nil {
		log.Fatalf(err.Error())
		return common.CreateErrorResponse(fmt.Errorf("error getting secrets")), nil
	}

	// Конвертируем []*common.SecretData в []common.SecretData
	serverSecrets := make([]common.SecretData, len(serverSecretsPtr))
	for i, secretPtr := range serverSecretsPtr {
		serverSecrets[i] = *secretPtr
	}
	log.Println("2")

	// Сохраняем изменения от клиента и собираем конфликты
	var conflicts []common.SecretData
	for _, clientSecret := range syncOp.Data {
		existingSecret, _ := s.getSecretByID(userID, clientSecret.ID)

		if existingSecret != nil {
			// Проверяем версию для обнаружения конфликтов
			if existingSecret.Version > clientSecret.Version {
				// Конфликт: серверная версия новее
				conflicts = append(conflicts, *existingSecret)
				continue
			}

			// Инкрементируем версию при сохранении
			clientSecret.Version = existingSecret.Version + 1
		} else {
			// Новая запись
			clientSecret.Version = 1
		}

		clientSecret.UserID = userID
		clientSecret.UpdatedAt = common.Now()
		if err := s.storage.SaveSecretData(&clientSecret); err != nil {
			return common.CreateErrorResponse(fmt.Errorf("error saving secret")), nil
		}
	}

	log.Println("3")

	syncResult := common.SyncResult{
		LastSync:  common.Now(),
		Data:      serverSecrets,
		Conflicts: conflicts,
	}

	log.Println("4")

	return common.CreateSuccessResponse(syncResult)
}

// UploadFile потоковая загрузка файла
func (s *GRPCServer) UploadFile(stream api.GophKeeper_UploadFileServer) error {
	log.Println("UploadFile")

	// Используем правильный ключ контекста
	userIDValue := stream.Context().Value(userIDKey)
	if userIDValue == nil {
		return status.Error(codes.Unauthenticated, "user not authenticated")
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return status.Error(codes.Unauthenticated, "invalid user ID")
	}

	var fileData []byte
	var fileName string
	var fileID string

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("failed to receive chunk: %v", err))
		}

		if fileID == "" {
			fileID = chunk.FileId
			fileName = chunk.FileName
		}

		fileData = append(fileData, chunk.ChunkData...)
	}

	secret := &common.SecretData{
		ID:        uuid.MustParse(fileID),
		UserID:    userID,
		Type:      common.BinaryDataType,
		Metadata:  fileName,
		Data:      fileData,
		Version:   1,
		CreatedAt: common.Now(),
		UpdatedAt: common.Now(),
	}

	if err := s.storage.SaveSecretData(secret); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to save file: %v", err))
	}

	return stream.SendAndClose(&api.UploadResponse{
		FileId:   fileID,
		FileSize: int64(len(fileData)),
		Success:  true,
	})
}

// DownloadFile потоковая выгрузка файла
func (s *GRPCServer) DownloadFile(req *api.DownloadRequest, stream api.GophKeeper_DownloadFileServer) error {
	log.Println("DownloadFile")

	// Используем правильный ключ контекста
	userIDValue := stream.Context().Value(userIDKey)
	if userIDValue == nil {
		return status.Error(codes.Unauthenticated, "user not authenticated")
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return status.Error(codes.Unauthenticated, "invalid user ID")
	}

	fileID, err := uuid.Parse(req.FileId)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid file ID")
	}

	secrets, err := s.storage.GetUserSecrets(userID)
	if err != nil {
		return status.Error(codes.NotFound, "file not found")
	}

	var fileSecret *common.SecretData
	for _, secret := range secrets {
		if secret.ID == fileID && secret.Type == common.BinaryDataType {
			fileSecret = secret
			break
		}
	}

	if fileSecret == nil {
		return status.Error(codes.NotFound, "file not found")
	}

	chunkSize := 64 * 1024
	data := fileSecret.Data
	totalChunks := (len(data) + chunkSize - 1) / chunkSize

	for i := 0; i < totalChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := &api.FileChunk{
			FileId:      fileSecret.ID.String(),
			FileName:    fileSecret.Metadata,
			ChunkData:   data[start:end],
			ChunkIndex:  int32(i),
			TotalChunks: int32(totalChunks),
		}

		if err := stream.Send(chunk); err != nil {
			return status.Error(codes.Internal, fmt.Sprintf("failed to send chunk: %v", err))
		}
	}

	return nil
}

// unaryAuthInterceptor перехватчик для униарных методов
func (s *GRPCServer) unaryAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Println("unaryAuthInterceptor")

	if info.FullMethod == "/gophkeeper.GophKeeper/Execute" {
		return handler(ctx, req)
	}

	userID, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, userIDKey, userID)
	return handler(ctx, req)
}

// streamAuthInterceptor перехватчик для потоковых методов
func (s *GRPCServer) streamAuthInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Println("streamAuthInterceptor")

	userID, err := s.authenticate(ss.Context())
	if err != nil {
		return err
	}

	ctx := context.WithValue(ss.Context(), userIDKey, userID)
	return handler(srv, &wrappedStream{ss, ctx})
}

func (s *GRPCServer) authenticate(ctx context.Context) (uuid.UUID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, status.Error(codes.Unauthenticated, "metadata not provided")
	}

	tokens := md["authorization"]
	if len(tokens) == 0 {
		return uuid.Nil, status.Error(codes.Unauthenticated, "authorization token not provided")
	}

	tokenString := tokens[0]
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	claims, err := common.ValidateJWTToken(tokenString, s.jwtSecret)
	if err != nil {
		return uuid.Nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return claims.UserID, nil
}

func (s *GRPCServer) getSecretByID(userID, secretID uuid.UUID) (*common.SecretData, error) {
	secrets, err := s.storage.GetUserSecrets(userID)
	if err != nil {
		return nil, err
	}

	for _, secret := range secrets {
		if secret.ID == secretID {
			return secret, nil
		}
	}

	return nil, storage.ErrSecretNotFound
}

func (s *GRPCServer) createErrorResponse(code codes.Code, message string) (*api.CommandResponse, error) {
	return &api.CommandResponse{
		Success: false,
		Error:   message,
	}, nil
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
