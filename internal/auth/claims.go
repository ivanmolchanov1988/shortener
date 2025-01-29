package auth

import (
	"github.com/golang-jwt/jwt/v4"
)

// Claims представляет данные токена JWT.
type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}
