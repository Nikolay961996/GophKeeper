package main

import (
	"github.com/spf13/cobra"
	"testing"

	"gophkeeper/internal/client/commands"
	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/grpc"

	"github.com/stretchr/testify/assert"
)

func TestCommandCreationManual(t *testing.T) {
	cfg := &config.Config{Token: "test-token"}

	// Mock зависимости
	authCommands := &commands.AuthCommands{}
	var grpcClient *grpc.GRPCClient

	// Проверяем создание всех команд
	commands := []struct {
		name string
		cmd  func() *cobra.Command
	}{
		{"version", versionCmd},
		{"user", func() *cobra.Command { return userCmd(cfg) }},
		{"register", func() *cobra.Command { return registerCmd(authCommands) }},
		{"login", func() *cobra.Command { return loginCmd(authCommands) }},
		{"sync", func() *cobra.Command { return syncCmd(cfg, grpcClient) }},
		{"list", func() *cobra.Command { return listCmd(cfg, grpcClient) }},
		{"show", func() *cobra.Command { return showCmd(cfg, grpcClient) }},
		{"del", func() *cobra.Command { return delCmd(cfg, grpcClient) }},
		{"add-login", func() *cobra.Command { return addLoginCmd(cfg, grpcClient) }},
		{"add-card", func() *cobra.Command { return addCardCmd(cfg, grpcClient) }},
		{"add-text", func() *cobra.Command { return addTextCmd(cfg, grpcClient) }},
		{"add-file", func() *cobra.Command { return addFileCmd(cfg, grpcClient) }},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			cmd := tc.cmd()
			assert.NotNil(t, cmd)
			assert.NotEmpty(t, cmd.Use)
			assert.NotEmpty(t, cmd.Short)
		})
	}
}

func TestGetDataManager_ErrorCases(t *testing.T) {
	// No token
	cfg := &config.Config{Token: ""}
	manager, err := getDataManager(false, cfg)
	assert.Error(t, err)
	assert.Nil(t, manager)

	// Need master password but not provided
	cfg.Token = "test-token"
	manager, err = getDataManager(true, cfg)
	assert.Error(t, err)
	assert.Nil(t, manager)
}
