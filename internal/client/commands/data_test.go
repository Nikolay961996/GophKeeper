package commands

import (
	"github.com/google/uuid"
	"testing"
	"time"

	"gophkeeper/internal/client/config"
	"gophkeeper/internal/common"

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
	GetCardDataFunc       func(id string) (*common.CardData, error)
	GetTextDataFunc       func(id string) (string, error)
	GetBinaryDataFunc     func(id string) (string, []byte, error)
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
	if m.GetCardDataFunc != nil {
		return m.GetCardDataFunc(id)
	}
	return nil, nil
}

func (m *MockDataManager) GetTextData(id string) (string, error) {
	if m.GetTextDataFunc != nil {
		return m.GetTextDataFunc(id)
	}
	return "", nil
}

func (m *MockDataManager) GetBinaryData(id string) (string, []byte, error) {
	if m.GetBinaryDataFunc != nil {
		return m.GetBinaryDataFunc(id)
	}
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

func TestAllDataCommandMethods(_ *testing.T) {
	cfg := &config.Config{Token: "test-token"}

	mockManager := &MockDataManager{
		SaveLoginPasswordFunc: func(name, login, password, site string) error { return nil },
		SaveCardDataFunc:      func(name, number, expiry, cvv, holder, bank string) error { return nil },
		SaveTextDataFunc:      func(name, text string) error { return nil },
		SaveBinaryDataFunc:    func(name string, data []byte, fileName string) error { return nil },
		ListDataFunc:          func() []*common.SecretData { return []*common.SecretData{} },
		GetSecretByIDFunc:     func(id string) *common.SecretData { return nil },
		DeleteDataFunc:        func(id string) error { return nil },
		SaveSecretFunc:        func(secret *common.SecretData) error { return nil },
		GetLoginPasswordFunc:  func(id string) (*common.LoginPasswordData, error) { return nil, nil },
		GetCardDataFunc:       func(id string) (*common.CardData, error) { return nil, nil },
		GetTextDataFunc:       func(id string) (string, error) { return "", nil },
		GetBinaryDataFunc:     func(id string) (string, []byte, error) { return "", nil, nil },
	}

	mockClient := &MockGRPCClient{
		SyncFunc: func(lastSync time.Time, data []common.SecretData) (*common.SyncResult, error) {
			return &common.SyncResult{}, nil
		},
	}

	dataCommands := NewDataCommands(cfg, mockManager, mockClient)

	_ = dataCommands.AddLoginPassword("name", "login", "pass", "site")
	_ = dataCommands.AddCard("card", "1111", "12/25", "123", "holder", "bank")
	_ = dataCommands.AddText("text", "content")
	_ = dataCommands.AddFile("file", "path")
	_ = dataCommands.List()
	_ = dataCommands.Get(uuid.New().String())
	_ = dataCommands.Delete(uuid.New().String())
	_ = dataCommands.Sync()

	_ = dataCommands.checkLocalFileExists(uuid.New())

	conflicts := []common.Conflict{}
	_, _ = dataCommands.resolveConflictsInteractively(conflicts)
	_ = dataCommands.applyConflictResolutions([]common.ConflictResolution{})
	_ = dataCommands.filterNonConflictData([]common.SecretData{}, conflicts)
	_ = dataCommands.filterLocalDataForSync([]common.SecretData{}, conflicts)
}

func TestDataCommands_AllMethods(_ *testing.T) {
	cfg := &config.Config{Token: "test-token"}

	mockManager := &MockDataManager{
		SaveLoginPasswordFunc: func(name, login, password, site string) error { return nil },
		SaveCardDataFunc:      func(name, number, expiry, cvv, holder, bank string) error { return nil },
		SaveTextDataFunc:      func(name, text string) error { return nil },
		SaveBinaryDataFunc:    func(name string, data []byte, fileName string) error { return nil },
		ListDataFunc:          func() []*common.SecretData { return []*common.SecretData{} },
		GetSecretByIDFunc:     func(id string) *common.SecretData { return nil },
		DeleteDataFunc:        func(id string) error { return nil },
		SaveSecretFunc:        func(secret *common.SecretData) error { return nil },
		GetLoginPasswordFunc: func(id string) (*common.LoginPasswordData, error) {
			return &common.LoginPasswordData{}, nil
		},
		GetCardDataFunc:   func(id string) (*common.CardData, error) { return &common.CardData{}, nil },
		GetTextDataFunc:   func(id string) (string, error) { return "text", nil },
		GetBinaryDataFunc: func(id string) (string, []byte, error) { return "file", []byte("data"), nil },
	}

	mockClient := &MockGRPCClient{
		SyncFunc: func(lastSync time.Time, data []common.SecretData) (*common.SyncResult, error) {
			return &common.SyncResult{}, nil
		},
	}

	dataCommands := NewDataCommands(cfg, mockManager, mockClient)

	_ = dataCommands.AddLoginPassword("name", "login", "pass", "site")
	_ = dataCommands.AddCard("card", "1111", "12/25", "123", "holder", "bank")
	_ = dataCommands.AddText("text", "content")
	_ = dataCommands.AddFile("file", "path")
	_ = dataCommands.List()
	_ = dataCommands.Get(uuid.New().String())
	_ = dataCommands.Delete(uuid.New().String())
	_ = dataCommands.Sync()
	_ = dataCommands.checkLocalFileExists(uuid.New())
}

func TestDataCommands_HelperMethods(_ *testing.T) {
	cfg := &config.Config{Token: "test-token"}
	mockManager := &MockDataManager{
		GetLoginPasswordFunc: func(id string) (*common.LoginPasswordData, error) {
			return &common.LoginPasswordData{}, nil
		},
		GetCardDataFunc: func(id string) (*common.CardData, error) { return &common.CardData{}, nil },
		GetTextDataFunc: func(id string) (string, error) { return "", nil },
		GetBinaryDataFunc: func(id string) (string, []byte, error) {
			return "", nil, nil
		},
	}
	mockClient := &MockGRPCClient{}

	dataCommands := NewDataCommands(cfg, mockManager, mockClient)

	conflicts := []common.Conflict{}
	_ = dataCommands.filterNonConflictData([]common.SecretData{}, conflicts)
	_ = dataCommands.filterLocalDataForSync([]common.SecretData{}, conflicts)
	_ = dataCommands.applyConflictResolutions([]common.ConflictResolution{})

	secretLogin := &common.SecretData{Type: common.LoginPasswordType}
	_ = dataCommands.printSecret(secretLogin)

	secretCard := &common.SecretData{Type: common.CardDataType}
	_ = dataCommands.printSecret(secretCard)

	secretText := &common.SecretData{Type: common.TextDataType}
	_ = dataCommands.printSecret(secretText)

	secretUnknown := &common.SecretData{Type: "unknown"}
	_ = dataCommands.printSecret(secretUnknown)
}
