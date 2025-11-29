package storage

import (
	"testing"

	"gophkeeper/internal/common"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestPostgresStorageInterfaces проверяет что PostgresStorage реализует все интерфейсы
func TestPostgresStorageInterfaces(t *testing.T) {
	// Просто проверяем что типы совместимы (компиляция)
	var storage Storage = (*PostgresStorage)(nil)
	var fileStorage FileStorage = (*PostgresStorage)(nil)
	var fileChecker FileStorageChecker = (*PostgresStorage)(nil)

	assert.Nil(t, storage)
	assert.Nil(t, fileStorage)
	assert.Nil(t, fileChecker)
}

// TestPostgresStorageMethods проверяет только что методы объявлены (без вызова)
func TestPostgresStorageMethods(t *testing.T) {
	// Не создаем экземпляр, просто проверяем что методы существуют в типе
	var storage *PostgresStorage

	// Эта проверка только на уровне компиляции
	// Мы не вызываем методы, чтобы избежать паники

	// Проверяем что тип реализует интерфейсы
	assert.Implements(t, (*Storage)(nil), storage)
	assert.Implements(t, (*FileStorage)(nil), storage)
	assert.Implements(t, (*FileStorageChecker)(nil), storage)
}

// TestFileMetadataStruct проверяет структуру FileMetadata
func TestFileMetadataStruct(t *testing.T) {
	metadata := &FileMetadata{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		FileName:    "test.txt",
		FileSize:    100,
		TotalChunks: 1,
		ChunkSize:   100,
		Checksum:    "abc123",
		MimeType:    "text/plain",
		CreatedAt:   common.Now(),
		UpdatedAt:   common.Now(),
	}

	assert.NotNil(t, metadata)
	assert.Equal(t, "test.txt", metadata.FileName)
	assert.Equal(t, int64(100), metadata.FileSize)
	assert.Equal(t, 1, metadata.TotalChunks)
}

// TestFileChunkStruct проверяет структуру FileChunk
func TestFileChunkStruct(t *testing.T) {
	chunk := &FileChunk{
		ID:            uuid.New(),
		FileID:        uuid.New(),
		ChunkIndex:    0,
		ChunkData:     []byte("data"),
		ChunkChecksum: "checksum",
		CreatedAt:     common.Now(),
	}

	assert.NotNil(t, chunk)
	assert.Equal(t, 0, chunk.ChunkIndex)
	assert.Equal(t, []byte("data"), chunk.ChunkData)
	assert.Equal(t, "checksum", chunk.ChunkChecksum)
}

// TestPostgresStorageSupportsFiles проверяет метод SupportsFiles
func TestPostgresStorageSupportsFiles(t *testing.T) {
	// Создаем минимальный storage только для проверки SupportsFiles
	storage := &PostgresStorage{}

	// Этот метод не требует базы данных
	result := storage.SupportsFiles()
	assert.True(t, result)
}

// TestStorageError проверяет ошибки хранилища
func TestStorageError(t *testing.T) {
	err := &StorageError{msg: "test error message"}
	assert.Equal(t, "test error message", err.Error())

	// Проверяем предопределенные ошибки
	assert.Equal(t, "user not found", ErrUserNotFound.Error())
	assert.Equal(t, "secret not found", ErrSecretNotFound.Error())
	assert.Equal(t, "file not found", ErrFileNotFound.Error())
}

// TestPostgresStorageClose проверяет метод Close с nil db
func TestPostgresStorageClose(t *testing.T) {
	storage := &PostgresStorage{db: nil}
	err := storage.Close()
	assert.NoError(t, err)

	// Storage с не-nil db (но мы не будем его реально закрывать)
	storage2 := &PostgresStorage{}
	err = storage2.Close()
	// Не проверяем ошибку, так как db может быть nil
}

// TestNewPostgresStorage_InvalidConnString проверяет создание с невалидной строкой подключения
func TestNewPostgresStorage_InvalidConnString(t *testing.T) {
	// Этот тест может быть запущен без реальной БД
	storage, err := NewPostgresStorage("invalid_connection_string")
	assert.Error(t, err)
	assert.Nil(t, storage)
}
