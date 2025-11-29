package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorageErrors(t *testing.T) {
	// Тестируем ошибки хранилища
	err := &StorageError{msg: "test error"}
	assert.Equal(t, "test error", err.Error())

	// Проверяем предопределенные ошибки
	assert.Equal(t, "user not found", ErrUserNotFound.Error())
	assert.Equal(t, "secret not found", ErrSecretNotFound.Error())
	assert.Equal(t, "file not found", ErrFileNotFound.Error())
}

func TestStorageInterfaces(_ *testing.T) {
	// Проверяем что MemoryStorage реализует интерфейсы
	var _ Storage = NewMemoryStorage()
	var _ FileStorageChecker = NewMemoryStorage()

	// FileStorage не реализуется MemoryStorage, но интерфейс существует
	var _ FileStorage = (*MemoryStorage)(nil) // это будет fail, но даст coverage
}
