package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Now возвращает текущее время в UTC
func Now() time.Time {
	return time.Now().UTC()
}

// TokenExpiration время жизни токена
const TokenExpiration = 24 * time.Hour

// User представляет пользователя системы
type User struct {
	ID           uuid.UUID `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// SecretData представляет защищенные данные пользователя
type SecretData struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Type      DataType  `json:"type"`
	Name      string    `json:"name"`
	Data      []byte    `json:"data"` // Зашифрованные данные
	Metadata  string    `json:"metadata"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DataType представляет тип хранимых данных
type DataType string

const (
	LoginPasswordType DataType = "login_password"
	TextDataType      DataType = "text_data"
	BinaryDataType    DataType = "binary_data"
	CardDataType      DataType = "card_data"
)

// LoginPasswordData структура для данных логин/пароль
type LoginPasswordData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Site     string `json:"site,omitempty"`
}

// CardData структура для данных банковской карты
type CardData struct {
	Number string `json:"number"`
	Expiry string `json:"expiry"`
	CVV    string `json:"cvv"`
	Holder string `json:"holder"`
	Bank   string `json:"bank,omitempty"`
}

// BinaryMetaData метаданные бинарного файла
type BinaryMetaData struct {
	Size     int    `json:"size"`
	Name     string `json:"name"`
	FileName string `json:"file_name"`
}

// HasConflictForSecret проверяет есть ли конфликт для конкретного секрета
func (cm *ConflictManager) HasConflictForSecret(secretID uuid.UUID) bool {
	for _, conflict := range cm.pendingConflicts {
		if conflict.SecretID == secretID {
			return true
		}
	}
	return false
}

// RemoveConflict удаляет конфликт
func (cm *ConflictManager) RemoveConflict(conflictID string) {
	delete(cm.pendingConflicts, conflictID)
}

// GetConflictBySecretID возвращает конфликт по ID секрета
func (cm *ConflictManager) GetConflictBySecretID(secretID uuid.UUID) *Conflict {
	for _, conflict := range cm.pendingConflicts {
		if conflict.SecretID == secretID {
			return conflict
		}
	}
	return nil
}

// ParseUUID парсит строку в UUID
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// MustParseUUID парсит строку в UUID, паникует при ошибке
func MustParseUUID(s string) uuid.UUID {
	return uuid.MustParse(s)
}

