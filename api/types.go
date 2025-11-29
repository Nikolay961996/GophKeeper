// Package api and proto functions
package api

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"gophkeeper/internal/common"
)

// ConvertToProto преобразует common.SecretData в protobuf SecretData
func ConvertToProto(secret *common.SecretData) *SecretData {
	return &SecretData{
		Id:        secret.ID.String(),
		UserId:    secret.UserID.String(),
		Type:      string(secret.Type),
		Name:      secret.Metadata,
		Data:      secret.Data,
		Metadata:  secret.Metadata,
		Version:   int32(secret.Version),
		CreatedAt: timestamppb.New(secret.CreatedAt),
		UpdatedAt: timestamppb.New(secret.UpdatedAt),
	}
}

// ConvertFromProto преобразует protobuf SecretData в common.SecretData
func ConvertFromProto(protoSecret *SecretData) (*common.SecretData, error) {
	id, err := common.ParseUUID(protoSecret.Id)
	if err != nil {
		return nil, err
	}

	userID, err := common.ParseUUID(protoSecret.UserId)
	if err != nil {
		return nil, err
	}

	return &common.SecretData{
		ID:        id,
		UserID:    userID,
		Type:      common.DataType(protoSecret.Type),
		Metadata:  protoSecret.Metadata,
		Data:      protoSecret.Data,
		Version:   int(protoSecret.Version),
		CreatedAt: protoSecret.CreatedAt.AsTime(),
		UpdatedAt: protoSecret.UpdatedAt.AsTime(),
	}, nil
}

// ConvertUserToProto преобразует common.User в protobuf User
func ConvertUserToProto(user *common.User) *User {
	return &User{
		Id:        user.ID.String(),
		Login:     user.Login,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}
}

// TimestampToTime преобразует protobuf Timestamp в time.Time
func TimestampToTime(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime()
}

// TimeToTimestamp преобразует time.Time в protobuf Timestamp
func TimeToTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
