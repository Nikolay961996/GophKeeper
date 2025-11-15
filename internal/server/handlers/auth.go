package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"gophkeeper/internal/common"
	"gophkeeper/internal/server/middleware"
	"gophkeeper/storage"

	"github.com/google/uuid"
)

// AuthHandler обработчик аутентификации
type AuthHandler struct {
	storage   storage.Storage
	jwtSecret string
}

// NewAuthHandler создает новый AuthHandler
func NewAuthHandler(storage storage.Storage, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		storage:   storage,
		jwtSecret: jwtSecret,
	}
}

// Register обрабатывает регистрацию пользователя
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req common.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Проверяем, что пользователь не существует
	existingUser, _ := h.storage.GetUserByLogin(req.Login)
	if existingUser != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Хешируем пароль
	passwordHash, err := common.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Создаем пользователя
	user := &common.User{
		ID:           uuid.New(),
		Login:        req.Login,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}

	if err := h.storage.CreateUser(user); err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Генерируем токен
	token, err := common.GenerateJWTToken(user.ID, user.Login, h.jwtSecret, 24*time.Hour)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := common.AuthResponse{
		Token: token,
		User:  *user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Login обрабатывает вход пользователя
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req common.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ищем пользователя
	user, err := h.storage.GetUserByLogin(req.Login)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Проверяем пароль
	if !common.CheckPasswordHash(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Генерируем токен
	token, err := common.GenerateJWTToken(user.ID, user.Login, h.jwtSecret, 24*time.Hour)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := common.AuthResponse{
		Token: token,
		User:  *user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
