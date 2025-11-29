package storage

import (
	"os"
	"testing"
	"time"

	"gophkeeper/internal/common"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestDB() string {
	if connStr := os.Getenv("TEST_DATABASE_URL"); connStr != "" {
		return connStr
	}
	return "postgres://test:test@localhost:5432/gophkeeper_test?sslmode=disable"
}

func TestNewPostgresStorage(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping Postgres tests")
	}

	storage, err := NewPostgresStorage(getTestDB())
	require.NoError(t, err)
	assert.NotNil(t, storage)

	assert.True(t, storage.SupportsFiles())

	defer storage.Close()
}

func TestPostgresStorage_UserOperations(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping Postgres tests")
	}

	storage, err := NewPostgresStorage(getTestDB())
	require.NoError(t, err)
	defer storage.Close()

	defer func() {
		storage.db.Exec("DELETE FROM users WHERE login LIKE 'testuser%'")
	}()

	user := &common.User{
		ID:           uuid.New(),
		Login:        "testuser_" + uuid.New().String(),
		PasswordHash: "hashedpassword",
		CreatedAt:    common.Now(),
	}

	err = storage.CreateUser(user)
	require.NoError(t, err)

	foundUser, err := storage.GetUserByLogin(user.Login)
	require.NoError(t, err)
	assert.Equal(t, user.ID, foundUser.ID)
	assert.Equal(t, user.Login, foundUser.Login)

	foundUserByID, err := storage.GetUserByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, foundUserByID.ID)

	err = storage.CreateUser(user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	_, err = storage.GetUserByLogin("nonexistent")
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)

	_, err = storage.GetUserByID(uuid.New())
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
}

func TestPostgresStorage_SecretOperations(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping Postgres tests")
	}

	storage, err := NewPostgresStorage(getTestDB())
	require.NoError(t, err)
	defer storage.Close()

	user := &common.User{
		ID:           uuid.New(),
		Login:        "testuser_" + uuid.New().String(),
		PasswordHash: "hashedpassword",
		CreatedAt:    common.Now(),
	}
	err = storage.CreateUser(user)
	require.NoError(t, err)

	defer func() {
		storage.db.Exec("DELETE FROM users WHERE id = $1", user.ID)
	}()

	secret := &common.SecretData{
		ID:        uuid.New(),
		UserID:    user.ID,
		Type:      common.LoginPasswordType,
		Name:      "Test Secret",
		Metadata:  "test metadata",
		Data:      []byte("test data"),
		Version:   1,
		CreatedAt: common.Now(),
		UpdatedAt: common.Now(),
	}

	err = storage.SaveSecretData(secret)
	require.NoError(t, err)

	secrets, err := storage.GetUserSecrets(user.ID)
	require.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, secret.ID, secrets[0].ID)

	secretsSince, err := storage.GetSecretsSince(user.ID, common.Now().Add(-time.Hour))
	require.NoError(t, err)
	assert.Len(t, secretsSince, 1)

	secret.Version = 2
	secret.UpdatedAt = common.Now()
	err = storage.SaveSecretData(secret)
	require.NoError(t, err)

	err = storage.DeleteSecret(user.ID, secret.ID)
	require.NoError(t, err)

	secrets, err = storage.GetUserSecrets(user.ID)
	require.NoError(t, err)
	assert.Len(t, secrets, 0)

	err = storage.DeleteSecret(user.ID, uuid.New())
	assert.Error(t, err)
}

func TestPostgresStorage_FileMetadataOperations(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping Postgres tests")
	}

	storage, err := NewPostgresStorage(getTestDB())
	require.NoError(t, err)
	defer storage.Close()

	user := &common.User{
		ID:           uuid.New(),
		Login:        "testuser_" + uuid.New().String(),
		PasswordHash: "hashedpassword",
		CreatedAt:    common.Now(),
	}
	err = storage.CreateUser(user)
	require.NoError(t, err)

	defer func() {
		storage.db.Exec("DELETE FROM users WHERE id = $1", user.ID)
	}()

	fileMetadata := &FileMetadata{
		ID:          uuid.New(),
		UserID:      user.ID,
		FileName:    "testfile.txt",
		FileSize:    1024,
		TotalChunks: 5,
		ChunkSize:   256,
		Checksum:    "abc123",
		MimeType:    "text/plain",
		CreatedAt:   common.Now(),
		UpdatedAt:   common.Now(),
	}

	err = storage.CreateFileMetadata(fileMetadata)
	require.NoError(t, err)

	retrievedMetadata, err := storage.GetFileMetadata(fileMetadata.ID)
	require.NoError(t, err)
	assert.Equal(t, fileMetadata.ID, retrievedMetadata.ID)
	assert.Equal(t, fileMetadata.FileName, retrievedMetadata.FileName)

	_, err = storage.GetFileMetadata(uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	files, err := storage.GetUserFiles(user.ID)
	require.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, fileMetadata.ID, files[0].ID)

	err = storage.DeleteFile(fileMetadata.ID)
	require.NoError(t, err)

	_, err = storage.GetFileMetadata(fileMetadata.ID)
	assert.Error(t, err)
}

