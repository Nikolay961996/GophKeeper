package commands

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"path/filepath"
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
	binDir     string
}

// NewDataCommands создает новый DataCommands
func NewDataCommands(cfg *config.Config, dataManager *manager.DataManager, grpcClient *grpc.GRPCClient) *DataCommands {
	homeDir, err := os.UserHomeDir()
	if err != nil {

		log.Fatalf(err.Error())
	}
	binDir := filepath.Join(homeDir, ".gophkeeper", "bin")
	if err := os.MkdirAll(binDir, 0700); err != nil {
		log.Fatalf(err.Error())
	}

	return &DataCommands{
		cfg:        cfg,
		manager:    dataManager,
		grpcClient: grpcClient,
		resolver:   conflict.NewResolver(),
		binDir:     binDir,
	}
}

// Sync синхронизирует данные с сервером через gRPC
func (d *DataCommands) Sync() error {
	if d.cfg.Token == "" {
		return fmt.Errorf("not authenticated. Please login first")
	}

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

	localSecretsData := make([]common.SecretData, len(localSecrets))
	for i, secret := range localSecrets {
		localSecretsData[i] = *secret
	}

	// Выполняем синхронизацию через gRPC
	syncResult, err := d.grpcClient.Sync(lastSync, localSecretsData)
	if err != nil {
		return err
	}

	// Обнаруживаем конфликты
	conflictManager := common.NewConflictManager()
	detectedConflicts := conflictManager.DetectConflicts(localSecretsData, syncResult.Data, lastSync)

	// Обрабатываем конфликты если есть
	if len(detectedConflicts) > 0 {
		fmt.Printf("\n🚨 Found %d conflicts during sync!\n", len(detectedConflicts))
		resolutions, err := d.resolver.ResolveConflicts(detectedConflicts)
		if err != nil {
			return fmt.Errorf("error resolving conflicts: %v", err)
		}

		d.applyResolutions(resolutions)
	}

	// Сохраняем неконфликтные данные с сервера
	nonConflictData := d.filterNonConflictData(syncResult.Data, detectedConflicts)
	for _, serverSecret := range nonConflictData {
		if err := d.manager.SaveSecret(&serverSecret); err != nil {
			fmt.Printf("Warning: failed to save secret %s: %v\n", serverSecret.ID, err)
		}
	}

	err = d.SyncFiles(localSecretsData, syncResult.Data)
	if err != nil {
		return err
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
	fileName := filepath.Base(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	if err := d.manager.SaveBinaryData(name, data, fileName); err != nil {
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

// Get выводит конкретные данные
func (d *DataCommands) Get(id string) error {
	secret := d.manager.GetSecretByID(id)
	if secret == nil {
		return fmt.Errorf("data with ID %s not found", id)
	}

	return d.printSecret(secret)
}

// Delete удаляет данные
func (d *DataCommands) Delete(id string) error {
	return d.manager.DeleteData(id)
}

func (d *DataCommands) SyncFiles(local []common.SecretData, origin []common.SecretData) error {
	var localBins = map[uuid.UUID]bool{}
	for _, ld := range local {
		if ld.Type == common.BinaryDataType {
			localBins[ld.ID] = true
		}
	}

	var originBins = map[uuid.UUID]bool{}
	for _, ld := range origin {
		if ld.Type == common.BinaryDataType {
			originBins[ld.ID] = true
		}
	}

	for _, sd := range origin {
		if sd.Type == common.BinaryDataType {
			if _, ok := localBins[sd.ID]; !ok {
				err := d.grpcClient.DownloadFile(sd.ID.String(), filepath.Join(d.binDir, sd.ID.String()))
				if err != nil {
					log.Printf("Warning: failed to download file %s: %v\n", sd.ID.String(), err)
				}
			}
		}
	}

	for _, sd := range local {
		if sd.Type == common.BinaryDataType {
			if _, ok := originBins[sd.ID]; !ok || sd.Type == common.TextDataType {
				err := d.grpcClient.UploadFile(filepath.Join(d.binDir, sd.ID.String()), sd.ID.String())
				if err != nil {
					log.Printf("Warning: failed to upload file %s: %v\n", sd.ID.String(), err)
				}
			}
		}
	}

	return nil
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
		fileName, data, err := d.manager.GetBinaryData(secret.ID.String())
		if err != nil {
			return err
		}

		fullFileName := fmt.Sprintf("%s-%s", secret.ID.String(), fileName)
		err = os.WriteFile(fullFileName, data, 0644)
		if err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			return err
		}

		fmt.Printf("File %s succeccfully saved\n", fullFileName)
	}

	return nil
}

// handleConflicts обрабатывает конфликты автоматически (пока просто логируем)
func (d *DataCommands) handleConflicts(conflicts []common.SecretData) int {
	if len(conflicts) == 0 {
		return 0
	}

	fmt.Printf("Found %d conflicts:\n", len(conflicts))
	for i, c := range conflicts {
		fmt.Printf("%d. %s (v%d) - please resolve manually\n",
			i+1, c.Metadata, c.Version)
	}

	// TODO: Реализовать интерактивное разрешение конфликтов
	// Пока просто используем серверную версию
	for _, c := range conflicts {
		_ = d.manager.SaveSecret(&c)
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
	for _, c := range conflicts {
		conflictIDs[c.SecretID] = true
	}

	var nonConflict []common.SecretData
	for _, secret := range serverData {
		if !conflictIDs[secret.ID] {
			nonConflict = append(nonConflict, secret)
		}
	}

	return nonConflict
}
