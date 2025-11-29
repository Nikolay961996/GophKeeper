package commands

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"path/filepath"
	"strings"
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
	manager    manager.ManagerInterface
	grpcClient grpc.ClientInterface
	resolver   *conflict.Resolver
	binDir     string
}

// NewDataCommands создает новый DataCommands
func NewDataCommands(cfg *config.Config, dataManager manager.ManagerInterface, grpcClient grpc.ClientInterface) *DataCommands {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err.Error())
	}
	binDir := filepath.Join(homeDir, ".gophkeeper", "bin")
	if err = os.MkdirAll(binDir, 0700); err != nil {
		log.Fatalln(err.Error())
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
	resolvedCount := 0
	if len(detectedConflicts) > 0 {
		fmt.Printf("\n⚠️ Found %d conflicts during sync!\n", len(detectedConflicts))

		// Интерактивное разрешение конфликтов
		resolutions, e := d.resolveConflictsInteractively(detectedConflicts)
		if e != nil {
			return fmt.Errorf("error resolving conflicts: %v", e)
		}

		// Применяем разрешения
		resolvedCount = d.applyConflictResolutions(resolutions)

		fmt.Printf("✅ Resolved %d out of %d conflicts\n", resolvedCount, len(detectedConflicts))
	}

	// Сохраняем неконфликтные данные с сервера
	nonConflictData := d.filterNonConflictData(syncResult.Data, detectedConflicts)
	serverItemsSaved := 0
	for _, serverSecret := range nonConflictData {
		if err = d.manager.SaveSecret(&serverSecret); err != nil {
			fmt.Printf("Warning: failed to save secret %s: %v\n", serverSecret.ID, err)
		} else {
			serverItemsSaved++
		}
	}

	// Отправляем локальные данные на сервер (кроме тех, что были в конфликтах)
	localItemsToSend := d.filterLocalDataForSync(localSecretsData, detectedConflicts)
	if len(localItemsToSend) > 0 {
		// Для простоты отправляем все локальные данные заново
		// В реальной реализации здесь была бы более сложная логика
		_, err = d.grpcClient.Sync(lastSync, localItemsToSend)
		if err != nil {
			fmt.Printf("Warning: failed to send local changes to server: %v\n", err)
		}
	}

	err = d.SyncFiles(localSecretsData, syncResult.Data)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Sync completed. Received %d items, resolved %d conflicts, sent %d items\n",
		len(syncResult.Data), resolvedCount, len(localItemsToSend))

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

