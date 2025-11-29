package manager

import (
	"os"
	"testing"

	"gophkeeper/internal/client/config"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDataManager(t *testing.T) {
	// Создаем временную директорию для тестов
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	cfg := &config.Config{
		UserID: uuid.New().String(),
		Token:  "test-token",
	}

	manager, err := NewDataManager(cfg, "testpassword")
	require.NoError(t, err)
	assert.NotNil(t, manager)
}

func TestDataManager_SaveAndGetLoginPassword(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	userID := uuid.New()
	cfg := &config.Config{
		UserID: userID.String(),
		Token:  "test-token",
	}

	manager, err := NewDataManager(cfg, "testpassword")
	require.NoError(t, err)

	// Сохраняем логин/пароль
	err = manager.SaveLoginPassword("test site", "testuser", "testpass", "example.com")
	require.NoError(t, err)

	// Получаем список данных
	secrets := manager.ListData()
	assert.Len(t, secrets, 1)
	assert.Equal(t, "test site", secrets[0].Metadata)

	// Получаем конкретные данные
	secretID := secrets[0].ID.String()
	data, err := manager.GetLoginPassword(secretID)
	require.NoError(t, err)
	assert.Equal(t, "testuser", data.Login)
	assert.Equal(t, "testpass", data.Password)
	assert.Equal(t, "example.com", data.Site)
}

func TestDataManager_SaveAndGetCardData(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	userID := uuid.New()
	cfg := &config.Config{
		UserID: userID.String(),
		Token:  "test-token",
	}

	manager, err := NewDataManager(cfg, "testpassword")
	require.NoError(t, err)

	// Сохраняем данные карты
	err = manager.SaveCardData("test card", "4111111111111111", "12/25", "123", "John Doe", "Test Bank")
	require.NoError(t, err)

	// Получаем данные карты
	secrets := manager.ListData()
	secretID := secrets[0].ID.String()
	data, err := manager.GetCardData(secretID)
	require.NoError(t, err)
	assert.Equal(t, "4111111111111111", data.Number)
	assert.Equal(t, "12/25", data.Expiry)
	assert.Equal(t, "123", data.CVV)
	assert.Equal(t, "John Doe", data.Holder)
}

func TestDataManager_SaveAndGetTextData(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	userID := uuid.New()
	cfg := &config.Config{
		UserID: userID.String(),
		Token:  "test-token",
	}

	manager, err := NewDataManager(cfg, "testpassword")
	require.NoError(t, err)

	// Сохраняем текстовые данные
	text := "This is some sensitive text data"
	err = manager.SaveTextData("test text", text)
	require.NoError(t, err)

	// Получаем текстовые данные
	secrets := manager.ListData()
	secretID := secrets[0].ID.String()
	result, err := manager.GetTextData(secretID)
	require.NoError(t, err)
	assert.Equal(t, text, result)
}

func TestDataManager_DeleteData(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	userID := uuid.New()
	cfg := &config.Config{
		UserID: userID.String(),
		Token:  "test-token",
	}

	manager, err := NewDataManager(cfg, "testpassword")
	require.NoError(t, err)

	// Сохраняем данные
	err = manager.SaveTextData("test text", "data to delete")
	require.NoError(t, err)

	// Удаляем данные
	secrets := manager.ListData()
	secretID := secrets[0].ID.String()

	err = manager.DeleteData(secretID)
	require.NoError(t, err)

	// Проверяем, что данных нет
	secrets = manager.ListData()
	assert.Len(t, secrets, 0)
}

func TestDataManager_GetSecretByID(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	userID := uuid.New()
	cfg := &config.Config{
		UserID: userID.String(),
		Token:  "test-token",
	}

	manager, err := NewDataManager(cfg, "testpassword")
	require.NoError(t, err)

	// Сохраняем данные
	err = manager.SaveTextData("test text", "some data")
	require.NoError(t, err)

	// Ищем по ID
	secrets := manager.ListData()
	secretID := secrets[0].ID.String()

	secret := manager.GetSecretByID(secretID)
	assert.NotNil(t, secret)
	assert.Equal(t, "test text", secret.Metadata)

	// Ищем несуществующий ID
	secret = manager.GetSecretByID(uuid.New().String())
	assert.Nil(t, secret)
}
