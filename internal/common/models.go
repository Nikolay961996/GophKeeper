package common

import (
	"github.com/google/uuid"
	"time"
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
	Version   int       `json:"version"` // ← ВАЖНО: для разрешения конфликтов
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DataType представляет тип хранимых данных
type DataType string

const (
	LoginPassword DataType = "login_password"
	TextData      DataType = "text_data"
	BinaryData    DataType = "binary_data"
	CardData      DataType = "card_data" // ← ИСПРАВЛЕНО: было "Card"
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

// ParseUUID парсит строку в UUID
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// MustParseUUID парсит строку в UUID, паникует при ошибке
func MustParseUUID(s string) uuid.UUID {
	return uuid.MustParse(s)
}

// Now возвращает текущее время в UTC
func Now() time.Time {
	return time.Now().UTC()
}

// TokenExpiration время жизни токена
const TokenExpiration = 24 * time.Hour
