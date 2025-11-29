package api

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"gophkeeper/internal/common"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestConvertToProto(t *testing.T) {
	secret := &common.SecretData{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Type:      common.LoginPasswordType,
		Name:      "Test Secret",
		Metadata:  "test metadata",
		Data:      []byte("test data"),
		Version:   1,
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}

	protoSecret := ConvertToProto(secret)

	assert.Equal(t, secret.ID.String(), protoSecret.Id)
	assert.Equal(t, secret.UserID.String(), protoSecret.UserId)
	assert.Equal(t, string(secret.Type), protoSecret.Type)
	assert.Equal(t, secret.Metadata, protoSecret.Name) // Note: Metadata becomes Name
	assert.Equal(t, secret.Data, protoSecret.Data)
	assert.Equal(t, secret.Metadata, protoSecret.Metadata)
	assert.Equal(t, int32(secret.Version), protoSecret.Version)
	assert.True(t, protoSecret.CreatedAt.IsValid())
	assert.True(t, protoSecret.UpdatedAt.IsValid())
}

func TestConvertFromProto(t *testing.T) {
	protoSecret := &SecretData{
		Id:        uuid.New().String(),
		UserId:    uuid.New().String(),
		Type:      string(common.LoginPasswordType),
		Name:      "Test Name",
		Data:      []byte("test data"),
		Metadata:  "test metadata",
		Version:   1,
		CreatedAt: timestamppb.New(time.Now().Add(-time.Hour)),
		UpdatedAt: timestamppb.New(time.Now()),
	}

	secret, err := ConvertFromProto(protoSecret)
	require.NoError(t, err)

	assert.Equal(t, protoSecret.Id, secret.ID.String())
	assert.Equal(t, protoSecret.UserId, secret.UserID.String())
	assert.Equal(t, common.DataType(protoSecret.Type), secret.Type)
	assert.Equal(t, protoSecret.Metadata, secret.Metadata)
	assert.Equal(t, protoSecret.Data, secret.Data)
	assert.Equal(t, int(protoSecret.Version), secret.Version)
	assert.True(t, protoSecret.CreatedAt.AsTime().Equal(secret.CreatedAt))
	assert.True(t, protoSecret.UpdatedAt.AsTime().Equal(secret.UpdatedAt))
}

func TestConvertFromProto_InvalidUUID(t *testing.T) {
	protoSecret := &SecretData{
		Id:     "invalid-uuid",
		UserId: uuid.New().String(),
		Type:   string(common.LoginPasswordType),
	}

	_, err := ConvertFromProto(protoSecret)
	assert.Error(t, err)
}

func TestConvertUserToProto(t *testing.T) {
	user := &common.User{
		ID:        uuid.New(),
		Login:     "testuser",
		CreatedAt: time.Now(),
	}

	protoUser := ConvertUserToProto(user)

	assert.Equal(t, user.ID.String(), protoUser.Id)
	assert.Equal(t, user.Login, protoUser.Login)
	assert.True(t, protoUser.CreatedAt.IsValid())
}

func TestTimestampConversions(t *testing.T) {
	testTime := time.Now()

	// Time to Timestamp
	timestamp := TimeToTimestamp(testTime)
	assert.True(t, timestamp.IsValid())
	assert.True(t, timestamp.AsTime().Equal(testTime))

	// Timestamp to Time
	convertedTime := TimestampToTime(timestamp)
	assert.True(t, convertedTime.Equal(testTime))

	// Nil timestamp
	nilTime := TimestampToTime(nil)
	assert.True(t, nilTime.IsZero())
}
