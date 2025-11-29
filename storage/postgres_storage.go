package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"log"
	"time"

	"github.com/google/uuid"
	"gophkeeper/internal/common"
)

// FileMetadata представляет метаданные файла
type FileMetadata struct {
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	FileName    string    `json:"file_name"`
	Checksum    string    `json:"checksum"`
	MimeType    string    `json:"mime_type"`
	FileSize    int64     `json:"file_size"`
	TotalChunks int       `json:"total_chunks"`
	ChunkSize   int       `json:"chunk_size"`
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
}

// FileChunk представляет чанк файла
type FileChunk struct {
	CreatedAt     time.Time `json:"created_at"`
	ChunkChecksum string    `json:"chunk_checksum"`
	ChunkData     []byte    `json:"chunk_data"`
	ChunkIndex    int       `json:"chunk_index"`
	ID            uuid.UUID `json:"id"`
	FileID        uuid.UUID `json:"file_id"`
}

// PostgresStorage реализует хранение данных в PostgreSQL
type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage создает новый экземпляр PostgresStorage
func NewPostgresStorage(connString string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	// Настраиваем пул соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Инициализируем таблицы
	if err := initTables(db); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %v", err)
	}

	return &PostgresStorage{db: db}, nil
}

// Close закрывает соединение с базой данных
func (s *PostgresStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// InitFileTables инициализирует таблицы для файлов
func initFileTables(db *sql.DB) error {
	// Таблица для метаданных файлов
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS file_metadata (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			file_name TEXT NOT NULL,
			file_size BIGINT NOT NULL,
			total_chunks INTEGER NOT NULL,
			chunk_size INTEGER NOT NULL,
			checksum TEXT,
			mime_type TEXT,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create file_metadata table: %v", err)
	}

	// Таблица для чанков файлов
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS file_chunks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			file_id UUID NOT NULL REFERENCES file_metadata(id) ON DELETE CASCADE,
			chunk_index INTEGER NOT NULL,
			chunk_data BYTEA NOT NULL,
			chunk_checksum TEXT,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			UNIQUE(file_id, chunk_index)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create file_chunks table: %v", err)
	}

	// Добавляем колонку file_id в secrets если её нет
	_, err = db.Exec(`
		DO $$ 
		BEGIN 
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
						  WHERE table_name='secrets' AND column_name='file_id') THEN
				ALTER TABLE secrets ADD COLUMN file_id UUID REFERENCES file_metadata(id) ON DELETE SET NULL;
			END IF;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to add file_id column: %v", err)
	}

	// Индексы для производительности
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_file_chunks_file_id ON file_chunks(file_id);
		CREATE INDEX IF NOT EXISTS idx_file_chunks_file_id_index ON file_chunks(file_id, chunk_index);
		CREATE INDEX IF NOT EXISTS idx_file_metadata_user_id ON file_metadata(user_id);
	`)
	if err != nil {
		return fmt.Errorf("failed to create file indexes: %v", err)
	}

	return nil
}

