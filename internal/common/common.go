package common

import "time"

// Now возвращает текущее время в UTC
func Now() time.Time {
	return time.Now().UTC()
}

// TokenExpiration время жизни токена
const TokenExpiration = 24 * time.Hour
