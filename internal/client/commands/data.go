package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gophkeeper/internal/client/grpc"
	"io"
	"net/http"
	"os"
	"time"

	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/manager"
	"gophkeeper/internal/common"
)

// DataCommands обработчики команд для работы с данными
type DataCommands struct {
	cfg        *config.Config
	manager    *manager.DataManager
	grpcClient *grpc.GRPCClient
}

// NewDataCommands создает новый DataCommands
func NewDataCommands(cfg *config.Config, dataManager *manager.DataManager, grpcClient *grpc.GRPCClient) *DataCommands {
	return &DataCommands{
		cfg:        cfg,
		manager:    dataManager,
		grpcClient: grpcClient,
	}
}

// Sync синхронизирует данные с сервером через gRPC
func (d *DataCommands) Sync() error {
	if d.cfg.Token == "" {
		return fmt.Errorf("not authenticated. Please login first")
	}

	// Получаем локальные данные
	localSecrets := d.manager.ListData()

	// Находим время последней синхронизации
	var lastSync time.Time
	if len(localSecrets) > 0 {
		for _, secret := range localSecrets {
			if secret.UpdatedAt.After(lastSync) {
				lastSync = secret.UpdatedAt
			}
		}
	}

	// Конвертируем []*common.SecretData в []common.SecretData
	localSecretsData := make([]common.SecretData, len(localSecrets))
	for i, secret := range localSecrets {
		localSecretsData[i] = *secret
	}

	// Выполняем синхронизацию через gRPC
	syncResult, err := d.grpcClient.Sync(lastSync, localSecretsData)
	if err != nil {
		return err
	}

	// Сохраняем полученные данные
	for _, serverSecret := range syncResult.Data {
		if err := d.manager.SaveSecret(&serverSecret); err != nil {
			fmt.Printf("Warning: failed to save secret %s: %v\n", serverSecret.ID, err)
		}
	}

	fmt.Printf("Sync completed. Received %d items from server\n", len(syncResult.Data))
	if len(syncResult.Conflicts) > 0 {
		fmt.Printf("Warning: %d conflicts detected (using server version)\n", len(syncResult.Conflicts))
	}

	return nil
}

// AddLoginPassword добавляет логин/пароль
func (d *DataCommands) AddLoginPassword(name, login, password, site string) error {
	if err := d.manager.SaveLoginPassword(name, login, password, site); err != nil {
		return err
	}

	fmt.Printf("Login/password '%s' saved successfully\n", name)
	return d.Sync() // Синхронизируем после добавления
}

// AddCard добавляет данные карты
func (d *DataCommands) AddCard(name, number, expiry, cvv, holder, bank string) error {
	if err := d.manager.SaveCardData(name, number, expiry, cvv, holder, bank); err != nil {
		return err
	}

	fmt.Printf("Card data '%s' saved successfully\n", name)
	return d.Sync()
}

// AddText добавляет текстовые данные
func (d *DataCommands) AddText(name, text string) error {
	if err := d.manager.SaveTextData(name, text); err != nil {
		return err
	}

	fmt.Printf("Text data '%s' saved successfully\n", name)
	return d.Sync()
}

// AddFile добавляет файл
func (d *DataCommands) AddFile(name, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	if err := d.manager.SaveBinaryData(name, data); err != nil {
		return err
	}

	fmt.Printf("File '%s' saved successfully (%d bytes)\n", name, len(data))
	return d.Sync()
}

// List выводит список всех данных
func (d *DataCommands) List() error {
	secrets := d.manager.ListData()

	if len(secrets) == 0 {
		fmt.Println("No data stored")
		return nil
	}

	fmt.Printf("Stored data (%d items):\n", len(secrets))
	for i, secret := range secrets {
		fmt.Printf("%d. %s [%s] - %s\n", i+1, secret.Metadata, secret.Type, secret.UpdatedAt.Format("2006-01-02 15:04"))
	}

	return nil
}

// Get выводит конкретные данные
func (d *DataCommands) Get(id string) error {
	secret := d.manager.GetSecretByID(id)
	if secret == nil {
		return fmt.Errorf("data with ID %s not found", id)
	}

	switch secret.Type {
	case common.LoginPassword:
		data, err := d.manager.GetLoginPassword(id)
		if err != nil {
			return err
		}
		fmt.Printf("Login: %s\nPassword: %s\nSite: %s\n", data.Login, data.Password, data.Site)

	case common.Card:
		data, err := d.manager.GetCardData(id)
		if err != nil {
			return err
		}
		fmt.Printf("Number: %s\nExpiry: %s\nHolder: %s\nBank: %s\n", data.Number, data.Expiry, data.Holder, data.Bank)

	case common.TextData:
		data, err := d.manager.GetTextData(id)
		if err != nil {
			return err
		}
		fmt.Printf("Text: %s\n", data)

	case common.BinaryData:
		data, err := d.manager.GetBinaryData(id)
		if err != nil {
			return err
		}
		fmt.Printf("Binary data: %d bytes\n", len(data))
	}

	return nil
}
