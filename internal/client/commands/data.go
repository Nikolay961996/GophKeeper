package commands

import (
	"fmt"
	"github.com/google/uuid"
	"os"
	"time"

	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/conflict"
	"gophkeeper/internal/client/grpc"
	"gophkeeper/internal/client/manager"
	"gophkeeper/internal/common"
)

// DataCommands обработчики команд для работы с данными
type DataCommands struct {
	cfg        *config.Config
	manager    *manager.DataManager
	grpcClient *grpc.GRPCClient
	resolver   *conflict.Resolver
}

// NewDataCommands создает новый DataCommands
func NewDataCommands(cfg *config.Config, dataManager *manager.DataManager, grpcClient *grpc.GRPCClient) *DataCommands {
	return &DataCommands{
		cfg:        cfg,
		manager:    dataManager,
		grpcClient: grpcClient,
		resolver:   conflict.NewResolver(),
	}
}

// Sync синхронизирует данные с сервером через gRPC
func (d *DataCommands) Sync() error {
	if d.cfg.Token == "" {
		return fmt.Errorf("not authenticated. Please login first")
	}

	fmt.Println("1")
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
	fmt.Println("2")

	// Конвертируем []*common.SecretData в []common.SecretData
	localSecretsData := make([]common.SecretData, len(localSecrets))
	for i, secret := range localSecrets {
		localSecretsData[i] = *secret
	}

	fmt.Println("2.1")

	// Выполняем синхронизацию через gRPC
	syncResult, err := d.grpcClient.Sync(lastSync, localSecretsData)
	if err != nil {
		return err
	}

	fmt.Println("3")

	// Обнаруживаем конфликты
	conflictManager := common.NewConflictManager()
	detectedConflicts := conflictManager.DetectConflicts(localSecretsData, syncResult.Data, lastSync)

	fmt.Println("4")

	// Обрабатываем конфликты если есть
	if len(detectedConflicts) > 0 {
		fmt.Printf("\n🚨 Found %d conflicts during sync!\n", len(detectedConflicts))
		resolutions, err := d.resolver.ResolveConflicts(detectedConflicts)
		if err != nil {
			return fmt.Errorf("error resolving conflicts: %v", err)
		}

		d.applyResolutions(resolutions)
	}

	fmt.Println("5")

	// Сохраняем неконфликтные данные с сервера
	nonConflictData := d.filterNonConflictData(syncResult.Data, detectedConflicts)
	for _, serverSecret := range nonConflictData {
		if err := d.manager.SaveSecret(&serverSecret); err != nil {
			fmt.Printf("Warning: failed to save secret %s: %v\n", serverSecret.ID, err)
		}
	}

	fmt.Printf("✅ Sync completed. Received %d items, resolved %d conflicts, send %d items\n",
		len(syncResult.Data), len(detectedConflicts), len(localSecretsData))

	return nil
}

// AddLoginPassword добавляет логин/пароль
func (d *DataCommands) AddLoginPassword(name, login, password, site string) error {
	if err := d.manager.SaveLoginPassword(name, login, password, site); err != nil {
		return err
	}

	fmt.Printf("Login/password '%s' saved successfully\n", name)
	return d.Sync()
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
		fmt.Printf("%d. %s [%s] (%s) - %s\n", i+1, secret.Metadata, secret.Type, secret.ID, secret.UpdatedAt.Format("2006-01-02 15:04"))
	}

	return nil
}

func (d *DataCommands) GetByPosition(pos int64) error {
	secret := d.manager.GetSecretByPosition(pos)
	if secret == nil {
		return fmt.Errorf("data with position %d not found", pos)
	}

	return d.printSecret(secret)
}

// Get выводит конкретные данные
func (d *DataCommands) Get(id string) error {
	secret := d.manager.GetSecretByID(id)
	if secret == nil {
		return fmt.Errorf("data with ID %s not found", id)
	}

	return d.printSecret(secret)
}

func (d *DataCommands) printSecret(secret *common.SecretData) error {
	switch secret.Type {
	case common.LoginPasswordType:
		data, err := d.manager.GetLoginPassword(secret.ID.String())
		if err != nil {
			return err
		}
		fmt.Printf("Login: %s\nPassword: %s\nSite: %s\n", data.Login, data.Password, data.Site)

	case common.CardDataType:
		data, err := d.manager.GetCardData(secret.ID.String())
		if err != nil {
			return err
		}
		fmt.Printf("Number: %s\nExpiry: %s\nHolder: %s\nBank: %s\n", data.Number, data.Expiry, data.Holder, data.Bank)

	case common.TextDataType:
		data, err := d.manager.GetTextData(secret.ID.String())
		if err != nil {
			return err
		}
		fmt.Printf("Text: %s\n", data)

	case common.BinaryDataType:
		data, err := d.manager.GetBinaryData(secret.ID.String())
		if err != nil {
			return err
		}
		fmt.Printf("Binary data: %d bytes\n", len(data))
	}

	return nil
}

// handleConflicts обрабатывает конфликты автоматически (пока просто логируем)
func (d *DataCommands) handleConflicts(conflicts []common.SecretData) int {
	if len(conflicts) == 0 {
		return 0
	}

	fmt.Printf("Found %d conflicts:\n", len(conflicts))
	for i, conflict := range conflicts {
		fmt.Printf("%d. %s (v%d) - please resolve manually\n",
			i+1, conflict.Metadata, conflict.Version)
	}

	// TODO: Реализовать интерактивное разрешение конфликтов
	// Пока просто используем серверную версию
	for _, conflict := range conflicts {
		d.manager.SaveSecret(&conflict)
	}

	return len(conflicts)
}

// applyResolutions применяет разрешения конфликтов
func (d *DataCommands) applyResolutions(resolutions []common.ConflictResolution) {
	appliedCount := 0
	for _, resolution := range resolutions {
		if resolution.Winner != nil {
			if err := d.manager.SaveSecret(resolution.Winner); err != nil {
				fmt.Printf("Warning: failed to apply resolution for conflict %s: %v\n",
					resolution.ConflictID, err)
			} else {
				appliedCount++
			}
		}
	}
	fmt.Printf("Applied %d conflict resolutions\n", appliedCount)
}

// filterNonConflictData фильтрует данные без конфликтов
func (d *DataCommands) filterNonConflictData(serverData []common.SecretData, conflicts []common.Conflict) []common.SecretData {
	conflictIDs := make(map[uuid.UUID]bool)
	for _, conflict := range conflicts {
		conflictIDs[conflict.SecretID] = true
	}

	var nonConflict []common.SecretData
	for _, secret := range serverData {
		if !conflictIDs[secret.ID] {
			nonConflict = append(nonConflict, secret)
		}
	}

	return nonConflict
}
