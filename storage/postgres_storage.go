package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"gophkeeper/internal/common"
)

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

	// Инициализируем таблицы
	if err := initTables(db); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %v", err)
	}

	return &PostgresStorage{db: db}, nil
}

// initTables создает необходимые таблицы если они не существуют
func initTables(db *sql.DB) error {
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

// Close закрывает соединение с базой данных
func (s *PostgresStorage) Close() error {
	return s.db.Close()
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

// SaveSecretData сохраняет секретные данные
func (s *PostgresStorage) SaveSecretData(data *common.SecretData) error {
	var existingVersion int
	checkQuery := `SELECT version FROM secrets WHERE id = $1 AND user_id = $2`
	err := s.db.QueryRow(checkQuery, data.ID, data.UserID).Scan(&existingVersion)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing secret: %v", err)
	}

	if err == sql.ErrNoRows {
		query := `
			INSERT INTO secrets (id, user_id, type, name, data, metadata, version, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
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
		)
	} else {
		query := `
			UPDATE secrets 
			SET type = $1, name = $2, data = $3, metadata = $4, version = $5, updated_at = $6
			WHERE id = $7 AND user_id = $8
		`
		_, err = s.db.Exec(
			query,
			string(data.Type),
			data.Name,
			data.Data,
			data.Metadata,
			data.Version,
			data.UpdatedAt,
			data.ID,
			data.UserID,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to save secret data: %v", err)
	}

	return nil
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

		err := rows.Scan(
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

		err := rows.Scan(
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

// HealthCheck проверяет соединение с базой данных
func (s *PostgresStorage) HealthCheck() error {
	return s.db.Ping()
}

// GetStatistics возвращает статистику по базе данных (для мониторинга)
func (s *PostgresStorage) GetStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Количество пользователей
	var userCount int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		return nil, err
	}
	stats["user_count"] = userCount

	// Количество секретов
	var secretCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM secrets").Scan(&secretCount)
	if err != nil {
		return nil, err
	}
	stats["secret_count"] = secretCount

	// Распределение по типам данных
	typeStats := make(map[string]int)
	rows, err := s.db.Query("SELECT type, COUNT(*) FROM secrets GROUP BY type")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var dataType string
		var count int
		if err := rows.Scan(&dataType, &count); err != nil {
			return nil, err
		}
		typeStats[dataType] = count
	}
	stats["type_distribution"] = typeStats

	return stats, nil
}
