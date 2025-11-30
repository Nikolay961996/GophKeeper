// Package cobrakeeper for CLI commands
package cobrakeeper

import (
	"fmt"
	"github.com/spf13/cobra"
	"gophkeeper/internal/client/commands"
	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/manager"
	"gophkeeper/internal/common"

	"os"
)

var (
	dataManager *manager.DataManager
)

func VersionCmd(version, commit, date string) *cobra.Command {
	var v = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GophKeeper Client\n")
			fmt.Printf("Version: %s\n", version)
			fmt.Printf("Commit: %s\n", commit)
			fmt.Printf("Build Date: %s\n", date)
		},
	}

	return v
}

func UserCmd(cfg *config.Config) *cobra.Command {
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
			if userID, exists := claims["user_id"]; exists {
				fmt.Printf("user_id: %s\n", userID)
			} else {
				fmt.Println("Поле 'user_id' не найдено в токене")
			}
		},
	}

	return version
}

func SyncCmd(cfg *config.Config, grpcClient *grpc.GRPCClient, masterPassword string) *cobra.Command {
	var sync = &cobra.Command{
		Use:   "sync",
		Short: "Synchronize data with server",
		Run: func(cmd *cobra.Command, args []string) {
			m, err := GetDataManager(false, cfg, masterPassword)
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

func ShowCmd(cfg *config.Config, grpcClient *grpc.GRPCClient, masterPassword string) *cobra.Command {
	var show = &cobra.Command{
		Use:   "show [id]",
		Short: "Show data by id",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			m, err := GetDataManager(true, cfg, masterPassword)
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

func DelCmd(cfg *config.Config, grpcClient *grpc.GRPCClient, masterPassword string) *cobra.Command {
	var del = &cobra.Command{
		Use:   "del [id]",
		Short: "Delete data by id",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			m, err := GetDataManager(false, cfg, masterPassword)
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

func LoginCmd(authCommands *commands.AuthCommands) *cobra.Command {
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

func RegisterCmd(authCommands *commands.AuthCommands) *cobra.Command {
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

func ListCmd(cfg *config.Config, grpcClient *grpc.GRPCClient, masterPassword string) *cobra.Command {
	var list = &cobra.Command{
		Use:   "list",
		Short: "List all stored data",
		Run: func(cmd *cobra.Command, args []string) {
			m, err := GetDataManager(false, cfg, masterPassword)
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

func AddLoginCmd(cfg *config.Config, grpcClient *grpc.GRPCClient, masterPassword string) *cobra.Command {
	var addLogin = &cobra.Command{
		Use:   "add-login [name] [login] [password]",
		Short: "Add login/password",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			m, err := GetDataManager(true, cfg, masterPassword)
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

func AddCardCmd(cfg *config.Config, grpcClient *grpc.GRPCClient, masterPassword string) *cobra.Command {
	var addCard = &cobra.Command{
		Use:   "add-card [name] [number] [expiry] [cvv] [holder] [bank]",
		Short: "Add card data",
		Args:  cobra.ExactArgs(6),
		Run: func(cmd *cobra.Command, args []string) {
			m, err := GetDataManager(true, cfg, masterPassword)
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

func AddTextCmd(cfg *config.Config, grpcClient *grpc.GRPCClient, masterPassword string) *cobra.Command {
	var addText = &cobra.Command{
		Use:   "add-text [name] [text]",
		Short: "Add text data",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			m, err := GetDataManager(true, cfg, masterPassword)
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

func AddFileCmd(cfg *config.Config, grpcClient *grpc.GRPCClient, masterPassword string) *cobra.Command {
	var addFile = &cobra.Command{
		Use:   "add-file [name] [path]",
		Short: "Add file data",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			m, err := GetDataManager(true, cfg, masterPassword)
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

func GetDataManager(needMasterPassword bool, cfg *config.Config, masterPassword string) (*manager.DataManager, error) {
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
