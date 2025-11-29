package crypto

import (
	"testing"

	"gophkeeper/internal/common"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClientCrypto(t *testing.T) {
	crypto := NewClientCrypto("testpassword")
	assert.NotNil(t, crypto)
	assert.NotNil(t, crypto.masterKey)
	assert.Len(t, crypto.masterKey, 32)
}

func TestEncryptDecryptData(t *testing.T) {
	crypto := NewClientCrypto("testpassword")

	loginData := common.LoginPasswordData{
		Login:    "testuser",
		Password: "testpass",
		Site:     "example.com",
	}

	secret, err := crypto.EncryptData(common.LoginPasswordType, loginData, "test login")
	require.NoError(t, err)
	assert.NotNil(t, secret)
	assert.Equal(t, common.LoginPasswordType, secret.Type)
	assert.Equal(t, "test login", secret.Metadata)
	assert.NotEmpty(t, secret.Data)

	var decryptedData common.LoginPasswordData
	err = crypto.DecryptData(secret, &decryptedData)
	require.NoError(t, err)
	assert.Equal(t, loginData.Login, decryptedData.Login)
	assert.Equal(t, loginData.Password, decryptedData.Password)
	assert.Equal(t, loginData.Site, decryptedData.Site)
}

func TestEncryptDecryptCardData(t *testing.T) {
	crypto := NewClientCrypto("testpassword")

	cardData := common.CardData{
		Number: "4111111111111111",
		Expiry: "12/25",
		CVV:    "123",
		Holder: "John Doe",
		Bank:   "Test Bank",
	}

	secret, err := crypto.EncryptData(common.CardDataType, cardData, "test card")
	require.NoError(t, err)

	var decryptedData common.CardData
	err = crypto.DecryptData(secret, &decryptedData)
	require.NoError(t, err)
	assert.Equal(t, cardData.Number, decryptedData.Number)
	assert.Equal(t, cardData.Expiry, decryptedData.Expiry)
	assert.Equal(t, cardData.CVV, decryptedData.CVV)
}

func TestEncryptDecryptTextData(t *testing.T) {
	crypto := NewClientCrypto("testpassword")

	textData := "This is some sensitive text data"

	secret, err := crypto.EncryptData(common.TextDataType, textData, "test text")
	require.NoError(t, err)

	var decryptedData string
	err = crypto.DecryptData(secret, &decryptedData)
	require.NoError(t, err)
	assert.Equal(t, textData, decryptedData)
}

func TestDecryptWithWrongPassword(t *testing.T) {
	crypto1 := NewClientCrypto("password1")
	crypto2 := NewClientCrypto("password2")

	loginData := common.LoginPasswordData{
		Login:    "testuser",
		Password: "testpass",
	}

	secret, err := crypto1.EncryptData(common.LoginPasswordType, loginData, "test login")
	require.NoError(t, err)

	var decryptedData common.LoginPasswordData
	err = crypto2.DecryptData(secret, &decryptedData)
	assert.Error(t, err)
}

func TestChangeMasterKey(t *testing.T) {
	oldPassword := "oldpassword"
	newPassword := "newpassword"

	crypto := NewClientCrypto(oldPassword)

	loginData := common.LoginPasswordData{
		Login:    "testuser",
		Password: "testpass",
	}

	secret, err := crypto.EncryptData(common.LoginPasswordType, loginData, "test login")
	require.NoError(t, err)

	secrets := []*common.SecretData{secret}

	reencryptedSecrets, err := crypto.ChangeMasterKey(newPassword, secrets)
	require.NoError(t, err)
	assert.Len(t, reencryptedSecrets, 1)

	var decryptedData common.LoginPasswordData
	err = crypto.DecryptData(reencryptedSecrets[0], &decryptedData)
	require.NoError(t, err)
	assert.Equal(t, loginData.Login, decryptedData.Login)
	assert.Equal(t, loginData.Password, decryptedData.Password)
}
