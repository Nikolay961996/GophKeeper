package common

import (
	"time"

	"github.com/google/uuid"
)

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
	LoginPassword DataType = "login_password"
	TextData      DataType = "text_data"
	BinaryData    DataType = "binary_data"
	Card          DataType = "card_data"
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

// AuthRequest запрос на аутентификацию
type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// AuthResponse ответ на аутентификацию
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// SyncRequest запрос на синхронизацию
type SyncRequest struct {
	LastSync time.Time    `json:"last_sync"`
	Data     []SecretData `json:"data"`
}

// SyncResponse ответ на синхронизацию
type SyncResponse struct {
	LastSync  time.Time    `json:"last_sync"`
	Data      []SecretData `json:"data"`
	Conflicts []SecretData `json:"conflicts,omitempty"`
}

// ParseUUID парсит строку в UUID
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// MustParseUUID парсит строку в UUID, паникует при ошибке
func MustParseUUID(s string) uuid.UUID {
	return uuid.MustParse(s)
}
