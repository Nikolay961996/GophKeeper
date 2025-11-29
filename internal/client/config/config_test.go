package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Default(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldHome := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Setenv("USERPROFILE", oldHome)

	// Remove any existing config file
	configPath := filepath.Join(tmpDir, ".gophkeeper", "config.json")
	os.RemoveAll(filepath.Dir(configPath))

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "localhost:8081", config.ServerURL)
	assert.Empty(t, config.Token)
	assert.Empty(t, config.UserID)
}

func TestLoadConfig_FromFile(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldHome := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Setenv("USERPROFILE", oldHome)

	// Create config directory and file
	configDir := filepath.Join(tmpDir, ".gophkeeper")
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err)

	configData := `{
		"server_url": "test-server:9090",
		"token": "test-token",
		"user_id": "test-user-id"
	}`

	configPath := filepath.Join(configDir, "config.json")
	err = os.WriteFile(configPath, []byte(configData), 0600)
	require.NoError(t, err)

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "test-server:9090", config.ServerURL)
	assert.Equal(t, "test-token", config.Token)
	assert.Equal(t, "test-user-id", config.UserID)
}

func TestSaveConfig(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldHome := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Setenv("USERPROFILE", oldHome)

	config := &Config{
		ServerURL: "test-server:9090",
		Token:     "test-token",
		UserID:    "test-user-id",
	}

	err := SaveConfig(config)
	require.NoError(t, err)

	// Verify file was created and can be loaded
	loadedConfig, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, config.ServerURL, loadedConfig.ServerURL)
	assert.Equal(t, config.Token, loadedConfig.Token)
	assert.Equal(t, config.UserID, loadedConfig.UserID)
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	oldHome := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", tmpDir)
	defer os.Setenv("USERPROFILE", oldHome)

	// Create config directory and invalid config file
	configDir := filepath.Join(tmpDir, ".gophkeeper")
	err := os.MkdirAll(configDir, 0700)
	require.NoError(t, err)

	configPath := filepath.Join(configDir, "config.json")
	err = os.WriteFile(configPath, []byte("invalid json"), 0600)
	require.NoError(t, err)

	_, err = LoadConfig()
	assert.Error(t, err)
}