// SyncFiles синхронизирует файлы между клиентом и сервером
func (d *DataCommands) SyncFiles(local []common.SecretData, server []common.SecretData) error {
	localBins := make(map[uuid.UUID]*common.SecretData)
	serverBins := make(map[uuid.UUID]*common.SecretData)

	for i := range local {
		fmt.Printf("Local: %s - %s\n", local[i].ID, local[i].Type)
		if local[i].Type == common.BinaryDataType {
			localBins[local[i].ID] = &local[i]
		}
	}

	for i := range server {
		fmt.Printf("Server: %s - %s\n", server[i].ID, server[i].Type)
		if server[i].Type == common.BinaryDataType {
			serverBins[server[i].ID] = &server[i]
		}
	}

	fmt.Printf("File sync: %d local files, %d server files\n", len(localBins), len(serverBins))

	if len(serverBins) == 0 && len(localBins) > 0 {
		fmt.Printf("WARNING: Server returned 0 files but claims files exist. Metadata sync issue?\n")
	}

	// Скачиваем отсутствующие файлы с сервера
	downloadCount := 0
	for id := range serverBins {
		localFileExists := d.checkLocalFileExists(id)

		if !localFileExists {
			fmt.Printf("Downloading file %s from server...\n", id)
			err := d.grpcClient.DownloadFile(id.String(), filepath.Join(d.binDir, id.String()))
			if err != nil {
				log.Printf("Failed to download file %s: %v\n", id, err)
			} else {
				downloadCount++
			}
		}
	}

	// Загружаем на сервер ТОЛЬКО если файла действительно нет на сервере
	uploadCount := 0
	for id := range localBins {
		_, existsOnServer := serverBins[id]
		localFileExists := d.checkLocalFileExists(id)

		if !existsOnServer && localFileExists {
			fmt.Printf("Attempting to upload new file %s to server...\n", id)
			err := d.grpcClient.UploadFile(filepath.Join(d.binDir, id.String()), id.String())
			if err != nil {
				if strings.Contains(err.Error(), "duplicate key") {
					fmt.Printf("File %s already exists on server (duplicate key)\n", id)
				} else if strings.Contains(err.Error(), "EOF") {
					log.Printf("Connection error uploading file %s: %v\n", id, err)
				} else {
					log.Printf("Failed to upload file %s: %v\n", id, err)
				}
			} else {
				uploadCount++
				fmt.Printf("Successfully uploaded file %s\n", id)
			}
		} else if existsOnServer {
			fmt.Printf("File %s already on server\n", id)
		}
	}

	fmt.Printf("File sync completed: downloaded %d, uploaded %d\n", downloadCount, uploadCount)
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

// checkLocalFileExists проверяет существует ли локальный файл
func (d *DataCommands) checkLocalFileExists(fileID uuid.UUID) bool {
	filePath := filepath.Join(d.binDir, fileID.String())
	_, err := os.Stat(filePath)
	return err == nil
}

//// getLocalFileModTime возвращает время модификации локального файла
//func (d *DataCommands) getLocalFileModTime(fileID uuid.UUID) time.Time {
//	filePath := filepath.Join(d.binDir, fileID.String())
//	info, err := os.Stat(filePath)
//	if err != nil {
//		return time.Time{}
//	}
//	return info.ModTime()
//}

/*
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
*/

//// applyResolutions применяет разрешения конфликтов
//func (d *DataCommands) applyResolutions(resolutions []common.ConflictResolution) {
//	appliedCount := 0
//	for _, resolution := range resolutions {
//		if resolution.Winner != nil {
//			if err := d.manager.SaveSecret(resolution.Winner); err != nil {
//				fmt.Printf("Warning: failed to apply resolution for conflict %s: %v\n",
//					resolution.ConflictID, err)
//			} else {
//				appliedCount++
//			}
//		}
//	}
//	fmt.Printf("Applied %d conflict resolutions\n", appliedCount)
//}

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

// resolveConflictsInteractively интерактивно разрешает конфликты
func (d *DataCommands) resolveConflictsInteractively(conflicts []common.Conflict) ([]common.ConflictResolution, error) {
	var resolutions []common.ConflictResolution

	for i, c := range conflicts {
		fmt.Printf("\n=== Conflict %d/%d: %s ===\n", i+1, len(conflicts), c.Reason)

		resolution, err := d.promptConflictResolution(c)
		if err != nil {
			return nil, err
		}

		if resolution != nil {
			resolutions = append(resolutions, *resolution)
		}
	}

	return resolutions, nil
}

// promptConflictResolution запрашивает у пользователя как разрешить конфликт
func (d *DataCommands) promptConflictResolution(conflict common.Conflict) (*common.ConflictResolution, error) {
	// Показываем информацию о конфликте
	d.displayConflictDetails(conflict)

	// Предлагаем варианты разрешения
	for {
		fmt.Println("\nHow would you like to resolve this conflict?")
		fmt.Println("1. Keep local version")
		fmt.Println("2. Keep server version")
		fmt.Println("3. Keep newer version (by modification time)")
		fmt.Println("4. Skip this conflict for now")

		if conflict.LocalSecret != nil && conflict.RemoteSecret != nil {
			fmt.Println("5. Show detailed comparison")
		}

		fmt.Print("Choose option (1-5): ")

		var choice int
		_, err := fmt.Scanln(&choice)
		if err != nil {
			return nil, err
		}

		switch choice {
		case 1: // Keep local
			if conflict.LocalSecret == nil {
				fmt.Println("Cannot keep local - local version is deleted")
				continue
			}
			return &common.ConflictResolution{
				ConflictID: conflict.ID,
				Winner:     conflict.LocalSecret,
				Action:     "keep_local",
				ResolvedAt: time.Now(),
			}, nil

		case 2: // Keep server
			if conflict.RemoteSecret == nil {
				fmt.Println("Cannot keep server - server version is deleted")
				continue
			}
			return &common.ConflictResolution{
				ConflictID: conflict.ID,
				Winner:     conflict.RemoteSecret,
				Action:     "keep_remote",
				ResolvedAt: time.Now(),
			}, nil

		case 3: // Keep newer
			if conflict.LocalSecret != nil && conflict.RemoteSecret != nil {
				var winner *common.SecretData
				if conflict.LocalSecret.UpdatedAt.After(conflict.RemoteSecret.UpdatedAt) {
					winner = conflict.LocalSecret
					fmt.Println("Keeping local version (newer)")
				} else {
					winner = conflict.RemoteSecret
					fmt.Println("Keeping server version (newer)")
				}
				return &common.ConflictResolution{
					ConflictID: conflict.ID,
					Winner:     winner,
					Action:     "keep_newer",
					ResolvedAt: time.Now(),
				}, nil
			} else {
				fmt.Println("Cannot compare versions - one side is deleted")
				continue
			}

		case 4: // Skip
			fmt.Println("Skipping this conflict")
			return nil, nil

		case 5: // Show details
			if conflict.LocalSecret != nil && conflict.RemoteSecret != nil {
				d.showDetailedComparison(conflict)
			} else {
				fmt.Println("Detailed comparison not available")
			}
			continue

		default:
			fmt.Println("Invalid option, please try again")
			continue
		}
	}
}

// displayConflictDetails показывает детали конфликта
func (d *DataCommands) displayConflictDetails(conflict common.Conflict) {
	fmt.Printf("Secret: %s\n", conflict.SecretID)
	fmt.Printf("Type: %s\n", conflict.Type)
	fmt.Printf("Reason: %s\n", conflict.Reason)

	if conflict.LocalSecret != nil {
		fmt.Printf("Local version: v%d, updated: %s\n",
			conflict.LocalSecret.Version,
			conflict.LocalSecret.UpdatedAt.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("Local version: DELETED")
	}

	if conflict.RemoteSecret != nil {
		fmt.Printf("Server version: v%d, updated: %s\n",
			conflict.RemoteSecret.Version,
			conflict.RemoteSecret.UpdatedAt.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Println("Server version: DELETED")
	}
}

// showDetailedComparison показывает детальное сравнение
func (d *DataCommands) showDetailedComparison(conflict common.Conflict) {
	fmt.Println("\n--- Detailed Comparison ---")
	fmt.Printf("%-20s | %-20s | %-20s\n", "FIELD", "LOCAL", "SERVER")
	fmt.Printf("%-20s | %-20s | %-20s\n", "--------------------", "--------------------", "--------------------")

	fmt.Printf("%-20s | %-20s | %-20s\n",
		"Name",
		conflict.LocalSecret.Metadata,
		conflict.RemoteSecret.Metadata)

	fmt.Printf("%-20s | %-20s | %-20s\n",
		"Type",
		string(conflict.LocalSecret.Type),
		string(conflict.RemoteSecret.Type))

	fmt.Printf("%-20s | %-20d | %-20d\n",
		"Version",
		conflict.LocalSecret.Version,
		conflict.RemoteSecret.Version)

	fmt.Printf("%-20s | %-20s | %-20s\n",
		"Last Modified",
		conflict.LocalSecret.UpdatedAt.Format("01/02 15:04"),
		conflict.RemoteSecret.UpdatedAt.Format("01/02 15:04"))

	fmt.Printf("%-20s | %-20d | %-20d\n",
		"Data Size",
		len(conflict.LocalSecret.Data),
		len(conflict.RemoteSecret.Data))
}

// applyConflictResolutions применяет разрешения конфликтов
func (d *DataCommands) applyConflictResolutions(resolutions []common.ConflictResolution) int {
	appliedCount := 0

	for _, resolution := range resolutions {
		if resolution.Winner != nil {
			if err := d.manager.SaveSecret(resolution.Winner); err != nil {
				fmt.Printf("Warning: failed to apply resolution for conflict %s: %v\n",
					resolution.ConflictID, err)
			} else {
				appliedCount++
				fmt.Printf("Applied resolution for conflict %s: %s\n",
					resolution.ConflictID, resolution.Action)
			}
		}
	}

	return appliedCount
}

// filterLocalDataForSync фильтрует локальные данные для отправки на сервер
func (d *DataCommands) filterLocalDataForSync(localData []common.SecretData, conflicts []common.Conflict) []common.SecretData {
	conflictSecretIDs := make(map[uuid.UUID]bool)
	for _, c := range conflicts {
		conflictSecretIDs[c.SecretID] = true
	}

	var filtered []common.SecretData
	for _, secret := range localData {
		if !conflictSecretIDs[secret.ID] {
			filtered = append(filtered, secret)
		}
	}

	return filtered
}
