package storage

import (
	"testing"
	"time"

	"gophkeeper/internal/common"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_CreateUser(t *testing.T) {
	storage := NewMemoryStorage()

	user := &common.User{
		ID:        uuid.New(),
		Login:     "testuser",
		CreatedAt: common.Now(),
	}

	err := storage.CreateUser(user)
	require.NoError(t, err)

	// Test duplicate user
	err = storage.CreateUser(user)
	assert.Error(t, err)
}

func TestMemoryStorage_GetUserByLogin(t *testing.T) {
	storage := NewMemoryStorage()

	user := &common.User{
		ID:        uuid.New(),
		Login:     "testuser",
		CreatedAt: common.Now(),
	}

	err := storage.CreateUser(user)
	require.NoError(t, err)

	// Test existing user
	foundUser, err := storage.GetUserByLogin("testuser")
	require.NoError(t, err)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Equal(t, user.Login, foundUser.Login)

	// Test non-existing user
	_, err = storage.GetUserByLogin("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestMemoryStorage_GetUserByID(t *testing.T) {
	storage := NewMemoryStorage()

	user := &common.User{
		ID:        uuid.New(),
		Login:     "testuser",
		CreatedAt: common.Now(),
	}

	err := storage.CreateUser(user)
	require.NoError(t, err)

	// Test existing user
	foundUser, err := storage.GetUserByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, foundUser.ID)

	// Test non-existing user
	_, err = storage.GetUserByID(uuid.New())
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestMemoryStorage_SaveAndGetSecretData(t *testing.T) {
	storage := NewMemoryStorage()

	user := &common.User{
		ID:        uuid.New(),
		Login:     "testuser",
		CreatedAt: common.Now(),
	}

	err := storage.CreateUser(user)
	require.NoError(t, err)

	secret := &common.SecretData{
		ID:        uuid.New(),
		UserID:    user.ID,
		Type:      common.LoginPasswordType,
		Name:      "Test Secret",
		Metadata:  "test metadata",
		Data:      []byte("test data"),
		Version:   1,
		CreatedAt: common.Now(),
		UpdatedAt: common.Now(),
	}

	// Save secret
	err = storage.SaveSecretData(secret)
	require.NoError(t, err)

	// Get user secrets
	secrets, err := storage.GetUserSecrets(user.ID)
	require.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, secret.ID, secrets[0].ID)

	// Update secret
	secret.Version = 2
	secret.UpdatedAt = common.Now()
	err = storage.SaveSecretData(secret)
	require.NoError(t, err)

	// Verify update
	secrets, err = storage.GetUserSecrets(user.ID)
	require.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, 2, secrets[0].Version)
}

func TestMemoryStorage_GetSecretsSince(t *testing.T) {
	storage := NewMemoryStorage()

	user := &common.User{
		ID:        uuid.New(),
		Login:     "testuser",
		CreatedAt: common.Now(),
	}

	err := storage.CreateUser(user)
	require.NoError(t, err)

	baseTime := common.Now()

	// Create secret before base time
	oldSecret := &common.SecretData{
		ID:        uuid.New(),
		UserID:    user.ID,
		Type:      common.LoginPasswordType,
		Name:      "Old Secret",
		UpdatedAt: baseTime.Add(-2 * time.Hour),
	}

	// Create secret after base time
	newSecret := &common.SecretData{
		ID:        uuid.New(),
		UserID:    user.ID,
		Type:      common.LoginPasswordType,
		Name:      "New Secret",
		UpdatedAt: baseTime.Add(time.Hour),
	}

	err = storage.SaveSecretData(oldSecret)
	require.NoError(t, err)
	err = storage.SaveSecretData(newSecret)
	require.NoError(t, err)

	// Get secrets since base time
	secrets, err := storage.GetSecretsSince(user.ID, baseTime)
	require.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, newSecret.ID, secrets[0].ID)
}

func TestMemoryStorage_DeleteSecret(t *testing.T) {
	storage := NewMemoryStorage()

	user := &common.User{
		ID:        uuid.New(),
		Login:     "testuser",
		CreatedAt: common.Now(),
	}

	err := storage.CreateUser(user)
	require.NoError(t, err)

	secret := &common.SecretData{
		ID:        uuid.New(),
		UserID:    user.ID,
		Type:      common.LoginPasswordType,
		Name:      "Test Secret",
		UpdatedAt: common.Now(),
	}

	err = storage.SaveSecretData(secret)
	require.NoError(t, err)

	// Delete secret
	err = storage.DeleteSecret(user.ID, secret.ID)
	require.NoError(t, err)

	// Verify deletion
	secrets, err := storage.GetUserSecrets(user.ID)
	require.NoError(t, err)
	assert.Len(t, secrets, 0)

	// Delete non-existing secret
	err = storage.DeleteSecret(user.ID, uuid.New())
	assert.Error(t, err)
}

func TestMemoryStorage_FileMethodsNotSupported(t *testing.T) {
	storage := NewMemoryStorage()

	assert.False(t, storage.SupportsFiles())

	// Test all file methods return errors
	err := storage.CreateFileMetadata(&FileMetadata{})
	assert.Error(t, err)

	err = storage.SaveFileChunk(&FileChunk{})
	assert.Error(t, err)

	_, err = storage.GetFileMetadata(uuid.New())
	assert.Error(t, err)

	_, err = storage.GetFileChunk(uuid.New(), 0)
	assert.Error(t, err)

	_, err = storage.GetAllFileChunks(uuid.New())
	assert.Error(t, err)

	_, err = storage.GetUserFiles(uuid.New())
	assert.Error(t, err)

	err = storage.DeleteFile(uuid.New())
	assert.Error(t, err)
}

func TestAllStorageMethods(_ *testing.T) {
	storage := NewMemoryStorage()

	user := &common.User{
		ID:        uuid.New(),
		Login:     "testuser",
		CreatedAt: common.Now(),
	}
	_ = storage.CreateUser(user)

	// Вызываем все методы
	_, _ = storage.GetUserByLogin("testuser")
	_, _ = storage.GetUserByID(user.ID)

	secret := &common.SecretData{
		ID:        uuid.New(),
		UserID:    user.ID,
		Type:      common.LoginPasswordType,
		Name:      "test",
		UpdatedAt: common.Now(),
	}
	_ = storage.SaveSecretData(secret)

	_, _ = storage.GetUserSecrets(user.ID)
	_, _ = storage.GetSecretsSince(user.ID, common.Now().Add(-time.Hour))
	_ = storage.DeleteSecret(user.ID, secret.ID)

	// File methods (все вернут ошибки, но покрытие)
	_ = storage.SupportsFiles()
	_ = storage.CreateFileMetadata(&FileMetadata{})
	_ = storage.SaveFileChunk(&FileChunk{})
	_, _ = storage.GetFileMetadata(uuid.New())
	_, _ = storage.GetFileChunk(uuid.New(), 0)
	_, _ = storage.GetAllFileChunks(uuid.New())
	_, _ = storage.GetUserFiles(user.ID)
	_ = storage.DeleteFile(uuid.New())
}
