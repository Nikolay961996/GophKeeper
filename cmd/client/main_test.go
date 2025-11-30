package main

import (
	cobrakeeper "gophkeeper/internal/client/cobra"
	"gophkeeper/internal/client/commands"
	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/grpc"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestCommandCreation тестирует только создание команд
func TestCommandCreation(t *testing.T) {
	cfg := &config.Config{Token: "test-token"}
	authCommands := &commands.AuthCommands{}
	grpcClient := &grpc.GRPCClient{}

	commands := []*cobra.Command{
		cobrakeeper.VersionCmd(version, commit, date),
		cobrakeeper.UserCmd(cfg),
		cobrakeeper.RegisterCmd(authCommands),
		cobrakeeper.LoginCmd(authCommands),
		cobrakeeper.SyncCmd(cfg, grpcClient, masterPassword),
		cobrakeeper.ListCmd(cfg, grpcClient, masterPassword),
		cobrakeeper.ShowCmd(cfg, grpcClient, masterPassword),
		cobrakeeper.DelCmd(cfg, grpcClient, masterPassword),
		cobrakeeper.AddLoginCmd(cfg, grpcClient, masterPassword),
		cobrakeeper.AddCardCmd(cfg, grpcClient, masterPassword),
		cobrakeeper.AddTextCmd(cfg, grpcClient, masterPassword),
		cobrakeeper.AddFileCmd(cfg, grpcClient, masterPassword),
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

	manager, err := cobrakeeper.GetDataManager(true, cfg, masterPassword)
	assert.Error(t, err)
	assert.Nil(t, manager)

	cfg.Token = ""
	manager, err = cobrakeeper.GetDataManager(true, cfg, masterPassword)
	assert.Error(t, err)
	assert.Nil(t, manager)
}
