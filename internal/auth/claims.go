// Package auth предоставляет механизмы аутентификации и авторизации пользователей.
// Содержит структуры и методы для работы с JWT-токенами и обработкой прав доступа.
package auth

import (
	"github.com/golang-jwt/jwt/v4"
)

// Claims представляет данные токена JWT.
type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}
