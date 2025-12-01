package main

import (
	"fmt"
	cobrakeeper "gophkeeper/internal/client/cobra"
	"log"
	"os"

	"gophkeeper/internal/client/commands"
	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/grpc"

	"github.com/spf13/cobra"
)

var (
	version        = "dev"
	commit         = "none"
	date           = "unknown"
	masterPassword string
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	grpcClient, err := grpc.NewGRPCClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer func(grpcClient *grpc.GRPCClient) {
		err := grpcClient.Close()
		if err != nil {
			log.Fatalf("Failed to close gRPC client: %v", err)
		}
	}(grpcClient)

	runCLI(cfg, grpcClient)
}

func runCLI(cfg *config.Config, grpcClient *grpc.GRPCClient) {
	var rootCmd = &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper - secure password manager",
		Long:  "GophKeeper is a secure client-server password manager",
	}
	rootCmd.PersistentFlags().StringVarP(&masterPassword, "password", "p", "", "Master password for encryption")

	authCommands := commands.NewAuthCommands(cfg, grpcClient)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(
		cobrakeeper.VersionCmd(version, commit, date),
		cobrakeeper.RegisterCmd(authCommands),
		cobrakeeper.LoginCmd(authCommands),
		cobrakeeper.SyncCmd(cfg, grpcClient, masterPassword),
		cobrakeeper.AddLoginCmd(cfg, grpcClient, masterPassword),
		cobrakeeper.ListCmd(cfg, grpcClient, masterPassword),
		cobrakeeper.ShowCmd(cfg, grpcClient, masterPassword),
		cobrakeeper.AddCardCmd(cfg, grpcClient, masterPassword),
		cobrakeeper.AddTextCmd(cfg, grpcClient, masterPassword),
		cobrakeeper.AddFileCmd(cfg, grpcClient, masterPassword),
		cobrakeeper.UserCmd(cfg),
		cobrakeeper.DelCmd(cfg, grpcClient, masterPassword),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
