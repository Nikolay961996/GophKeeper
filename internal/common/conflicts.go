package common

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"time"
)

// ConflictType тип конфликта
type ConflictType string

const (
	ConflictBothModified ConflictType = "both_modified" // Обе версии изменены
	ConflictDeleted      ConflictType = "deleted"       // Одна версия удалена
	ConflictVersion      ConflictType = "version"       // Конфликт версий
)

// Conflict представляет конфликт данных
type Conflict struct {
	ID           string       `json:"id"`
	SecretID     uuid.UUID    `json:"secret_id"`
	Type         ConflictType `json:"type"`
	LocalSecret  *SecretData  `json:"local_secret,omitempty"`  // Локальная версия
	RemoteSecret *SecretData  `json:"remote_secret,omitempty"` // Серверная версия
	Reason       string       `json:"reason"`
	DetectedAt   time.Time    `json:"detected_at"`
}

// ConflictResolution разрешение конфликта
type ConflictResolution struct {
	ConflictID string      `json:"conflict_id"`
	Winner     *SecretData `json:"winner"` // Какая версия побеждает
	Action     string      `json:"action"` // keep_local, keep_remote, merge, cancel
	ResolvedAt time.Time   `json:"resolved_at"`
}

// ConflictManager управляет конфликтами
type ConflictManager struct {
	pendingConflicts map[string]*Conflict
}

// NewConflictManager создает новый менеджер конфликтов
func NewConflictManager() *ConflictManager {
	return &ConflictManager{
		pendingConflicts: make(map[string]*Conflict),
	}
}

// DetectConflicts обнаруживает конфликты между локальными и серверными данными
func (cm *ConflictManager) DetectConflicts(localSecrets, remoteSecrets []SecretData, lastSync time.Time) []Conflict {
	var conflicts []Conflict

	localMap := make(map[uuid.UUID]SecretData)
	for _, secret := range localSecrets {
		localMap[secret.ID] = secret
	}

	remoteMap := make(map[uuid.UUID]SecretData)
	for _, secret := range remoteSecrets {
		remoteMap[secret.ID] = secret
	}

	// Проверяем конфликты для существующих записей
	for id, local := range localMap {
		remote, exists := remoteMap[id]
		if !exists {
			continue
		}

		// Обе версии изменялись после последней синхронизации
		if local.UpdatedAt.After(lastSync) && remote.UpdatedAt.After(lastSync) &&
			local.Version != remote.Version {

			conflict := Conflict{
				ID:           fmt.Sprintf("%s_%d_%d", id, local.Version, remote.Version),
				SecretID:     id,
				Type:         ConflictBothModified,
				LocalSecret:  &local,
				RemoteSecret: &remote,
				Reason:       fmt.Sprintf("Both versions modified after last sync (local v%d, remote v%d)", local.Version, remote.Version),
				DetectedAt:   Now(),
			}
			conflicts = append(conflicts, conflict)
		}
	}

	// Проверяем удаленные записи
	for id, remote := range remoteMap {
		local, exists := localMap[id]
		if !exists && remote.UpdatedAt.After(lastSync) {
			// Запись удалена локально, но изменена на сервере
			conflict := Conflict{
				ID:           fmt.Sprintf("%s_deleted", id),
				SecretID:     id,
				Type:         ConflictDeleted,
				LocalSecret:  nil, // Локально удалено
				RemoteSecret: &remote,
				Reason:       "Locally deleted but modified on server",
				DetectedAt:   Now(),
			}
			conflicts = append(conflicts, conflict)
		}
	}

	return conflicts
}

// AddConflict добавляет конфликт в pending
func (cm *ConflictManager) AddConflict(conflict Conflict) {
	cm.pendingConflicts[conflict.ID] = &conflict
}

// ResolveConflict разрешает конфликт
func (cm *ConflictManager) ResolveConflict(conflictID string, action string) (*ConflictResolution, error) {
	conflict, exists := cm.pendingConflicts[conflictID]
	if !exists {
		return nil, fmt.Errorf("conflict not found: %s", conflictID)
	}

	var winner *SecretData
	switch action {
	case "keep_local":
		winner = conflict.LocalSecret
	case "keep_remote":
		winner = conflict.RemoteSecret
	case "keep_newer":
		if conflict.LocalSecret.UpdatedAt.After(conflict.RemoteSecret.UpdatedAt) {
			winner = conflict.LocalSecret
		} else {
			winner = conflict.RemoteSecret
		}
	case "cancel":
		// Отмена - не сохранять ничего
		winner = nil
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}

	resolution := &ConflictResolution{
		ConflictID: conflictID,
		Winner:     winner,
		Action:     action,
		ResolvedAt: Now(),
	}

	// Удаляем из pending
	delete(cm.pendingConflicts, conflictID)

	return resolution, nil
}

// GetPendingConflicts возвращает pending конфликты
func (cm *ConflictManager) GetPendingConflicts() []Conflict {
	var conflicts []Conflict
	for _, conflict := range cm.pendingConflicts {
		conflicts = append(conflicts, *conflict)
	}
	return conflicts
}

// HasPendingConflicts проверяет есть ли pending конфликты
func (cm *ConflictManager) HasPendingConflicts() bool {
	return len(cm.pendingConflicts) > 0
}

// CompareSecrets сравнивает два секрета (для отображения различий)
func CompareSecrets(local, remote *SecretData) string {
	var differences []string

	if local.Metadata != remote.Metadata {
		differences = append(differences, fmt.Sprintf("Name: '%s' vs '%s'", local.Metadata, remote.Metadata))
	}

	if local.Version != remote.Version {
		differences = append(differences, fmt.Sprintf("Version: %d vs %d", local.Version, remote.Version))
	}

	if !local.UpdatedAt.Equal(remote.UpdatedAt) {
		differences = append(differences, fmt.Sprintf("Updated: %s vs %s",
			local.UpdatedAt.Format("2006-01-02 15:04"),
			remote.UpdatedAt.Format("2006-01-02 15:04")))
	}

	if len(differences) == 0 {
		return "No visible differences (encrypted content may differ)"
	}

	result := "Differences:\n"
	for i, diff := range differences {
		result += fmt.Sprintf("  %d. %s\n", i+1, diff)
	}
	return result
}