// Update initTables to include file tables
func initTables(db *sql.DB) error {
	// Существующий код для users и secrets...
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			login TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS secrets (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			type TEXT NOT NULL,
			name TEXT NOT NULL,
			data BYTEA NOT NULL,
			metadata TEXT NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
			UNIQUE(user_id, id)
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create secrets table: %v", err)
	}

	// Инициализируем таблицы для файлов
	if err = initFileTables(db); err != nil {
		return err
	}

	// Существующие индексы...
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_secrets_user_id ON secrets(user_id);
		CREATE INDEX IF NOT EXISTS idx_secrets_updated_at ON secrets(updated_at);
		CREATE INDEX IF NOT EXISTS idx_secrets_user_updated ON secrets(user_id, updated_at);
	`)
	if err != nil {
		return fmt.Errorf("failed to create indexes: %v", err)
	}

	return nil
}

// File Methods

// CreateFileMetadata создает метаданные файла
func (s *PostgresStorage) CreateFileMetadata(metadata *FileMetadata) error {
	query := `
		INSERT INTO file_metadata (id, user_id, file_name, file_size, total_chunks, chunk_size, checksum, mime_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := s.db.Exec(
		query,
		metadata.ID,
		metadata.UserID,
		metadata.FileName,
		metadata.FileSize,
		metadata.TotalChunks,
		metadata.ChunkSize,
		metadata.Checksum,
		metadata.MimeType,
		metadata.CreatedAt,
		metadata.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create file metadata: %v", err)
	}

	return nil
}

// SaveFileChunk сохраняет чанк файла
func (s *PostgresStorage) SaveFileChunk(chunk *FileChunk) error {
	query := `
		INSERT INTO file_chunks (file_id, chunk_index, chunk_data, chunk_checksum, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (file_id, chunk_index) 
		DO UPDATE SET chunk_data = EXCLUDED.chunk_data, chunk_checksum = EXCLUDED.chunk_checksum, created_at = EXCLUDED.created_at
	`

	_, err := s.db.Exec(
		query,
		chunk.FileID,
		chunk.ChunkIndex,
		chunk.ChunkData,
		chunk.ChunkChecksum,
		chunk.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save file chunk: %v", err)
	}

	return nil
}

// GetFileMetadata возвращает метаданные файла
func (s *PostgresStorage) GetFileMetadata(fileID uuid.UUID) (*FileMetadata, error) {
	query := `
		SELECT id, user_id, file_name, file_size, total_chunks, chunk_size, checksum, mime_type, created_at, updated_at
		FROM file_metadata
		WHERE id = $1
	`

	metadata := &FileMetadata{}
	err := s.db.QueryRow(query, fileID).Scan(
		&metadata.ID,
		&metadata.UserID,
		&metadata.FileName,
		&metadata.FileSize,
		&metadata.TotalChunks,
		&metadata.ChunkSize,
		&metadata.Checksum,
		&metadata.MimeType,
		&metadata.CreatedAt,
		&metadata.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, &StorageError{"file not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %v", err)
	}

	return metadata, nil
}

// GetFileChunk возвращает конкретный чанк файла
func (s *PostgresStorage) GetFileChunk(fileID uuid.UUID, chunkIndex int) (*FileChunk, error) {
	query := `
		SELECT id, file_id, chunk_index, chunk_data, chunk_checksum, created_at
		FROM file_chunks
		WHERE file_id = $1 AND chunk_index = $2
	`

	chunk := &FileChunk{}
	err := s.db.QueryRow(query, fileID, chunkIndex).Scan(
		&chunk.ID,
		&chunk.FileID,
		&chunk.ChunkIndex,
		&chunk.ChunkData,
		&chunk.ChunkChecksum,
		&chunk.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, &StorageError{"chunk not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get file chunk: %v", err)
	}

	return chunk, nil
}

// GetAllFileChunks возвращает все чанки файла в порядке индекса
func (s *PostgresStorage) GetAllFileChunks(fileID uuid.UUID) ([]*FileChunk, error) {
	query := `
		SELECT id, file_id, chunk_index, chunk_data, chunk_checksum, created_at
		FROM file_chunks
		WHERE file_id = $1
		ORDER BY chunk_index ASC
	`

	rows, err := s.db.Query(query, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query file chunks: %v", err)
	}
	defer rows.Close()

	var chunks []*FileChunk
	for rows.Next() {
		chunk := &FileChunk{}
		err = rows.Scan(
			&chunk.ID,
			&chunk.FileID,
			&chunk.ChunkIndex,
			&chunk.ChunkData,
			&chunk.ChunkChecksum,
			&chunk.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file chunk: %v", err)
		}
		chunks = append(chunks, chunk)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating file chunks: %v", err)
	}

	return chunks, nil
}

// GetUserFiles возвращает все файлы пользователя
func (s *PostgresStorage) GetUserFiles(userID uuid.UUID) ([]*FileMetadata, error) {
	query := `
		SELECT id, user_id, file_name, file_size, total_chunks, chunk_size, checksum, mime_type, created_at, updated_at
		FROM file_metadata
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user files: %v", err)
	}
	defer rows.Close()

	var files []*FileMetadata
	for rows.Next() {
		file := &FileMetadata{}
		err = rows.Scan(
			&file.ID,
			&file.UserID,
			&file.FileName,
			&file.FileSize,
			&file.TotalChunks,
			&file.ChunkSize,
			&file.Checksum,
			&file.MimeType,
			&file.CreatedAt,
			&file.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file metadata: %v", err)
		}
		files = append(files, file)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user files: %v", err)
	}

	return files, nil
}

// DeleteFile удаляет файл и все его чанки
func (s *PostgresStorage) DeleteFile(fileID uuid.UUID) error {
	// Удаляем чанки (каскадно удалится из-за ON DELETE CASCADE)
	// Но лучше явно удалить для ясности
	_, err := s.db.Exec("DELETE FROM file_chunks WHERE file_id = $1", fileID)
	if err != nil {
		return fmt.Errorf("failed to delete file chunks: %v", err)
	}

	// Удаляем метаданные
	result, err := s.db.Exec("DELETE FROM file_metadata WHERE id = $1", fileID)
	if err != nil {
		return fmt.Errorf("failed to delete file metadata: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return &StorageError{"file not found"}
	}

	return nil
}

// SaveSecretData для поддержки file_id
func (s *PostgresStorage) SaveSecretData(data *common.SecretData) error {
	// Проверяем, есть ли file_id в данных
	var fileID *uuid.UUID
	if data.Type == common.BinaryDataType {
		// Для бинарных данных может быть ссылка на файл
		// В метаданных может храниться file_id
		var meta struct {
			FileID string `json:"file_id,omitempty"`
		}
		if err := json.Unmarshal([]byte(data.Metadata), &meta); err == nil && meta.FileID != "" {
			if id, err := uuid.Parse(meta.FileID); err == nil {
				fileID = &id
			}
		}
	}

	// Сначала проверяем, существует ли уже запись
	var existingVersion int
	var existingFileID *uuid.UUID
	checkQuery := `SELECT version, file_id FROM secrets WHERE id = $1 AND user_id = $2`
	err := s.db.QueryRow(checkQuery, data.ID, data.UserID).Scan(&existingVersion, &existingFileID)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing secret: %v", err)
	}

	if err == sql.ErrNoRows {
		// Новая запись
		query := `
			INSERT INTO secrets (id, user_id, type, name, data, metadata, version, created_at, updated_at, file_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		_, err = s.db.Exec(
			query,
			data.ID,
			data.UserID,
			string(data.Type),
			data.Name,
			data.Data,
			data.Metadata,
			data.Version,
			data.CreatedAt,
			data.UpdatedAt,
			fileID,
		)
	} else {
		// Обновление существующей записи
		query := `
			UPDATE secrets 
			SET type = $1, name = $2, data = $3, metadata = $4, version = $5, updated_at = $6, file_id = $7
			WHERE id = $8 AND user_id = $9
		`
		_, err = s.db.Exec(
			query,
			string(data.Type),
			data.Name,
			data.Data,
			data.Metadata,
			data.Version,
			data.UpdatedAt,
			fileID,
			data.ID,
			data.UserID,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to save secret data: %v", err)
	}

	return nil
}

// CreateUser создает нового пользователя
func (s *PostgresStorage) CreateUser(user *common.User) error {
	query := `
        INSERT INTO users (id, login, password_hash, created_at)
        VALUES ($1, $2, $3, $4)
    `

	_, err := s.db.Exec(
		query,
		user.ID,
		user.Login,
		user.PasswordHash,
		user.CreatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return &StorageError{"user already exists"}
			}
		}
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

// GetUserByLogin возвращает пользователя по логину
func (s *PostgresStorage) GetUserByLogin(login string) (*common.User, error) {
	query := `
        SELECT id, login, password_hash, created_at
        FROM users
        WHERE login = $1
    `

	user := &common.User{}
	err := s.db.QueryRow(query, login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by login: %v", err)
	}

	return user, nil
}

// GetUserByID возвращает пользователя по ID
func (s *PostgresStorage) GetUserByID(id uuid.UUID) (*common.User, error) {
	query := `
        SELECT id, login, password_hash, created_at
        FROM users
        WHERE id = $1
    `

	user := &common.User{}
	err := s.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %v", err)
	}

	return user, nil
}

// GetUserSecrets возвращает все секреты пользователя
func (s *PostgresStorage) GetUserSecrets(userID uuid.UUID) ([]*common.SecretData, error) {
	query := `
        SELECT id, user_id, type, name, data, metadata, version, created_at, updated_at
        FROM secrets
        WHERE user_id = $1
        ORDER BY updated_at DESC
    `

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user secrets: %v", err)
	}
	defer rows.Close()

	var secrets []*common.SecretData
	for rows.Next() {
		secret := &common.SecretData{}
		var typeStr string

		err = rows.Scan(
			&secret.ID,
			&secret.UserID,
			&typeStr,
			&secret.Name,
			&secret.Data,
			&secret.Metadata,
			&secret.Version,
			&secret.CreatedAt,
			&secret.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan secret: %v", err)
		}

		secret.Type = common.DataType(typeStr)
		secrets = append(secrets, secret)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating secrets: %v", err)
	}

	return secrets, nil
}

