package common

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNow(t *testing.T) {
	before := time.Now().UTC()
	now := Now()
	after := time.Now().UTC()

	assert.True(t, now.After(before) || now.Equal(before))
	assert.True(t, now.Before(after) || now.Equal(after))
}

func TestParseUUID(t *testing.T) {
	testUUID := uuid.New()

	// Valid UUID
	parsed, err := ParseUUID(testUUID.String())
	assert.NoError(t, err)
	assert.Equal(t, testUUID, parsed)

	// Invalid UUID
	_, err = ParseUUID("invalid-uuid")
	assert.Error(t, err)
}

func TestMustParseUUID(t *testing.T) {
	testUUID := uuid.New()

	// Valid UUID
	parsed := MustParseUUID(testUUID.String())
	assert.Equal(t, testUUID, parsed)

	// Should panic on invalid UUID
	assert.Panics(t, func() {
		MustParseUUID("invalid-uuid")
	})
}

func TestHashPasswordAndCheckPasswordHash(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)

	// Valid password
	assert.True(t, CheckPasswordHash(password, hash))

	// Invalid password
	assert.False(t, CheckPasswordHash("wrongpassword", hash))
}

func TestGenerateEncryptionKey(t *testing.T) {
	password := "testpassword"
	key := GenerateEncryptionKey(password)

	assert.Len(t, key, 32) // SHA256 produces 32-byte hash
	assert.NotEmpty(t, key)
}

func TestEncryptDecryptData(t *testing.T) {
	key := GenerateEncryptionKey("testpassword")
	originalData := []byte("sensitive data to encrypt")

	// Test encryption
	encrypted, err := EncryptData(originalData, key)
	require.NoError(t, err)
	assert.NotNil(t, encrypted)
	assert.NotEqual(t, originalData, encrypted)

	// Test decryption
	decrypted, err := DecryptData(encrypted, key)
	require.NoError(t, err)
	assert.Equal(t, originalData, decrypted)

	// Test decryption with wrong key
	wrongKey := GenerateEncryptionKey("wrongpassword")
	_, err = DecryptData(encrypted, wrongKey)
	assert.Error(t, err)

	// Test decryption with corrupted data
	_, err = DecryptData([]byte("tooshort"), key)
	assert.Error(t, err)
}

func TestGenerateRandomKey(t *testing.T) {
	key, err := GenerateRandomKey(32)
	require.NoError(t, err)
	assert.Len(t, key, 64) // hex encoded 32 bytes = 64 chars
}

func TestJWTTokenGenerationAndValidation(t *testing.T) {
	userID := uuid.New()
	login := "testuser"
	secret := "test-secret-key"

	// Generate token
	token, err := GenerateJWTToken(userID, login, secret, time.Hour)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	claims, err := ValidateJWTToken(token, secret)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, login, claims.Login)

	// Invalid token
	_, err = ValidateJWTToken("invalid.token", secret)
	assert.Error(t, err)

	// Wrong secret
	_, err = ValidateJWTToken(token, "wrong-secret")
	assert.Error(t, err)
}

func TestGetJWTClaims(t *testing.T) {
	userID := uuid.New()
	login := "testuser"
	secret := "test-secret-key"

	token, err := GenerateJWTToken(userID, login, secret, time.Hour)
	require.NoError(t, err)

	claims, err := GetJWTClaims(token)
	require.NoError(t, err)

	assert.Equal(t, login, claims["login"])
	assert.Equal(t, userID.String(), claims["user_id"])
}

func TestConflictManager(t *testing.T) {
	cm := NewConflictManager()

	secretID := uuid.New()
	conflict := Conflict{
		ID:         "test-conflict",
		SecretID:   secretID,
		Type:       ConflictBothModified,
		Reason:     "test reason",
		DetectedAt: Now(),
	}

	// Test adding conflict
	cm.AddConflict(conflict)
	assert.True(t, cm.HasConflictForSecret(secretID))
	assert.True(t, cm.HasPendingConflicts())

	// Test getting pending conflicts
	conflicts := cm.GetPendingConflicts()
	assert.Len(t, conflicts, 1)
	assert.Equal(t, "test-conflict", conflicts[0].ID)

	// Test getting conflict by secret ID
	foundConflict := cm.GetConflictBySecretID(secretID)
	assert.NotNil(t, foundConflict)
	assert.Equal(t, conflict.ID, foundConflict.ID)

	// Test resolving conflict
	resolution, err := cm.ResolveConflict("test-conflict", "keep_local")
	require.NoError(t, err)
	assert.Equal(t, "keep_local", resolution.Action)

	// Test conflict removal
	cm.RemoveConflict("test-conflict")
	assert.False(t, cm.HasConflictForSecret(secretID))
	assert.False(t, cm.HasPendingConflicts())
}

func TestDetectConflicts(t *testing.T) {
	cm := NewConflictManager()
	lastSync := Now().Add(-time.Hour)
	secretID := uuid.New()

	localSecrets := []SecretData{
		{
			ID:        secretID,
			Version:   2,
			UpdatedAt: Now(),
			Metadata:  "local version",
		},
	}

	remoteSecrets := []SecretData{
		{
			ID:        secretID,
			Version:   3,
			UpdatedAt: Now(),
			Metadata:  "remote version",
		},
	}

	conflicts := cm.DetectConflicts(localSecrets, remoteSecrets, lastSync)
	assert.Len(t, conflicts, 1)
	assert.Equal(t, ConflictBothModified, conflicts[0].Type)
	assert.Equal(t, secretID, conflicts[0].SecretID)
}

func TestCompareSecrets(t *testing.T) {
	local := &SecretData{
		Metadata:  "local name",
		Version:   1,
		UpdatedAt: Now().Add(-time.Hour),
	}

	remote := &SecretData{
		Metadata:  "remote name",
		Version:   2,
		UpdatedAt: Now(),
	}

	result := CompareSecrets(local, remote)
	assert.Contains(t, result, "Name: 'local name' vs 'remote name'")
	assert.Contains(t, result, "Version: 1 vs 2")
}

func TestOperationMarshaling(t *testing.T) {
	// Test operation request creation
	op := RegisterOp{
		Login:    "testuser",
		Password: "testpass",
	}

	req, err := CreateOperationRequest(OpRegister, op)
	require.NoError(t, err)
	assert.Equal(t, OpRegister, req.Type)
	assert.NotEmpty(t, req.Payload)

	// Test unmarshaling
	var unmarshaledOp RegisterOp
	err = UnmarshalOperation(req.Payload, &unmarshaledOp)
	require.NoError(t, err)
	assert.Equal(t, op.Login, unmarshaledOp.Login)
	assert.Equal(t, op.Password, unmarshaledOp.Password)

	// Test success response
	successResp, err := CreateSuccessResponse(op)
	require.NoError(t, err)
	assert.True(t, successResp.Success)
	assert.NotEmpty(t, successResp.Payload)

	// Test error response
	testErr := fmt.Errorf("test error")
	errorResp := CreateErrorResponse(testErr)
	assert.False(t, errorResp.Success)
	assert.Equal(t, "test error", errorResp.Error)
}
