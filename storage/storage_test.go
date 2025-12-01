package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorageErrors(t *testing.T) {
	err := &StorageError{msg: "test error"}
	assert.Equal(t, "test error", err.Error())

	assert.Equal(t, "user not found", ErrUserNotFound.Error())
	assert.Equal(t, "secret not found", ErrSecretNotFound.Error())
	assert.Equal(t, "file not found", ErrFileNotFound.Error())
}

func TestStorageInterfaces(_ *testing.T) {
	var _ Storage = NewMemoryStorage()
	var _ FileStorageChecker = NewMemoryStorage()
	var _ FileStorage = (*MemoryStorage)(nil)
}
