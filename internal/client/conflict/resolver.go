package conflict

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"gophkeeper/internal/common"
)

// Resolver интерактивно разрешает конфликты
type Resolver struct {
	conflictManager *common.ConflictManager
	reader          *bufio.Reader
}

// NewResolver создает новый resolver
func NewResolver() *Resolver {
	return &Resolver{
		conflictManager: common.NewConflictManager(),
		reader:          bufio.NewReader(os.Stdin),
	}
}

// ResolveConflicts интерактивно разрешает конфликты
func (r *Resolver) ResolveConflicts(conflicts []common.Conflict) ([]common.ConflictResolution, error) {
	var resolutions []common.ConflictResolution

	for i, conflict := range conflicts {
		fmt.Printf("\n=== Conflict %d/%d ===\n", i+1, len(conflicts))

		// Автоматическое разрешение для простых случаев
		if autoResolution := r.tryAutoResolve(conflict); autoResolution != nil {
			fmt.Printf("Automatically resolved: %s\n", autoResolution.Action)
			resolutions = append(resolutions, *autoResolution)
			continue
		}

		// Интерактивное разрешение для сложных случаев
		resolution, err := r.resolveSingleConflict(conflict)
		if err != nil {
			return nil, err
		}

		if resolution != nil {
			resolutions = append(resolutions, *resolution)
		}
	}

	return resolutions, nil
}

// resolveSingleConflict разрешает один конфликт
func (r *Resolver) resolveSingleConflict(conflict common.Conflict) (*common.ConflictResolution, error) {
	// Показываем информацию о конфликте
	r.displayConflictInfo(conflict)

	// Предлагаем варианты разрешения
	action, err := r.promptForAction(conflict)
	if err != nil {
		return nil, err
	}

	if action == "skip" {
		fmt.Println("Skipping conflict for now...")
		return nil, nil
	}

	resolution, err := r.conflictManager.ResolveConflict(conflict.ID, action)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Resolved: chosen %s version\n", action)
	return resolution, nil
}

// displayConflictInfo показывает информацию о конфликте
func (r *Resolver) displayConflictInfo(conflict common.Conflict) {
	fmt.Printf("Conflict: %s\n", conflict.Reason)
	//fmt.Printf("Secret: %s (ID: %s)\n", conflict.LocalSecret.Metadata, conflict.SecretID)

	if conflict.LocalSecret != nil && conflict.RemoteSecret != nil {
		differences := common.CompareSecrets(conflict.LocalSecret, conflict.RemoteSecret)
		fmt.Println(differences)

		fmt.Println("\nLocal version:")
		r.displaySecretPreview(conflict.LocalSecret)

		fmt.Println("\nRemote version:")
		r.displaySecretPreview(conflict.RemoteSecret)
	}
}

// displaySecretPreview показывает превью секрета
func (r *Resolver) displaySecretPreview(secret *common.SecretData) {
	fmt.Printf("  Type: %s\n", secret.Type)
	fmt.Printf("  Version: %d\n", secret.Version)
	fmt.Printf("  Last modified: %s\n", secret.UpdatedAt.Format("2006-01-02 15:04"))
	fmt.Printf("  Size: %d bytes\n", len(secret.Data))

	// Для текстовых данных можем показать превью
	if secret.Type == common.TextDataType {
		// В реальной реализации здесь была бы попытка расшифровки
		fmt.Printf("  Preview: [encrypted text data]\n")
	}
}

// promptForAction запрашивает действие у пользователя
func (r *Resolver) promptForAction(conflict common.Conflict) (string, error) {
	fmt.Println("\nHow would you like to resolve this conflict?")
	fmt.Println("  1. Keep local version")
	fmt.Println("  2. Keep remote version")
	fmt.Println("  3. Keep newer version")
	fmt.Println("  4. Skip for now")
	fmt.Println("  5. Show differences in detail")

	if conflict.LocalSecret != nil && conflict.RemoteSecret != nil {
		fmt.Println("  6. Compare side by side")
	}

	fmt.Print("Choose option (1-6): ")

	input, err := r.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)

	switch input {
	case "1":
		return "keep_local", nil
	case "2":
		return "keep_remote", nil
	case "3":
		return "keep_newer", nil
	case "4":
		return "skip", nil
	case "5":
		r.showDetailedDifferences(conflict)
		return r.promptForAction(conflict) // Рекурсивно запрашиваем снова
	case "6":
		if conflict.LocalSecret != nil && conflict.RemoteSecret != nil {
			r.showSideBySide(conflict)
			return r.promptForAction(conflict)
		}
		fallthrough
	default:
		fmt.Println("Invalid option, please try again")
		return r.promptForAction(conflict)
	}
}

