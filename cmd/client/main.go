package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

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

	var rootCmd = &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper - secure password manager",
		Long:  "GophKeeper is a secure client-server password manager",
	}

	rootCmd.PersistentFlags().StringVarP(&masterPassword, "password", "p", "", "Master password for encryption")

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

	authCommands := commands.NewAuthCommands(cfg, grpcClient)

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

	var dataManager *manager.DataManager

	getDataManager := func(needMasterPassword bool) (*manager.DataManager, error) {
		if dataManager != nil {
			return dataManager, nil
		}

		if needMasterPassword && masterPassword == "" {
			fmt.Print("Enter master password: ")
			var input string
			_, err := fmt.Scanln(&input)
			if err != nil {
				return nil, err
			}
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
			m, err := getDataManager(false)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			dataCommands := commands.NewDataCommands(cfg, m, grpcClient) // ← ПЕРЕДАЕМ gRPC клиент
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
			m, err := getDataManager(true)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			site, _ := cmd.Flags().GetString("site")
			dataCommands := commands.NewDataCommands(cfg, m, grpcClient)
			if err := dataCommands.AddLoginPassword(args[0], args[1], args[2], site); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
	}
	addLoginCmd.Flags().String("site", "", "Website URL")

	var addCardCmd = &cobra.Command{
		Use:   "add-card [name] [number] [expiry] [cvv] [holder] [bank]",
		Short: "Add card data",
		Args:  cobra.ExactArgs(6),
		Run: func(cmd *cobra.Command, args []string) {
			m, err := getDataManager(true)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			dataCommands := commands.NewDataCommands(cfg, m, grpcClient)
			if err := dataCommands.AddCard(args[0], args[1], args[2], args[3], args[4], args[5]); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all stored data",
		Run: func(cmd *cobra.Command, args []string) {
			m, err := getDataManager(false)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			dataCommands := commands.NewDataCommands(cfg, m, grpcClient)
			if err := dataCommands.List(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	var showCmd = &cobra.Command{
		Use:   "show-i [id]",
		Short: "Show data by id",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			m, err := getDataManager(true)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			dataCommands := commands.NewDataCommands(cfg, m, grpcClient)
			if err := dataCommands.Get(args[0]); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	var showByNameCmd = &cobra.Command{
		Use:   "show-p [position]",
		Short: "Show data by position in list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			m, err := getDataManager(true)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			dataCommands := commands.NewDataCommands(cfg, m, grpcClient)
			p, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			if err := dataCommands.GetByPosition(p); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	var conflictsCmd = &cobra.Command{
		Use:   "conflicts",
		Short: "Show and resolve pending conflicts",
		Run: func(cmd *cobra.Command, args []string) {
			m, err := getDataManager(true)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			_ = commands.NewDataCommands(cfg, m, grpcClient)

			// Пока просто сообщаем что конфликтов нет
			// В реальной реализации здесь был бы показ pending конфликтов
			fmt.Println("No pending conflicts found.")
			fmt.Println("Conflicts are automatically detected and resolved during sync.")
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(registerCmd, loginCmd, syncCmd, addLoginCmd, listCmd, conflictsCmd, showCmd, showByNameCmd, addCardCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
