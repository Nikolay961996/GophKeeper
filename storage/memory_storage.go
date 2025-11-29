package storage

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"gophkeeper/internal/common"
)

// MemoryStorage реализует хранилище в памяти
type MemoryStorage struct {
	users   map[uuid.UUID]*common.User
	secrets map[uuid.UUID][]*common.SecretData
	mu      sync.RWMutex
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

	// Проверяем, нет ли уже пользователя с таким логином
	for _, u := range s.users {
		if u.Login == user.Login {
			return &StorageError{"user already exists"}
		}
	}

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

	userSecrets, exists := s.secrets[data.UserID]
	if !exists {
		return ErrUserNotFound
	}

	// Проверяем, существует ли уже запись с таким ID
	for i, secret := range userSecrets {
		if secret.ID == data.ID {
			userSecrets[i] = data
			return nil
		}
	}

	// Если не существует, добавляем новую запись
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

	// Возвращаем копию, чтобы избежать гонок данных
	result := make([]*common.SecretData, len(secrets))
	copy(result, secrets)
	return result, nil
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

// File methods - заглушки для MemoryStorage

// SupportsFiles проверяет, поддерживает ли хранилище работу с файлами
func (s *MemoryStorage) SupportsFiles() bool {
	return false
}

// CreateFileMetadata создает метаданные файла (не реализовано для in-memory)
func (s *MemoryStorage) CreateFileMetadata(metadata *FileMetadata) error {
	return &StorageError{"file storage not supported in memory mode"}
}

// SaveFileChunk сохраняет чанк файла (не реализовано для in-memory)
func (s *MemoryStorage) SaveFileChunk(chunk *FileChunk) error {
	return &StorageError{"file storage not supported in memory mode"}
}

// GetFileMetadata возвращает метаданные файла (не реализовано для in-memory)
func (s *MemoryStorage) GetFileMetadata(fileID uuid.UUID) (*FileMetadata, error) {
	return nil, &StorageError{"file storage not supported in memory mode"}
}

// GetFileChunk возвращает конкретный чанк файла (не реализовано для in-memory)
func (s *MemoryStorage) GetFileChunk(fileID uuid.UUID, chunkIndex int) (*FileChunk, error) {
	return nil, &StorageError{"file storage not supported in memory mode"}
}

// GetAllFileChunks возвращает все чанки файла (не реализовано для in-memory)
func (s *MemoryStorage) GetAllFileChunks(fileID uuid.UUID) ([]*FileChunk, error) {
	return nil, &StorageError{"file storage not supported in memory mode"}
}

// GetUserFiles возвращает все файлы пользователя (не реализовано для in-memory)
func (s *MemoryStorage) GetUserFiles(userID uuid.UUID) ([]*FileMetadata, error) {
	return nil, &StorageError{"file storage not supported in memory mode"}
}

// DeleteFile удаляет файл и все его чанки (не реализовано для in-memory)
func (s *MemoryStorage) DeleteFile(fileID uuid.UUID) error {
	return &StorageError{"file storage not supported in memory mode"}
}
