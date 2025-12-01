package api

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"

	"gophkeeper/internal/common"

	"github.com/google/uuid"
)

func TestAllApiFunctions(_ *testing.T) {
	secret := &common.SecretData{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Type:      common.LoginPasswordType,
		Name:      "test",
		Metadata:  "metadata",
		Data:      []byte("data"),
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = ConvertToProto(secret)

	protoSecret := &SecretData{
		Id:        uuid.New().String(),
		UserId:    uuid.New().String(),
		Type:      string(common.LoginPasswordType),
		Name:      "name",
		Data:      []byte("data"),
		Metadata:  "metadata",
		Version:   1,
		CreatedAt: timestamppb.New(time.Now()),
		UpdatedAt: timestamppb.New(time.Now()),
	}
	_, _ = ConvertFromProto(protoSecret)

	protoSecretInvalid := &SecretData{
		Id:     "invalid-uuid",
		UserId: uuid.New().String(),
		Type:   string(common.LoginPasswordType),
	}
	_, _ = ConvertFromProto(protoSecretInvalid)

	user := &common.User{
		ID:        uuid.New(),
		Login:     "testuser",
		CreatedAt: time.Now(),
	}
	_ = ConvertUserToProto(user)

	_ = TimestampToTime(nil)
	_ = TimestampToTime(timestamppb.New(time.Now()))
	_ = TimeToTimestamp(time.Time{})
	_ = TimeToTimestamp(time.Now())
}
