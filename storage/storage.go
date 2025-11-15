package storage

import (
	"time"

	"github.com/google/uuid"
	"gophkeeper/internal/common"
)

// Storage интерфейс определяет методы для работы с хранилищем
type Storage interface {
	// User methods
	CreateUser(user *common.User) error
	GetUserByLogin(login string) (*common.User, error)
	GetUserByID(id uuid.UUID) (*common.User, error)

	// Secret data methods
	SaveSecretData(data *common.SecretData) error
	GetUserSecrets(userID uuid.UUID) ([]*common.SecretData, error)
	GetSecretsSince(userID uuid.UUID, since time.Time) ([]*common.SecretData, error)
	DeleteSecret(userID, secretID uuid.UUID) error
}

// StorageError представляет ошибку хранилища
type StorageError struct {
	msg string
}

func (e *StorageError) Error() string {
	return e.msg
}

// Predefined storage errors
var (
	ErrUserNotFound   = &StorageError{"user not found"}
	ErrSecretNotFound = &StorageError{"secret not found"}
)
