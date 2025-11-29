package main

import (
	"gophkeeper/internal/client/commands"
	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/grpc"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestCommandCreation тестирует только создание команд (быстрое покрытие)
func TestCommandCreation(t *testing.T) {
	// Создаем минимальные зависимости
	cfg := &config.Config{Token: "test-token"}

	// Mock команды
	authCommands := &commands.AuthCommands{}

	// Mock gRPC клиент (нужно будет создать интерфейс)
	grpcClient := &grpc.GRPCClient{}

	// Тестируем создание всех команд
	commands := []*cobra.Command{
		versionCmd(),
		userCmd(cfg),
		registerCmd(authCommands),
		loginCmd(authCommands),
		syncCmd(cfg, grpcClient),
		listCmd(cfg, grpcClient),
		showCmd(cfg, grpcClient),
		delCmd(cfg, grpcClient),
		addLoginCmd(cfg, grpcClient),
		addCardCmd(cfg, grpcClient),
		addTextCmd(cfg, grpcClient),
		addFileCmd(cfg, grpcClient),
	}

	for _, cmd := range commands {
		assert.NotNil(t, cmd, "Command should not be nil")
		assert.NotEmpty(t, cmd.Use, "Command Use should not be empty")
		assert.NotEmpty(t, cmd.Short, "Command Short should not be empty")
	}
}

// TestVersionVariables проверяет глобальные переменные версии
func TestVersionVariables(t *testing.T) {
	assert.NotEmpty(t, version)
	assert.NotEmpty(t, commit)
	assert.NotEmpty(t, date)
}

// TestGetDataManager проверяет функцию getDataManager
func TestGetDataManager(t *testing.T) {
	cfg := &config.Config{Token: "test-token"}

	// Тестируем без master password
	manager, err := getDataManager(true, cfg)
	assert.Error(t, err) // Должна быть ошибка аутентификации
	assert.Nil(t, manager)

	// Тестируем с master password (но без токена)
	cfg.Token = ""
	manager, err = getDataManager(true, cfg)
	assert.Error(t, err)
	assert.Nil(t, manager)
}
