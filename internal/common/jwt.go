package common

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"time"
)

// JWTClaims кастомные claims для JWT
type JWTClaims struct {
	jwt.RegisteredClaims
	Login  string    `json:"login"`
	UserID uuid.UUID `json:"user_id"`
}

var (
	ErrInvalidToken = errors.New("invalid token")
)

// GenerateJWTToken генерирует JWT токен
func GenerateJWTToken(userID uuid.UUID, login, secret string, expiration time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID: userID,
		Login:  login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWTToken валидирует JWT токен
func ValidateJWTToken(tokenString, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

func GetJWTClaims(tokenString string) (jwt.MapClaims, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		fmt.Printf("Ошибка парсинга токена: %v\n", err)
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Printf("Не удалось извлечь claims из токена: %v\n", err)
		return nil, err
	}

	return claims, nil
}
