package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config представляет конфигурацию клиента
type Config struct {
	ServerURL string `json:"server_url"`
	Token     string `json:"token,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

// LoadConfig загружает конфигурацию из файла
func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	config := &Config{
		ServerURL: "http://localhost:8080",
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig сохраняет конфигурацию в файл
func SaveConfig(config *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Создаем директорию, если она не существует
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}

// getConfigPath возвращает путь к файлу конфигурации
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".gophkeeper", "config.json"), nil
}
