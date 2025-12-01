// Package common for server and client
package common

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

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
	CreatedAt    time.Time `json:"created_at"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"-"`
	ID           uuid.UUID `json:"id"`
}

// SecretData представляет защищенные данные пользователя
type SecretData struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Type      DataType  `json:"type"`
	Name      string    `json:"name"`
	Metadata  string    `json:"metadata"`
	Data      []byte    `json:"data"`
	Version   int       `json:"version"`
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
}

// DataType представляет тип хранимых данных
type DataType string

const (
	LoginPasswordType DataType = "login_password"
	TextDataType      DataType = "text_data"
	BinaryDataType    DataType = "binary_data"
	CardDataType      DataType = "card_data"
)

type ContextKey string

const (
	UserIDKey ContextKey = "userID"
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
	Name     string `json:"name"`
	FileName string `json:"file_name"`
	Size     int    `json:"size"`
}

// ParseUUID парсит строку в UUID
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// MustParseUUID парсит строку в UUID, паникует при ошибке
func MustParseUUID(s string) uuid.UUID {
	return uuid.MustParse(s)
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

// GenerateRandomKey генерирует случайный ключ
func GenerateRandomKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
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
