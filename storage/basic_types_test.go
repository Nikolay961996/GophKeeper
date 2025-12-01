package storage

import (
	"testing"
	"time"

	"gophkeeper/internal/common"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestBasicTypes проверяет базовые структуры данных
func TestBasicTypes(t *testing.T) {
	fileMeta := FileMetadata{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		FileName:    "test.jpg",
		FileSize:    1024,
		TotalChunks: 4,
		ChunkSize:   256,
		Checksum:    "md5hash",
		MimeType:    "image/jpeg",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.Equal(t, "test.jpg", fileMeta.FileName)
	assert.Equal(t, int64(1024), fileMeta.FileSize)
	assert.Equal(t, 4, fileMeta.TotalChunks)

	fileChunk := FileChunk{
		ID:            uuid.New(),
		FileID:        uuid.New(),
		ChunkIndex:    2,
		ChunkData:     []byte{1, 2, 3, 4},
		ChunkChecksum: "chunkhash",
		CreatedAt:     time.Now(),
	}

	assert.Equal(t, 2, fileChunk.ChunkIndex)
	assert.Len(t, fileChunk.ChunkData, 4)
}

// TestInterfaceCompliance проверяет соответствие интерфейсам
func TestInterfaceCompliance(t *testing.T) {
	var storage *PostgresStorage

	_ = Storage(storage)
	_ = FileStorage(storage)
	_ = FileStorageChecker(storage)

	memoryStorage := NewMemoryStorage()
	_ = Storage(memoryStorage)
	_ = FileStorageChecker(memoryStorage)
}

// TestMemoryStorageFileMethods проверяет file методы MemoryStorage (всегда ошибки)
func TestMemoryStorageFileMethods(t *testing.T) {
	storage := NewMemoryStorage()

	err := storage.CreateFileMetadata(&FileMetadata{})
	assert.Error(t, err)

	err = storage.SaveFileChunk(&FileChunk{})
	assert.Error(t, err)

	_, err = storage.GetFileMetadata(uuid.New())
	assert.Error(t, err)

	_, err = storage.GetFileChunk(uuid.New(), 0)
	assert.Error(t, err)

	_, err = storage.GetAllFileChunks(uuid.New())
	assert.Error(t, err)

	_, err = storage.GetUserFiles(uuid.New())
	assert.Error(t, err)

	err = storage.DeleteFile(uuid.New())
	assert.Error(t, err)

	assert.False(t, storage.SupportsFiles())
}

// TestCommonSecretData проверяет работу с common.SecretData
func TestCommonSecretData(t *testing.T) {
	secret := &common.SecretData{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Type:      common.LoginPasswordType,
		Name:      "Test Secret",
		Metadata:  `{"site": "example.com"}`,
		Data:      []byte("encrypted data"),
		Version:   1,
		CreatedAt: common.Now(),
		UpdatedAt: common.Now(),
	}

	assert.Equal(t, common.LoginPasswordType, secret.Type)
	assert.Equal(t, "Test Secret", secret.Name)
	assert.Equal(t, 1, secret.Version)
}
