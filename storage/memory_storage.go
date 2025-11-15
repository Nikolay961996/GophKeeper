package storage

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"gophkeeper/internal/common"
)

// MemoryStorage реализует хранилище в памяти
type MemoryStorage struct {
	mu      sync.RWMutex
	users   map[uuid.UUID]*common.User
	secrets map[uuid.UUID][]*common.SecretData // user_id -> []SecretData
}

// NewMemoryStorage создает новый экземпляр MemoryStorage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		users:   make(map[uuid.UUID]*common.User),
		secrets: make(map[uuid.UUID][]*common.SecretData),
	}
}

// CreateUser создает нового пользователя
func (s *MemoryStorage) CreateUser(user *common.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[user.ID] = user
	s.secrets[user.ID] = make([]*common.SecretData, 0)
	return nil
}

// GetUserByLogin возвращает пользователя по логину
func (s *MemoryStorage) GetUserByLogin(login string) (*common.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Login == login {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

// GetUserByID возвращает пользователя по ID
func (s *MemoryStorage) GetUserByID(id uuid.UUID) (*common.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// SaveSecretData сохраняет секретные данные
func (s *MemoryStorage) SaveSecretData(data *common.SecretData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userSecrets := s.secrets[data.UserID]
	for i, secret := range userSecrets {
		if secret.ID == data.ID {
			userSecrets[i] = data
			return nil
		}
	}

	s.secrets[data.UserID] = append(userSecrets, data)
	return nil
}

// GetUserSecrets возвращает все секреты пользователя
func (s *MemoryStorage) GetUserSecrets(userID uuid.UUID) ([]*common.SecretData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	secrets, exists := s.secrets[userID]
	if !exists {
		return nil, ErrUserNotFound
	}

	return secrets, nil
}

// GetSecretsSince возвращает секреты, измененные после указанной даты
func (s *MemoryStorage) GetSecretsSince(userID uuid.UUID, since time.Time) ([]*common.SecretData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	allSecrets, exists := s.secrets[userID]
	if !exists {
		return nil, ErrUserNotFound
	}

	var filtered []*common.SecretData
	for _, secret := range allSecrets {
		if secret.UpdatedAt.After(since) {
			filtered = append(filtered, secret)
		}
	}

	return filtered, nil
}

// DeleteSecret удаляет секрет
func (s *MemoryStorage) DeleteSecret(userID, secretID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	userSecrets, exists := s.secrets[userID]
	if !exists {
		return ErrUserNotFound
	}

	for i, secret := range userSecrets {
		if secret.ID == secretID {
			s.secrets[userID] = append(userSecrets[:i], userSecrets[i+1:]...)
			return nil
		}
	}

	return ErrSecretNotFound
}

// Ошибки хранилища
var (
	ErrUserNotFound   = &StorageError{"user not found"}
	ErrSecretNotFound = &StorageError{"secret not found"}
)

// StorageError ошибка хранилища
type StorageError struct {
	msg string
}

func (e *StorageError) Error() string {
	return e.msg
}
