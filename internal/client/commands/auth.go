package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gophkeeper/internal/client/config"
	"gophkeeper/internal/common"
)

// AuthCommands обработчики команд аутентификации
type AuthCommands struct {
	cfg *config.Config
}

// NewAuthCommands создает новый AuthCommands
func NewAuthCommands(cfg *config.Config) *AuthCommands {
	return &AuthCommands{
		cfg: cfg,
	}
}

// Register регистрирует нового пользователя
func (a *AuthCommands) Register(login, password string) error {
	req := common.AuthRequest{
		Login:    login,
		Password: password,
	}

	resp, err := a.makeAuthRequest("/api/register", req)
	if err != nil {
		return err
	}

	a.cfg.Token = resp.Token
	a.cfg.UserID = resp.User.ID.String()

	if err := config.SaveConfig(a.cfg); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	fmt.Printf("Successfully registered user: %s\n", login)
	return nil
}

// Login выполняет вход пользователя
func (a *AuthCommands) Login(login, password string) error {
	req := common.AuthRequest{
		Login:    login,
		Password: password,
	}

	resp, err := a.makeAuthRequest("/api/login", req)
	if err != nil {
		return err
	}

	a.cfg.Token = resp.Token
	a.cfg.UserID = resp.User.ID.String()

	if err := config.SaveConfig(a.cfg); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	fmt.Printf("Successfully logged in as: %s\n", login)
	return nil
}

// makeAuthRequest выполняет запрос аутентификации
func (a *AuthCommands) makeAuthRequest(endpoint string, req common.AuthRequest) (*common.AuthResponse, error) {
	url := a.cfg.ServerURL + endpoint

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth failed: %s", string(body))
	}

	var authResp common.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &authResp, nil
}