// showDetailedDifferences показывает детальные различия
func (r *Resolver) showDetailedDifferences(conflict common.Conflict) {
	fmt.Println("\n=== Detailed Differences ===")

	if conflict.LocalSecret == nil {
		fmt.Println("Local: DELETED")
		fmt.Println("Remote: EXISTS")
		return
	}

	if conflict.RemoteSecret == nil {
		fmt.Println("Local: EXISTS")
		fmt.Println("Remote: DELETED")
		return
	}

	fmt.Printf("Name: Local='%s' vs Remote='%s'\n",
		conflict.LocalSecret.Metadata, conflict.RemoteSecret.Metadata)
	fmt.Printf("Type: Local=%s vs Remote=%s\n",
		conflict.LocalSecret.Type, conflict.RemoteSecret.Type)
	fmt.Printf("Version: Local=v%d vs Remote=v%d\n",
		conflict.LocalSecret.Version, conflict.RemoteSecret.Version)
	fmt.Printf("Last modified: Local=%s vs Remote=%s\n",
		conflict.LocalSecret.UpdatedAt.Format("2006-01-02 15:04:05"),
		conflict.RemoteSecret.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Data size: Local=%d bytes vs Remote=%d bytes\n",
		len(conflict.LocalSecret.Data), len(conflict.RemoteSecret.Data))
}

// showSideBySide показывает сравнение бок о бок
func (r *Resolver) showSideBySide(conflict common.Conflict) {
	fmt.Println("\n=== Side by Side Comparison ===")
	fmt.Printf("%-20s | %-20s\n", "LOCAL", "REMOTE")
	fmt.Printf("%-20s | %-20s\n", "--------------------", "--------------------")
	fmt.Printf("%-20s | %-20s\n",
		conflict.LocalSecret.Metadata,
		conflict.RemoteSecret.Metadata)
	fmt.Printf("%-20s | %-20s\n",
		string(conflict.LocalSecret.Type),
		string(conflict.RemoteSecret.Type))
	fmt.Printf("%-20s | %-20s\n",
		fmt.Sprintf("v%d", conflict.LocalSecret.Version),
		fmt.Sprintf("v%d", conflict.RemoteSecret.Version))
	fmt.Printf("%-20s | %-20s\n",
		conflict.LocalSecret.UpdatedAt.Format("01/02 15:04"),
		conflict.RemoteSecret.UpdatedAt.Format("01/02 15:04"))
}

// GetPendingResolutions возвращает pending разрешения
func (r *Resolver) GetPendingResolutions() []common.ConflictResolution {
	// В этой реализации мы разрешаем конфликты сразу
	return []common.ConflictResolution{}
}

// tryAutoResolve пытается автоматически разрешить конфликт
func (r *Resolver) tryAutoResolve(conflict common.Conflict) *common.ConflictResolution {
	// Если одна из версий удалена, а другая изменена
	if conflict.LocalSecret == nil && conflict.RemoteSecret != nil {
		// Локально удалено, на сервере изменено - спрашиваем пользователя
		return nil
	}

	if conflict.LocalSecret != nil && conflict.RemoteSecret == nil {
		// На сервере удалено, локально изменено - спрашиваем пользователя
		return nil
	}

	// Если версии идентичны, выбираем любую
	if conflict.LocalSecret != nil && conflict.RemoteSecret != nil {
		if conflict.LocalSecret.Version == conflict.RemoteSecret.Version &&
			bytes.Equal(conflict.LocalSecret.Data, conflict.RemoteSecret.Data) &&
			conflict.LocalSecret.Metadata == conflict.RemoteSecret.Metadata {

			return &common.ConflictResolution{
				ConflictID: conflict.ID,
				Winner:     conflict.LocalSecret, // можно выбрать любую
				Action:     "auto_identical",
				ResolvedAt: time.Now(),
			}
		}

		// Если серверная версия значительно новее, выбираем её
		serverNewer := conflict.RemoteSecret.UpdatedAt.Sub(conflict.LocalSecret.UpdatedAt) > time.Hour
		if serverNewer && conflict.RemoteSecret.Version > conflict.LocalSecret.Version+1 {
			return &common.ConflictResolution{
				ConflictID: conflict.ID,
				Winner:     conflict.RemoteSecret,
				Action:     "auto_server_newer",
				ResolvedAt: time.Now(),
			}
		}
	}

	return nil
}
