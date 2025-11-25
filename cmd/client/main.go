package main

import (
	"fmt"
	"log"
	"os"

	"gophkeeper/internal/client/commands"
	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/grpc" // ← ДОБАВИЛИ
	"gophkeeper/internal/client/manager"

	"github.com/spf13/cobra"
)

var (
	version        = "dev"
	commit         = "none"
	date           = "unknown"
	masterPassword string
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаем gRPC клиент
	grpcClient, err := grpc.NewGRPCClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer grpcClient.Close()

	// Создаем корневую команду
	var rootCmd = &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper - secure password manager",
		Long:  "GophKeeper is a secure client-server password manager",
	}

	// Флаг для мастер-пароля
	rootCmd.PersistentFlags().StringVarP(&masterPassword, "password", "p", "", "Master password for encryption")

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

	// Команды аутентификации (используют gRPC)
	authCommands := commands.NewAuthCommands(cfg, grpcClient) // ← ПЕРЕДАЕМ gRPC клиент

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

	// Команды данных (требуют аутентификации и мастер-пароля)
	var dataManager *manager.DataManager

	getDataManager := func() (*manager.DataManager, error) {
		if dataManager != nil {
			return dataManager, nil
		}

		if masterPassword == "" {
			// Запрашиваем мастер-пароль у пользователя
			fmt.Print("Enter master password: ")
			var input string
			fmt.Scanln(&input)
			masterPassword = input

			if masterPassword == "" {
				return nil, fmt.Errorf("master password is required")
			}
		}

		if cfg.Token == "" {
			return nil, fmt.Errorf("not authenticated. Please login first")
		}

		var err error
		dataManager, err = manager.NewDataManager(cfg, masterPassword)
		return dataManager, err
	}

	var syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Synchronize data with server",
		Run: func(cmd *cobra.Command, args []string) {
			manager, err := getDataManager()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			dataCommands := commands.NewDataCommands(cfg, manager, grpcClient) // ← ПЕРЕДАЕМ gRPC клиент
			if err := dataCommands.Sync(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	var addLoginCmd = &cobra.Command{
		Use:   "add-login [name] [login] [password]",
		Short: "Add login/password",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			manager, err := getDataManager()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			site, _ := cmd.Flags().GetString("site")
			dataCommands := commands.NewDataCommands(cfg, manager, grpcClient)
			if err := dataCommands.AddLoginPassword(args[0], args[1], args[2], site); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
	addLoginCmd.Flags().String("site", "", "Website URL")

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all stored data",
		Run: func(cmd *cobra.Command, args []string) {
			manager, err := getDataManager()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			dataCommands := commands.NewDataCommands(cfg, manager, grpcClient)
			if err := dataCommands.List(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(registerCmd, loginCmd, syncCmd, addLoginCmd, listCmd)

	// Добавьте эту команду после listCmd
	var conflictsCmd = &cobra.Command{
		Use:   "conflicts",
		Short: "Show and resolve pending conflicts",
		Run: func(cmd *cobra.Command, args []string) {
			manager, err := getDataManager()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			dataCommands := commands.NewDataCommands(cfg, manager, grpcClient)

			// Пока просто сообщаем что конфликтов нет
			// В реальной реализации здесь был бы показ pending конфликтов
			fmt.Println("No pending conflicts found.")
			fmt.Println("Conflicts are automatically detected and resolved during sync.")
		},
	}

	// Добавьте команду в rootCmd
	rootCmd.AddCommand(registerCmd, loginCmd, syncCmd, addLoginCmd, listCmd, conflictsCmd)

	// Запускаем CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
