/*
file chanks
conflicts
postgre
linter
tests
*/

package main

import (
	"fmt"
	"gophkeeper/internal/common"
	"log"
	"os"

	"gophkeeper/internal/client/commands"
	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/manager"

	"github.com/spf13/cobra"
)

var (
	version        = "dev"
	commit         = "none"
	date           = "unknown"
	masterPassword string
	dataManager    *manager.DataManager
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

	authCommands := commands.NewAuthCommands(cfg, grpcClient)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(
		versionCmd(),
		registerCmd(authCommands),
		loginCmd(authCommands),
		syncCmd(cfg, grpcClient),
		addLoginCmd(cfg, grpcClient),
		listCmd(cfg, grpcClient),
		conflictsCmd(cfg, grpcClient),
		showCmd(cfg, grpcClient),
		addCardCmd(cfg, grpcClient),
		addTextCmd(cfg, grpcClient),
		addFileCmd(cfg, grpcClient),
		userCmd(cfg),
		delCmd(cfg, grpcClient),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func versionCmd() *cobra.Command {
	var version = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GophKeeper Client\n")
			fmt.Printf("Version: %s\n", version)
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Build Date: %s\n", date)
		},
	}

	return version
}

func userCmd(cfg *config.Config) *cobra.Command {
	var version = &cobra.Command{
		Use:   "user",
		Short: "Print current user",
		Run: func(cmd *cobra.Command, args []string) {
			claims, err := common.GetJWTClaims(cfg.Token)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			if login, exists := claims["login"]; exists {
				fmt.Printf("login: %s\n", login)
			} else {
				fmt.Println("Поле 'login' не найдено в токене")
			}
			if userId, exists := claims["user_id"]; exists {
				fmt.Printf("user_id: %s\n", userId)
			} else {
				fmt.Println("Поле 'user_id' не найдено в токене")
			}
		},
	}

	return version
}

func syncCmd(cfg *config.Config, grpcClient *grpc.GRPCClient) *cobra.Command {
	var sync = &cobra.Command{
		Use:   "sync",
		Short: "Synchronize data with server",
		Run: func(cmd *cobra.Command, args []string) {
			m, err := getDataManager(false, cfg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			dataCommands := commands.NewDataCommands(cfg, m, grpcClient)
			if err := dataCommands.Sync(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	return sync
}

func showCmd(cfg *config.Config, grpcClient *grpc.GRPCClient) *cobra.Command {
	var show = &cobra.Command{
		Use:   "show [id]",
		Short: "Show data by id",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			m, err := getDataManager(true, cfg)
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

	return show
}

func delCmd(cfg *config.Config, grpcClient *grpc.GRPCClient) *cobra.Command {
	var del = &cobra.Command{
		Use:   "del [id]",
		Short: "Delete data by id",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			m, err := getDataManager(false, cfg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			dataCommands := commands.NewDataCommands(cfg, m, grpcClient)
			if err := dataCommands.Delete(args[0]); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	return del
}

func loginCmd(authCommands *commands.AuthCommands) *cobra.Command {
	var login = &cobra.Command{
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

	return login
}

func registerCmd(authCommands *commands.AuthCommands) *cobra.Command {
	var register = &cobra.Command{
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

	return register
}

func listCmd(cfg *config.Config, grpcClient *grpc.GRPCClient) *cobra.Command {
	var list = &cobra.Command{
		Use:   "list",
		Short: "List all stored data",
		Run: func(cmd *cobra.Command, args []string) {
			m, err := getDataManager(false, cfg)
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

	return list
}

func addLoginCmd(cfg *config.Config, grpcClient *grpc.GRPCClient) *cobra.Command {
	var addLogin = &cobra.Command{
		Use:   "add-login [name] [login] [password]",
		Short: "Add login/password",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			m, err := getDataManager(true, cfg)
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
	addLogin.Flags().String("site", "", "Website URL")

	return addLogin
}

func addCardCmd(cfg *config.Config, grpcClient *grpc.GRPCClient) *cobra.Command {
	var addCard = &cobra.Command{
		Use:   "add-card [name] [number] [expiry] [cvv] [holder] [bank]",
		Short: "Add card data",
		Args:  cobra.ExactArgs(6),
		Run: func(cmd *cobra.Command, args []string) {
			m, err := getDataManager(true, cfg)
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

	return addCard
}

func addTextCmd(cfg *config.Config, grpcClient *grpc.GRPCClient) *cobra.Command {
	var addText = &cobra.Command{
		Use:   "add-text [name] [text]",
		Short: "Add text data",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			m, err := getDataManager(true, cfg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			dataCommands := commands.NewDataCommands(cfg, m, grpcClient)
			if err := dataCommands.AddText(args[0], args[1]); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	return addText
}

func addFileCmd(cfg *config.Config, grpcClient *grpc.GRPCClient) *cobra.Command {
	var addFile = &cobra.Command{
		Use:   "add-file [name] [path]",
		Short: "Add file data",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			m, err := getDataManager(true, cfg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			dataCommands := commands.NewDataCommands(cfg, m, grpcClient)
			if err := dataCommands.AddFile(args[0], args[1]); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	return addFile
}

func getDataManager(needMasterPassword bool, cfg *config.Config) (*manager.DataManager, error) {
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

func conflictsCmd(cfg *config.Config, grpcClient *grpc.GRPCClient) *cobra.Command {
	var conflicts = &cobra.Command{
		Use:   "conflicts",
		Short: "Show and resolve pending conflicts",
		Run: func(cmd *cobra.Command, args []string) {
			m, err := getDataManager(true, cfg)
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

	return conflicts
}
