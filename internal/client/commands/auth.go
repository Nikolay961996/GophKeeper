package commands

import (
	"fmt"
	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/grpc"
)

// AuthCommands обработчики команд аутентификации
type AuthCommands struct {
	cfg        *config.Config
	grpcClient *grpc.GRPCClient
}

// NewAuthCommands создает новый AuthCommands
func NewAuthCommands(cfg *config.Config, grpcClient *grpc.GRPCClient) *AuthCommands {
	return &AuthCommands{
		cfg:        cfg,
		grpcClient: grpcClient,
	}
}

// Register регистрирует нового пользователя через gRPC
func (a *AuthCommands) Register(login, password string) error {
	authResult, err := a.grpcClient.Register(login, password)
	if err != nil {
		return err
	}

	a.cfg.Token = authResult.Token
	a.cfg.UserID = authResult.User.ID.String()

	if err := config.SaveConfig(a.cfg); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	fmt.Printf("Successfully registered user: %s\n", login)
	return nil
}

// Login выполняет вход пользователя через gRPC
func (a *AuthCommands) Login(login, password string) error {
	authResult, err := a.grpcClient.Login(login, password) // ← ИСПОЛЬЗУЕМ gRPC
	if err != nil {
		return err
	}

	a.cfg.Token = authResult.Token
	a.cfg.UserID = authResult.User.ID.String()

	if err := config.SaveConfig(a.cfg); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	fmt.Printf("Successfully logged in as: %s\n", login)
	return nil
}
