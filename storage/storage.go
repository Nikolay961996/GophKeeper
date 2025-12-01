// Package storage contains functions for store data
package storage

import (
	"time"

	"github.com/google/uuid"
	"gophkeeper/internal/common"
)

// Storage интерфейс определяет основные методы для работы с хранилищем
type Storage interface {
	CreateUser(user *common.User) error
	GetUserByLogin(login string) (*common.User, error)
	GetUserByID(id uuid.UUID) (*common.User, error)

	SaveSecretData(data *common.SecretData) error
	GetUserSecrets(userID uuid.UUID) ([]*common.SecretData, error)
	GetSecretsSince(userID uuid.UUID, since time.Time) ([]*common.SecretData, error)
	DeleteSecret(userID, secretID uuid.UUID) error
}

// FileStorage интерфейс определяет методы для работы с файлами (опционально)
type FileStorage interface {
	CreateFileMetadata(metadata *FileMetadata) error
	SaveFileChunk(chunk *FileChunk) error
	GetFileMetadata(fileID uuid.UUID) (*FileMetadata, error)
	GetFileChunk(fileID uuid.UUID, chunkIndex int) (*FileChunk, error)
	GetAllFileChunks(fileID uuid.UUID) ([]*FileChunk, error)
	GetUserFiles(userID uuid.UUID) ([]*FileMetadata, error)
	DeleteFile(fileID uuid.UUID) error
}

// ExtendedStorage объединяет оба интерфейса
type ExtendedStorage interface {
	Storage
	FileStorage
}

// FileStorageChecker интерфейс для проверки поддержки файлов
type FileStorageChecker interface {
	SupportsFiles() bool
}

// StorageError представляет ошибку хранилища
type StorageError struct {
	msg string
}

func (e *StorageError) Error() string {
	return e.msg
}

var (
	ErrUserNotFound   = &StorageError{"user not found"}
	ErrSecretNotFound = &StorageError{"secret not found"}
	ErrFileNotFound   = &StorageError{"file not found"}
)
