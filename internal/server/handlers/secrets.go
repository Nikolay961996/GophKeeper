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

// SecretsHandler обработчик секретных данных
type SecretsHandler struct {
	storage storage.Storage
}

// NewSecretsHandler создает новый SecretsHandler
func NewSecretsHandler(storage storage.Storage) *SecretsHandler {
	return &SecretsHandler{
		storage: storage,
	}
}

// Sync обрабатывает синхронизацию данных
func (h *SecretsHandler) Sync(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req common.SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Получаем изменения с сервера
	serverSecrets, err := h.storage.GetSecretsSince(userID, req.LastSync)
	if err != nil {
		http.Error(w, "Error getting secrets", http.StatusInternalServerError)
		return
	}

	// Сохраняем изменения от клиента
	var conflicts []common.SecretData
	for _, clientSecret := range req.Data {
		existingSecret, _ := h.getSecretByID(userID, clientSecret.ID)

		if existingSecret != nil && existingSecret.UpdatedAt.After(clientSecret.UpdatedAt) {
			// Конфликт версий
			conflicts = append(conflicts, *existingSecret)
			continue
		}

		clientSecret.UserID = userID
		clientSecret.UpdatedAt = time.Now()
		if err := h.storage.SaveSecretData(&clientSecret); err != nil {
			http.Error(w, "Error saving secret", http.StatusInternalServerError)
			return
		}
	}

	response := common.SyncResponse{
		LastSync:  time.Now(),
		Data:      serverSecrets,
		Conflicts: conflicts,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSecrets возвращает все секреты пользователя
func (h *SecretsHandler) GetSecrets(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	secrets, err := h.storage.GetUserSecrets(userID)
	if err != nil {
		http.Error(w, "Error getting secrets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(secrets)
}

// getSecretByID возвращает секрет по ID
func (h *SecretsHandler) getSecretByID(userID, secretID uuid.UUID) (*common.SecretData, error) {
	secrets, err := h.storage.GetUserSecrets(userID)
	if err != nil {
		return nil, err
	}

	for _, secret := range secrets {
		if secret.ID == secretID {
			return secret, nil
		}
	}

	return nil, storage.ErrSecretNotFound
}
