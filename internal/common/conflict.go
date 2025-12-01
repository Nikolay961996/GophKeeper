package common

import (
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
	DetectedAt   time.Time    `json:"detected_at"`
	LocalSecret  *SecretData  `json:"local_secret,omitempty"`
	RemoteSecret *SecretData  `json:"remote_secret,omitempty"`
	ID           string       `json:"id"`
	Type         ConflictType `json:"type"`
	Reason       string       `json:"reason"`
	SecretID     uuid.UUID    `json:"secret_id"`
}

// ConflictResolution разрешение конфликта
type ConflictResolution struct {
	ResolvedAt time.Time   `json:"resolved_at"`
	Winner     *SecretData `json:"winner"`
	ConflictID string      `json:"conflict_id"`
	Action     string      `json:"action"`
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
		_, exists := localMap[id]
		if !exists && remote.UpdatedAt.After(lastSync) {
			// Запись удалена локально, но изменена на сервере
			conflict := Conflict{
				ID:           fmt.Sprintf("%s_deleted", id),
				SecretID:     id,
				Type:         ConflictDeleted,
				LocalSecret:  nil,
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

// HasConflictForSecret проверяет есть ли конфликт для конкретного секрета
func (cm *ConflictManager) HasConflictForSecret(secretID uuid.UUID) bool {
	for _, conflict := range cm.pendingConflicts {
		if conflict.SecretID == secretID {
			return true
		}
	}
	return false
}

// RemoveConflict удаляет конфликт
func (cm *ConflictManager) RemoveConflict(conflictID string) {
	delete(cm.pendingConflicts, conflictID)
}

// GetConflictBySecretID возвращает конфликт по ID секрета
func (cm *ConflictManager) GetConflictBySecretID(secretID uuid.UUID) *Conflict {
	for _, conflict := range cm.pendingConflicts {
		if conflict.SecretID == secretID {
			return conflict
		}
	}
	return nil
}