func TestPostgresStorage_FileChunkOperations(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping Postgres tests")
	}

	storage, err := NewPostgresStorage(getTestDB())
	require.NoError(t, err)
	defer storage.Close()

	user := &common.User{
		ID:           uuid.New(),
		Login:        "testuser_" + uuid.New().String(),
		PasswordHash: "hashedpassword",
		CreatedAt:    common.Now(),
	}
	err = storage.CreateUser(user)
	require.NoError(t, err)

	fileMetadata := &FileMetadata{
		ID:          uuid.New(),
		UserID:      user.ID,
		FileName:    "testfile.txt",
		FileSize:    1024,
		TotalChunks: 3,
		ChunkSize:   256,
		CreatedAt:   common.Now(),
		UpdatedAt:   common.Now(),
	}
	err = storage.CreateFileMetadata(fileMetadata)
	require.NoError(t, err)

	defer func() {
		storage.db.Exec("DELETE FROM users WHERE id = $1", user.ID)
	}()

	chunk := &FileChunk{
		FileID:        fileMetadata.ID,
		ChunkIndex:    0,
		ChunkData:     []byte("chunk data 0"),
		ChunkChecksum: "checksum0",
		CreatedAt:     common.Now(),
	}

	err = storage.SaveFileChunk(chunk)
	require.NoError(t, err)

	retrievedChunk, err := storage.GetFileChunk(fileMetadata.ID, 0)
	require.NoError(t, err)
	assert.Equal(t, chunk.ChunkIndex, retrievedChunk.ChunkIndex)
	assert.Equal(t, chunk.ChunkData, retrievedChunk.ChunkData)

	_, err = storage.GetFileChunk(fileMetadata.ID, 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	chunks, err := storage.GetAllFileChunks(fileMetadata.ID)
	require.NoError(t, err)
	assert.Len(t, chunks, 1)

	chunk.ChunkData = []byte("updated chunk data")
	err = storage.SaveFileChunk(chunk)
	require.NoError(t, err)

	retrievedChunk, err = storage.GetFileChunk(fileMetadata.ID, 0)
	require.NoError(t, err)
	assert.Equal(t, []byte("updated chunk data"), retrievedChunk.ChunkData)
}

func TestPostgresStorage_EdgeCases(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping Postgres tests")
	}

	storage, err := NewPostgresStorage(getTestDB())
	require.NoError(t, err)
	defer storage.Close()

	_, err = NewPostgresStorage("invalid_connection_string")
	assert.Error(t, err)

	emptyStorage := &PostgresStorage{db: nil}
	err = emptyStorage.Close()
	assert.NoError(t, err)
}

func TestPostgresStorage_SecretWithFileReference(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping Postgres tests")
	}

	storage, err := NewPostgresStorage(getTestDB())
	require.NoError(t, err)
	defer storage.Close()

	user := &common.User{
		ID:           uuid.New(),
		Login:        "testuser_" + uuid.New().String(),
		PasswordHash: "hashedpassword",
		CreatedAt:    common.Now(),
	}
	err = storage.CreateUser(user)
	require.NoError(t, err)

	fileID := uuid.New()
	fileMetadata := &FileMetadata{
		ID:          fileID,
		UserID:      user.ID,
		FileName:    "binaryfile.bin",
		FileSize:    2048,
		TotalChunks: 2,
		ChunkSize:   1024,
		CreatedAt:   common.Now(),
		UpdatedAt:   common.Now(),
	}
	err = storage.CreateFileMetadata(fileMetadata)
	require.NoError(t, err)

	defer func() {
		storage.db.Exec("DELETE FROM users WHERE id = $1", user.ID)
	}()

	secret := &common.SecretData{
		ID:        uuid.New(),
		UserID:    user.ID,
		Type:      common.BinaryDataType,
		Name:      "Binary File Secret",
		Metadata:  `{"file_id": "` + fileID.String() + `"}`,
		Data:      []byte("file reference"),
		Version:   1,
		CreatedAt: common.Now(),
		UpdatedAt: common.Now(),
	}

	err = storage.SaveSecretData(secret)
	require.NoError(t, err)

	err = storage.DeleteSecret(user.ID, secret.ID)
	require.NoError(t, err)
}