// GetSecretsSince возвращает секреты, измененные после указанной даты
func (s *PostgresStorage) GetSecretsSince(userID uuid.UUID, since time.Time) ([]*common.SecretData, error) {
	query := `
        SELECT id, user_id, type, name, data, metadata, version, created_at, updated_at
        FROM secrets
        WHERE user_id = $1 AND updated_at > $2
        ORDER BY updated_at DESC
    `

	rows, err := s.db.Query(query, userID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query secrets since: %v", err)
	}
	defer rows.Close()

	var secrets []*common.SecretData
	for rows.Next() {
		secret := &common.SecretData{}
		var typeStr string

		err = rows.Scan(
			&secret.ID,
			&secret.UserID,
			&typeStr,
			&secret.Name,
			&secret.Data,
			&secret.Metadata,
			&secret.Version,
			&secret.CreatedAt,
			&secret.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan secret: %v", err)
		}

		secret.Type = common.DataType(typeStr)
		secrets = append(secrets, secret)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating secrets: %v", err)
	}

	return secrets, nil
}

// DeleteSecret удаляет секрет
func (s *PostgresStorage) DeleteSecret(userID, secretID uuid.UUID) error {
	// Сначала проверяем тип секрета, если это файл - удаляем его данные
	var secretType string
	var fileID *uuid.UUID

	checkQuery := `SELECT type, file_id FROM secrets WHERE id = $1 AND user_id = $2`
	err := s.db.QueryRow(checkQuery, secretID, userID).Scan(&secretType, &fileID)
	if err == sql.ErrNoRows {
		return ErrSecretNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to check secret type: %v", err)
	}

	// Если это бинарные данные с привязанным файлом - удаляем файл
	if secretType == string(common.BinaryDataType) && fileID != nil {
		if err = s.DeleteFile(*fileID); err != nil {
			log.Printf("Warning: failed to delete file %s: %v", fileID, err)
			// Продолжаем удаление секрета даже если файл не удалился
		}
	}

	// Удаляем сам секрет
	query := `DELETE FROM secrets WHERE id = $1 AND user_id = $2`
	result, err := s.db.Exec(query, secretID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return ErrSecretNotFound
	}

	return nil
}

// SupportsFiles проверяет, поддерживает ли хранилище работу с файлами
func (s *PostgresStorage) SupportsFiles() bool {
	return true
}

var _ Storage = (*PostgresStorage)(nil)
var _ FileStorage = (*PostgresStorage)(nil)
var _ FileStorageChecker = (*PostgresStorage)(nil)
