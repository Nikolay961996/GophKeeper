package manager

import (
	"os"
	"testing"

	"gophkeeper/internal/client/config"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetData_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Setenv("USERPROFILE", oldHome)

	cfg := &config.Config{
		UserID: uuid.New().String(),
		Token:  "test-token",
	}

	manager, err := NewDataManager(cfg, "testpassword")
	assert.NoError(t, err)

	_, err = manager.GetLoginPassword(uuid.New().String())
	assert.Error(t, err)

	_, err = manager.GetCardData(uuid.New().String())
	assert.Error(t, err)

	_, err = manager.GetTextData(uuid.New().String())
	assert.Error(t, err)

	_, _, err = manager.GetBinaryData(uuid.New().String())
	assert.Error(t, err)
}

func TestDeleteData_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Setenv("USERPROFILE", oldHome)

	cfg := &config.Config{
		UserID: uuid.New().String(),
		Token:  "test-token",
	}

	manager, err := NewDataManager(cfg, "testpassword")
	assert.NoError(t, err)

	err = manager.DeleteData(uuid.New().String())
	assert.Error(t, err)
}
