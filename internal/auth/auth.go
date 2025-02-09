// Package auth предоставляет функции для аутентификации и управления сессиями пользователей.
package auth

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func buildJWTString(secret string, userID string, duration time.Duration) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Println("!!! Ошибка подписания JWT:", err)
		return "", fmt.Errorf("failed to build a token: %w", err)
	}

	return tokenString, nil
}

// GetUserID получает и валедирует User ID.
func GetUserID(secret string, tokenString string) (string, error) {

	log.Println("!!! Проверяем JWT:", tokenString)
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		})

	if err != nil {
		log.Println("!!!  Ошибка парсинга токена:", err)
		return "", err
	}

	if !token.Valid {
		log.Println("!!! Ошибка: токен недействителен")
		return "", errors.New("invalid token")
	}

	return claims.UserID, nil
}
