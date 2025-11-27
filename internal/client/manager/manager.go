package manager

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"path/filepath"
	"sync"

	"gophkeeper/internal/client/config"
	"gophkeeper/internal/client/crypto"
	"gophkeeper/internal/common"
)

// DataManager управляет данными на клиенте
type DataManager struct {
	cfg       *config.Config
	crypto    *crypto.ClientCrypto
	localData map[string]*common.SecretData // ID -> SecretData
	mu        sync.RWMutex
	dataFile  string
}

// NewDataManager создает новый DataManager
func NewDataManager(cfg *config.Config, masterPassword string) (*DataManager, error) {
	dataFile, err := getDataFilePath()
	if err != nil {
		return nil, err
	}

	manager := &DataManager{
		cfg:       cfg,
		crypto:    crypto.NewClientCrypto(masterPassword),
		localData: make(map[string]*common.SecretData),
		dataFile:  dataFile,
	}

	// Загружаем локальные данные
	if err := manager.loadLocalData(); err != nil {
		fmt.Printf("Warning: could not load local data: %v\n", err)
	}

	return manager, nil
}

// SaveLoginPassword сохраняет логин/пароль
func (m *DataManager) SaveLoginPassword(name, login, password, site string) error {
	data := common.LoginPasswordData{
		Login:    login,
		Password: password,
		Site:     site,
	}

	secret, err := m.crypto.EncryptData(common.LoginPasswordType, data, name)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.localData[secret.ID.String()] = secret

	return m.saveLocalData()
}

// SaveCardData сохраняет данные банковской карты
func (m *DataManager) SaveCardData(name, number, expiry, cvv, holder, bank string) error {
	data := common.CardData{
		Number: number,
		Expiry: expiry,
		CVV:    cvv,
		Holder: holder,
		Bank:   bank,
	}

	secret, err := m.crypto.EncryptData(common.CardDataType, data, name)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.localData[secret.ID.String()] = secret
	return m.saveLocalData()
}

// SaveTextData сохраняет текстовые данные
func (m *DataManager) SaveTextData(name, text string) error {
	secret, err := m.crypto.EncryptData(common.TextDataType, text, name)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.localData[secret.ID.String()] = secret
	return m.saveLocalData()
}

// SaveBinaryData сохраняет бинарные данные
func (m *DataManager) SaveBinaryData(name string, data []byte) error {
	secret, err := m.crypto.EncryptData(common.BinaryDataType, data, name)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.localData[secret.ID.String()] = secret
	return m.saveLocalData()
}

// GetLoginPassword возвращает логин/пароль
func (m *DataManager) GetLoginPassword(id string) (*common.LoginPasswordData, error) {
	m.mu.RLock()
	secret, exists := m.localData[id]
	m.mu.RUnlock()

	// m.cfg.UserID
	if !exists || m.cfg.UserID != secret.UserID.String() {
		return nil, fmt.Errorf("data not found")
	}

	var result common.LoginPasswordData
	if err := m.crypto.DecryptData(secret, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetCardData возвращает данные карты
func (m *DataManager) GetCardData(id string) (*common.CardData, error) {
	m.mu.RLock()
	secret, exists := m.localData[id]
	m.mu.RUnlock()

	if !exists || m.cfg.UserID != secret.UserID.String() {
		return nil, fmt.Errorf("data not found")
	}

	var result common.CardData
	if err := m.crypto.DecryptData(secret, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTextData возвращает текстовые данные
func (m *DataManager) GetTextData(id string) (string, error) {
	m.mu.RLock()
	secret, exists := m.localData[id]
	m.mu.RUnlock()

	if !exists || m.cfg.UserID != secret.UserID.String() {
		return "", fmt.Errorf("data not found")
	}

	var result string
	if err := m.crypto.DecryptData(secret, &result); err != nil {
		return "", err
	}

	return result, nil
}

// GetBinaryData возвращает бинарные данные
func (m *DataManager) GetBinaryData(id string) ([]byte, error) {
	m.mu.RLock()
	secret, exists := m.localData[id]
	m.mu.RUnlock()

	if !exists || m.cfg.UserID != secret.UserID.String() {
		return nil, fmt.Errorf("data not found")
	}

	var result []byte
	if err := m.crypto.DecryptData(secret, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// ListData возвращает список всех данных
func (m *DataManager) ListData() []*common.SecretData {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*common.SecretData
	for _, secret := range m.localData {
		if secret.UserID.String() == m.cfg.UserID {
			result = append(result, secret)
		}
	}
	return result
}

// DeleteData удаляет данные
func (m *DataManager) DeleteData(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.localData[id]; !exists {
		return fmt.Errorf("data not found")
	}

	delete(m.localData, id)
	return m.saveLocalData()
}

// GetSecretByPosition возвращает секрет по позиции
func (m *DataManager) GetSecretByPosition(pos int64) *common.SecretData {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var i int64 = 1
	for _, secret := range m.localData {
		if secret.UserID.String() != m.cfg.UserID {
			continue
		}

		if i == pos {
			return secret
		}
		i++
	}

	return nil
}

// GetSecretByID возвращает секрет по ID
func (m *DataManager) GetSecretByID(id string) *common.SecretData {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.localData[id]
}

// SaveSecret сохраняет секрет (для синхронизации)
func (m *DataManager) SaveSecret(secret *common.SecretData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Проверяем версию для разрешения конфликтов
	existing, exists := m.localData[secret.ID.String()]
	if exists && existing.UpdatedAt.After(secret.UpdatedAt) {
		// Локальная версия новее, пропускаем
		return nil
	}

	m.localData[secret.ID.String()] = secret
	return m.saveLocalData()
}

// loadLocalData загружает локальные данные из файла
func (m *DataManager) loadLocalData() error {
	data, err := os.ReadFile(m.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Файл не существует - это нормально
		}
		return err
	}

	var secrets []*common.SecretData
	if err := json.Unmarshal(data, &secrets); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, secret := range secrets {
		m.localData[secret.ID.String()] = secret
	}

	return nil
}

// saveLocalData сохраняет локальные данные в файл
func (m *DataManager) saveLocalData() error {
	var secrets []*common.SecretData
	for _, secret := range m.localData {
		if uuid.Nil == secret.UserID {
			userID, err := uuid.Parse(m.cfg.UserID)
			if err != nil {
				return err
			}
			secret.UserID = userID
		}
		secrets = append(secrets, secret)
	}

	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(m.dataFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	log.Printf("Success saved to local")
	return os.WriteFile(m.dataFile, data, 0600)
}

// getDataFilePath возвращает путь к файлу данных
func getDataFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".gophkeeper", "data.json"), nil
}
