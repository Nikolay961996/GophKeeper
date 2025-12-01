package common

import (
	"encoding/json"
	"fmt"
	"time"
)

// OperationType тип операции
type OperationType string

const (
	OpRegister OperationType = "register"
	OpLogin    OperationType = "login"
	OpSync     OperationType = "sync"
)

// OperationRequest запрос операции
type OperationRequest struct {
	Type    OperationType `json:"type"`
	Payload []byte        `json:"payload"`
}

// OperationResponse ответ операции
type OperationResponse struct {
	Error   string `json:"error,omitempty"`
	Payload []byte `json:"payload,omitempty"`
	Success bool   `json:"success"`
}

// RegisterOp данные регистрации
type RegisterOp struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// LoginOp данные входа
type LoginOp struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// SyncOp данные синхронизации
type SyncOp struct {
	LastSync time.Time    `json:"last_sync"`
	Data     []SecretData `json:"data"`
}

// SyncResult результат синхронизации
type SyncResult struct {
	LastSync  time.Time    `json:"last_sync"`
	Data      []SecretData `json:"data"`
	Conflicts []SecretData `json:"conflicts,omitempty"`
}

// AuthResult результат аутентификации
type AuthResult struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// MarshalOperation маршалит операцию в JSON
func MarshalOperation(op interface{}) ([]byte, error) {
	return json.Marshal(op)
}

// UnmarshalOperation анмаршалит операцию из JSON
func UnmarshalOperation(data []byte, op interface{}) error {
	return json.Unmarshal(data, op)
}

// CreateOperationRequest создает запрос операции
func CreateOperationRequest(opType OperationType, payload interface{}) (*OperationRequest, error) {
	payloadData, err := MarshalOperation(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	return &OperationRequest{
		Type:    opType,
		Payload: payloadData,
	}, nil
}

// CreateSuccessResponse создает успешный ответ
func CreateSuccessResponse(payload interface{}) (*OperationResponse, error) {
	var payloadData []byte
	if payload != nil {
		var err error
		payloadData, err = MarshalOperation(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %v", err)
		}
	}

	return &OperationResponse{
		Success: true,
		Payload: payloadData,
	}, nil
}

// CreateErrorResponse создает ответ с ошибкой
func CreateErrorResponse(err error) *OperationResponse {
	return &OperationResponse{
		Success: false,
		Error:   err.Error(),
	}
}
