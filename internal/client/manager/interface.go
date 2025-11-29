package manager

import "gophkeeper/internal/common"

// ManagerInterface определяет контракт для менеджера данных
type ManagerInterface interface {
	SaveLoginPassword(name, login, password, site string) error
	SaveCardData(name, number, expiry, cvv, holder, bank string) error
	SaveTextData(name, text string) error
	SaveBinaryData(name string, data []byte, fileName string) error
	GetLoginPassword(id string) (*common.LoginPasswordData, error)
	GetCardData(id string) (*common.CardData, error)
	GetTextData(id string) (string, error)
	GetBinaryData(id string) (string, []byte, error)
	ListData() []*common.SecretData
	DeleteData(id string) error
	GetSecretByID(id string) *common.SecretData
	SaveSecret(secret *common.SecretData) error
}

var _ ManagerInterface = (*DataManager)(nil)
