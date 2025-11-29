package commands

import (
	"testing"
	"time"

	"gophkeeper/internal/client/config"
	"gophkeeper/internal/common"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDataManager для тестирования
type MockDataManager struct {
	SaveLoginPasswordFunc func(name, login, password, site string) error
	SaveCardDataFunc      func(name, number, expiry, cvv, holder, bank string) error
	SaveTextDataFunc      func(name, text string) error
	SaveBinaryDataFunc    func(name string, data []byte, fileName string) error
	ListDataFunc          func() []*common.SecretData
	GetSecretByIDFunc     func(id string) *common.SecretData
	DeleteDataFunc        func(id string) error
	SaveSecretFunc        func(secret *common.SecretData) error
	GetLoginPasswordFunc  func(id string) (*common.LoginPasswordData, error)
}

func (m *MockDataManager) SaveLoginPassword(name, login, password, site string) error {
	if m.SaveLoginPasswordFunc != nil {
		return m.SaveLoginPasswordFunc(name, login, password, site)
	}
	return nil
}

func (m *MockDataManager) SaveCardData(name, number, expiry, cvv, holder, bank string) error {
	if m.SaveCardDataFunc != nil {
		return m.SaveCardDataFunc(name, number, expiry, cvv, holder, bank)
	}
	return nil
}

func (m *MockDataManager) SaveTextData(name, text string) error {
	if m.SaveTextDataFunc != nil {
		return m.SaveTextDataFunc(name, text)
	}
	return nil
}

func (m *MockDataManager) SaveBinaryData(name string, data []byte, fileName string) error {
	if m.SaveBinaryDataFunc != nil {
		return m.SaveBinaryDataFunc(name, data, fileName)
	}
	return nil
}

func (m *MockDataManager) ListData() []*common.SecretData {
	if m.ListDataFunc != nil {
		return m.ListDataFunc()
	}
	return nil
}

func (m *MockDataManager) GetSecretByID(id string) *common.SecretData {
	if m.GetSecretByIDFunc != nil {
		return m.GetSecretByIDFunc(id)
	}
	return nil
}

func (m *MockDataManager) DeleteData(id string) error {
	if m.DeleteDataFunc != nil {
		return m.DeleteDataFunc(id)
	}
	return nil
}

func (m *MockDataManager) SaveSecret(secret *common.SecretData) error {
	if m.SaveSecretFunc != nil {
		return m.SaveSecretFunc(secret)
	}
	return nil
}

func (m *MockDataManager) GetLoginPassword(id string) (*common.LoginPasswordData, error) {
	if m.GetLoginPasswordFunc != nil {
		return m.GetLoginPasswordFunc(id)
	}
	return nil, nil
}

func (m *MockDataManager) GetCardData(id string) (*common.CardData, error) {
	return nil, nil
}

func (m *MockDataManager) GetTextData(id string) (string, error) {
	return "", nil
}

func (m *MockDataManager) GetBinaryData(id string) (string, []byte, error) {
	return "", nil, nil
}

func TestDataCommands_AddLoginPassword(t *testing.T) {
	cfg := &config.Config{Token: "test-token"}

	var saved bool
	mockManager := &MockDataManager{
		SaveLoginPasswordFunc: func(name, login, password, site string) error {
			assert.Equal(t, "test login", name)
			assert.Equal(t, "testuser", login)
			assert.Equal(t, "testpass", password)
			assert.Equal(t, "example.com", site)
			saved = true
			return nil
		},
	}

	mockClient := &MockGRPCClient{
		SyncFunc: func(lastSync time.Time, data []common.SecretData) (*common.SyncResult, error) {
			return &common.SyncResult{}, nil
		},
	}

	dataCommands := NewDataCommands(cfg, mockManager, mockClient)

	err := dataCommands.AddLoginPassword("test login", "testuser", "testpass", "example.com")
	require.NoError(t, err)
	assert.True(t, saved)
}

func TestDataCommands_AddCard(t *testing.T) {
	cfg := &config.Config{Token: "test-token"}

	var saved bool
	mockManager := &MockDataManager{
		SaveCardDataFunc: func(name, number, expiry, cvv, holder, bank string) error {
			assert.Equal(t, "test card", name)
			assert.Equal(t, "4111111111111111", number)
			assert.Equal(t, "12/25", expiry)
			assert.Equal(t, "123", cvv)
			assert.Equal(t, "John Doe", holder)
			assert.Equal(t, "Test Bank", bank)
			saved = true
			return nil
		},
	}

	mockClient := &MockGRPCClient{
		SyncFunc: func(lastSync time.Time, data []common.SecretData) (*common.SyncResult, error) {
			return &common.SyncResult{}, nil
		},
	}

	dataCommands := NewDataCommands(cfg, mockManager, mockClient)

	err := dataCommands.AddCard("test card", "4111111111111111", "12/25", "123", "John Doe", "Test Bank")
	require.NoError(t, err)
	assert.True(t, saved)
}

func TestDataCommands_AddText(t *testing.T) {
	cfg := &config.Config{Token: "test-token"}

	var saved bool
	mockManager := &MockDataManager{
		SaveTextDataFunc: func(name, text string) error {
			assert.Equal(t, "test text", name)
			assert.Equal(t, "This is some text data", text)
			saved = true
			return nil
		},
	}

	mockClient := &MockGRPCClient{
		SyncFunc: func(lastSync time.Time, data []common.SecretData) (*common.SyncResult, error) {
			return &common.SyncResult{}, nil
		},
	}

	dataCommands := NewDataCommands(cfg, mockManager, mockClient)

	err := dataCommands.AddText("test text", "This is some text data")
	require.NoError(t, err)
	assert.True(t, saved)
}

func TestDataCommands_List(t *testing.T) {
	cfg := &config.Config{Token: "test-token"}

	testSecrets := []*common.SecretData{
		{
			ID:        uuid.New(),
			Type:      common.LoginPasswordType,
			Metadata:  "Test Secret 1",
			Version:   1,
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Type:      common.CardDataType,
			Metadata:  "Test Secret 2",
			Version:   1,
			UpdatedAt: time.Now(),
		},
	}

	mockManager := &MockDataManager{
		ListDataFunc: func() []*common.SecretData {
			return testSecrets
		},
	}

	mockClient := &MockGRPCClient{}

	dataCommands := NewDataCommands(cfg, mockManager, mockClient)

	err := dataCommands.List()
	require.NoError(t, err)
}

func TestDataCommands_Get(t *testing.T) {
	cfg := &config.Config{Token: "test-token"}

	secretID := uuid.New()
	testSecret := &common.SecretData{
		ID:        secretID,
		Type:      common.LoginPasswordType,
		Metadata:  "Test Secret",
		Version:   1,
		UpdatedAt: time.Now(),
	}

	mockManager := &MockDataManager{
		GetSecretByIDFunc: func(id string) *common.SecretData {
			assert.Equal(t, secretID.String(), id)
			return testSecret
		},
		GetLoginPasswordFunc: func(id string) (*common.LoginPasswordData, error) {
			return &common.LoginPasswordData{
				Login:    "testuser",
				Password: "testpass",
				Site:     "example.com",
			}, nil
		},
	}

	mockClient := &MockGRPCClient{}

	dataCommands := NewDataCommands(cfg, mockManager, mockClient)

	err := dataCommands.Get(secretID.String())
	require.NoError(t, err)
}

func TestDataCommands_Delete(t *testing.T) {
	cfg := &config.Config{Token: "test-token"}

	secretID := uuid.New()
	var deleted bool

	mockManager := &MockDataManager{
		DeleteDataFunc: func(id string) error {
			assert.Equal(t, secretID.String(), id)
			deleted = true
			return nil
		},
	}

	mockClient := &MockGRPCClient{}

	dataCommands := NewDataCommands(cfg, mockManager, mockClient)

	err := dataCommands.Delete(secretID.String())
	require.NoError(t, err)
	assert.True(t, deleted)
}