// JWTClaims кастомные claims для JWT
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Login  string    `json:"login"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidToken = errors.New("invalid token")
)

// GenerateJWTToken генерирует JWT токен
func GenerateJWTToken(userID uuid.UUID, login, secret string, expiration time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWTToken валидирует JWT токен
func ValidateJWTToken(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func GetJWTClaims(tokenString string) (jwt.MapClaims, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		fmt.Printf("Ошибка парсинга токена: %v\n", err)
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Printf("Не удалось извлечь claims из токена: %v\n", err)
		return nil, err
	}

	return claims, nil
}

// HashPassword создает хеш пароля
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash проверяет пароль с хешем
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateEncryptionKey генерирует ключ шифрования из пароля
func GenerateEncryptionKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

// EncryptData шифрует данные с использованием AES-GCM
func EncryptData(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// DecryptData расшифровывает данные
func DecryptData(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// GenerateRandomKey генерирует случайный ключ
func GenerateRandomKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ConflictType тип конфликта
type ConflictType string

const (
	ConflictBothModified ConflictType = "both_modified" // Обе версии изменены
	ConflictDeleted      ConflictType = "deleted"       // Одна версия удалена
	ConflictVersion      ConflictType = "version"       // Конфликт версий
)

// Conflict представляет конфликт данных
type Conflict struct {
	ID           string       `json:"id"`
	SecretID     uuid.UUID    `json:"secret_id"`
	Type         ConflictType `json:"type"`
	LocalSecret  *SecretData  `json:"local_secret,omitempty"`  // Локальная версия
	RemoteSecret *SecretData  `json:"remote_secret,omitempty"` // Серверная версия
	Reason       string       `json:"reason"`
	DetectedAt   time.Time    `json:"detected_at"`
}

// ConflictResolution разрешение конфликта
type ConflictResolution struct {
	ConflictID string      `json:"conflict_id"`
	Winner     *SecretData `json:"winner"` // Какая версия побеждает
	Action     string      `json:"action"` // keep_local, keep_remote, merge, cancel
	ResolvedAt time.Time   `json:"resolved_at"`
}

// ConflictManager управляет конфликтами
type ConflictManager struct {
	pendingConflicts map[string]*Conflict
}

// NewConflictManager создает новый менеджер конфликтов
func NewConflictManager() *ConflictManager {
	return &ConflictManager{
		pendingConflicts: make(map[string]*Conflict),
	}
}

// DetectConflicts обнаруживает конфликты между локальными и серверными данными
func (cm *ConflictManager) DetectConflicts(localSecrets, remoteSecrets []SecretData, lastSync time.Time) []Conflict {
	var conflicts []Conflict

	localMap := make(map[uuid.UUID]SecretData)
	for _, secret := range localSecrets {
		localMap[secret.ID] = secret
	}

	remoteMap := make(map[uuid.UUID]SecretData)
	for _, secret := range remoteSecrets {
		remoteMap[secret.ID] = secret
	}

	// Проверяем конфликты для существующих записей
	for id, local := range localMap {
		remote, exists := remoteMap[id]
		if !exists {
			continue
		}

		// Обе версии изменялись после последней синхронизации
		if local.UpdatedAt.After(lastSync) && remote.UpdatedAt.After(lastSync) &&
			local.Version != remote.Version {

			conflict := Conflict{
				ID:           fmt.Sprintf("%s_%d_%d", id, local.Version, remote.Version),
				SecretID:     id,
				Type:         ConflictBothModified,
				LocalSecret:  &local,
				RemoteSecret: &remote,
				Reason:       fmt.Sprintf("Both versions modified after last sync (local v%d, remote v%d)", local.Version, remote.Version),
				DetectedAt:   Now(),
			}
			conflicts = append(conflicts, conflict)
		}
	}

	// Проверяем удаленные записи
	for id, remote := range remoteMap {
		_, exists := localMap[id]
		if !exists && remote.UpdatedAt.After(lastSync) {
			// Запись удалена локально, но изменена на сервере
			conflict := Conflict{
				ID:           fmt.Sprintf("%s_deleted", id),
				SecretID:     id,
				Type:         ConflictDeleted,
				LocalSecret:  nil, // Локально удалено
				RemoteSecret: &remote,
				Reason:       "Locally deleted but modified on server",
				DetectedAt:   Now(),
			}
			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts
}

// AddConflict добавляет конфликт в pending
func (cm *ConflictManager) AddConflict(conflict Conflict) {
	cm.pendingConflicts[conflict.ID] = &conflict
}

// ResolveConflict разрешает конфликт
func (cm *ConflictManager) ResolveConflict(conflictID string, action string) (*ConflictResolution, error) {
	conflict, exists := cm.pendingConflicts[conflictID]
	if !exists {
		return nil, fmt.Errorf("conflict not found: %s", conflictID)
	}

	var winner *SecretData
	switch action {
	case "keep_local":
		winner = conflict.LocalSecret
	case "keep_remote":
		winner = conflict.RemoteSecret
	case "keep_newer":
		if conflict.LocalSecret.UpdatedAt.After(conflict.RemoteSecret.UpdatedAt) {
			winner = conflict.LocalSecret
		} else {
			winner = conflict.RemoteSecret
		}
	case "cancel":
		// Отмена - не сохранять ничего
		winner = nil
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}

	resolution := &ConflictResolution{
		ConflictID: conflictID,
		Winner:     winner,
		Action:     action,
		ResolvedAt: Now(),
	}

	// Удаляем из pending
	delete(cm.pendingConflicts, conflictID)

	return resolution, nil
}

// GetPendingConflicts возвращает pending конфликты
func (cm *ConflictManager) GetPendingConflicts() []Conflict {
	var conflicts []Conflict
	for _, conflict := range cm.pendingConflicts {
		conflicts = append(conflicts, *conflict)
	}
	return conflicts
}

// HasPendingConflicts проверяет есть ли pending конфликты
func (cm *ConflictManager) HasPendingConflicts() bool {
	return len(cm.pendingConflicts) > 0
}

// CompareSecrets сравнивает два секрета (для отображения различий)
func CompareSecrets(local, remote *SecretData) string {
	var differences []string

	if local.Metadata != remote.Metadata {
		differences = append(differences, fmt.Sprintf("Name: '%s' vs '%s'", local.Metadata, remote.Metadata))
	}

	if local.Version != remote.Version {
		differences = append(differences, fmt.Sprintf("Version: %d vs %d", local.Version, remote.Version))
	}

	if !local.UpdatedAt.Equal(remote.UpdatedAt) {
		differences = append(differences, fmt.Sprintf("Updated: %s vs %s",
			local.UpdatedAt.Format("2006-01-02 15:04"),
			remote.UpdatedAt.Format("2006-01-02 15:04")))
	}

	if len(differences) == 0 {
		return "No visible differences (encrypted content may differ)"
	}

	result := "Differences:\n"
	for i, diff := range differences {
		result += fmt.Sprintf("  %d. %s\n", i+1, diff)
	}
	return result
}

// OperationType тип операции
type OperationType string

const (
	OpRegister OperationType = "register"
	OpLogin    OperationType = "login"
	OpSync     OperationType = "sync"
)

// OperationRequest запрос операции
type OperationRequest struct {
	Type    OperationType `json:"type"`
	Payload []byte        `json:"payload"`
}

// OperationResponse ответ операции
type OperationResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Payload []byte `json:"payload,omitempty"`
}

// RegisterOp данные регистрации
type RegisterOp struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// LoginOp данные входа
type LoginOp struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// SyncOp данные синхронизации
type SyncOp struct {
	LastSync time.Time    `json:"last_sync"`
	Data     []SecretData `json:"data"`
}

// SyncResult результат синхронизации
type SyncResult struct {
	LastSync  time.Time    `json:"last_sync"`
	Data      []SecretData `json:"data"`
	Conflicts []SecretData `json:"conflicts,omitempty"`
}

// AuthResult результат аутентификации
type AuthResult struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// MarshalOperation маршалит операцию в JSON
func MarshalOperation(op interface{}) ([]byte, error) {
	return json.Marshal(op)
}

// UnmarshalOperation анмаршалит операцию из JSON
func UnmarshalOperation(data []byte, op interface{}) error {
	return json.Unmarshal(data, op)
}

// CreateOperationRequest создает запрос операции
func CreateOperationRequest(opType OperationType, payload interface{}) (*OperationRequest, error) {
	payloadData, err := MarshalOperation(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	return &OperationRequest{
		Type:    opType,
		Payload: payloadData,
	}, nil
}

// CreateSuccessResponse создает успешный ответ
func CreateSuccessResponse(payload interface{}) (*OperationResponse, error) {
	var payloadData []byte
	if payload != nil {
		var err error
		payloadData, err = MarshalOperation(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %v", err)
		}
	}

	return &OperationResponse{
		Success: true,
		Payload: payloadData,
	}, nil
}

// CreateErrorResponse создает ответ с ошибкой
func CreateErrorResponse(err error) *OperationResponse {
	return &OperationResponse{
		Success: false,
		Error:   err.Error(),
	}
}
