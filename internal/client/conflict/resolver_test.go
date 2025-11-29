package conflict

import (
	"testing"
	"time"

	"gophkeeper/internal/common"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewResolver(t *testing.T) {
	resolver := NewResolver()
	assert.NotNil(t, resolver)
	assert.NotNil(t, resolver.conflictManager)
	assert.NotNil(t, resolver.reader)
}

func TestResolver_TryAutoResolve_Identical(t *testing.T) {
	resolver := NewResolver()

	secretID := uuid.New()
	secret := &common.SecretData{
		ID:        secretID,
		Version:   1,
		UpdatedAt: time.Now(),
		Metadata:  "test",
		Data:      []byte("data"),
	}

	conflict := common.Conflict{
		ID:           "test-conflict",
		SecretID:     secretID,
		Type:         common.ConflictBothModified,
		LocalSecret:  secret,
		RemoteSecret: secret, // идентичные секреты
		Reason:       "test",
		DetectedAt:   time.Now(),
	}

	resolution := resolver.tryAutoResolve(conflict)
	assert.NotNil(t, resolution)
	assert.Equal(t, "auto_identical", resolution.Action)
	assert.Equal(t, secret, resolution.Winner)
}

func TestResolver_TryAutoResolve_ServerNewer(t *testing.T) {
	resolver := NewResolver()

	secretID := uuid.New()
	localSecret := &common.SecretData{
		ID:        secretID,
		Version:   1,
		UpdatedAt: time.Now().Add(-2 * time.Hour),
		Metadata:  "local",
		Data:      []byte("local data"),
	}

	remoteSecret := &common.SecretData{
		ID:        secretID,
		Version:   3, // версия значительно новее
		UpdatedAt: time.Now(),
		Metadata:  "remote",
		Data:      []byte("remote data"),
	}

	conflict := common.Conflict{
		ID:           "test-conflict",
		SecretID:     secretID,
		Type:         common.ConflictBothModified,
		LocalSecret:  localSecret,
		RemoteSecret: remoteSecret,
		Reason:       "test",
		DetectedAt:   time.Now(),
	}

	resolution := resolver.tryAutoResolve(conflict)
	assert.NotNil(t, resolution)
	assert.Equal(t, "auto_server_newer", resolution.Action)
	assert.Equal(t, remoteSecret, resolution.Winner)
}

func TestResolver_DisplayConflictInfo(t *testing.T) {
	resolver := NewResolver()

	secretID := uuid.New()
	localSecret := &common.SecretData{
		ID:        secretID,
		Type:      common.LoginPasswordType,
		Version:   1,
		UpdatedAt: time.Now(),
		Metadata:  "local secret",
		Data:      []byte("local data"),
	}

	remoteSecret := &common.SecretData{
		ID:        secretID,
		Type:      common.LoginPasswordType,
		Version:   2,
		UpdatedAt: time.Now().Add(time.Hour),
		Metadata:  "remote secret",
		Data:      []byte("remote data"),
	}

	conflict := common.Conflict{
		ID:           "test-conflict",
		SecretID:     secretID,
		Type:         common.ConflictBothModified,
		LocalSecret:  localSecret,
		RemoteSecret: remoteSecret,
		Reason:       "Both versions modified",
		DetectedAt:   time.Now(),
	}

	// Этот тест в основном проверяет, что функция не паникует
	// В реальном использовании она выводит информацию в консоль
	resolver.displayConflictInfo(conflict)
}

func TestResolver_ShowDetailedDifferences(t *testing.T) {
	resolver := NewResolver()

	secretID := uuid.New()
	localSecret := &common.SecretData{
		ID:        secretID,
		Type:      common.LoginPasswordType,
		Version:   1,
		UpdatedAt: time.Now(),
		Metadata:  "local",
		Data:      []byte("data"),
	}

	remoteSecret := &common.SecretData{
		ID:        secretID,
		Type:      common.LoginPasswordType,
		Version:   2,
		UpdatedAt: time.Now().Add(time.Hour),
		Metadata:  "remote",
		Data:      []byte("data"),
	}

	conflict := common.Conflict{
		ID:           "test-conflict",
		SecretID:     secretID,
		Type:         common.ConflictBothModified,
		LocalSecret:  localSecret,
		RemoteSecret: remoteSecret,
		Reason:       "test",
		DetectedAt:   time.Now(),
	}

	// Проверяем, что функция не паникует
	resolver.showDetailedDifferences(conflict)
}
