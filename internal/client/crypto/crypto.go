// Package crypto contains encryption and decryption logics
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
	"gophkeeper/internal/common"
)

// ClientCrypto управляет шифрованием на клиенте
type ClientCrypto struct {
	masterKey []byte
}

// NewClientCrypto создает новый ClientCrypto
func NewClientCrypto(masterPassword string) *ClientCrypto {
	key := deriveKey(masterPassword)
	return &ClientCrypto{
		masterKey: key,
	}
}

// deriveKey создает ключ из мастер-пароля
func deriveKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

// EncryptData шифрует данные перед отправкой на сервер
func (c *ClientCrypto) EncryptData(dataType common.DataType, data interface{}, metadata string) (*common.SecretData, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	encryptedData, err := c.encrypt(jsonData)
	if err != nil {
		return nil, err
	}

	secretData := &common.SecretData{
		ID:        uuid.New(),
		Type:      dataType,
		Data:      encryptedData,
		Metadata:  metadata,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return secretData, nil
}

// DecryptData расшифровывает данные с сервера
func (c *ClientCrypto) DecryptData(secretData *common.SecretData, result interface{}) error {
	decryptedData, err := c.decrypt(secretData.Data)
	if err != nil {
		return err
	}

	return json.Unmarshal(decryptedData, result)
}

// encrypt шифрует данные с использованием AES-GCM
func (c *ClientCrypto) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt расшифровывает данные
func (c *ClientCrypto) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.masterKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// ChangeMasterKey изменяет мастер-ключ и перешифровывает данные
func (c *ClientCrypto) ChangeMasterKey(newMasterPassword string, secrets []*common.SecretData) ([]*common.SecretData, error) {
	newKey := deriveKey(newMasterPassword)
	var reencryptedSecrets []*common.SecretData

	for _, secret := range secrets {
		var temp interface{}
		if err := c.DecryptData(secret, &temp); err != nil {
			return nil, err
		}

		// Шифруем новым ключом
		oldKey := c.masterKey
		c.masterKey = newKey

		newSecret, err := c.EncryptData(secret.Type, temp, secret.Metadata)
		if err != nil {
			c.masterKey = oldKey
			return nil, err
		}

		newSecret.ID = secret.ID
		newSecret.UserID = secret.UserID
		newSecret.Version = secret.Version + 1
		newSecret.CreatedAt = secret.CreatedAt
		newSecret.UpdatedAt = time.Now()

		reencryptedSecrets = append(reencryptedSecrets, newSecret)
	}

	c.masterKey = newKey
	return reencryptedSecrets, nil
}
