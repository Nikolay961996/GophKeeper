package main

import (
	"fmt"
	"log"
	"os"

	"gophkeeper/internal/client/commands"
	"gophkeeper/internal/client/config"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаем корневую команду
	var rootCmd = &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper - secure password manager",
		Long:  "GophKeeper is a secure client-server password manager",
	}

	// Команда версии
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GophKeeper Client\n")
			fmt.Printf("Version: %s\n", version)
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Build Date: %s\n", date)
		},
	})

	// Команды аутентификации
	authCommands := commands.NewAuthCommands(cfg)

	var registerCmd = &cobra.Command{
		Use:   "register [login] [password]",
		Short: "Register a new user",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := authCommands.Register(args[0], args[1]); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	var loginCmd = &cobra.Command{
		Use:   "login [login] [password]",
		Short: "Login user",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := authCommands.Login(args[0], args[1]); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(registerCmd, loginCmd)

	// Запускаем CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
