// Package auth реализует работу с куками, используемыми для хранения аутентификационных данных.
package auth

import (
	"log"
	"net/http"
	"time"

	"github.com/ivanmolchanov1988/shortener/pkg/utils"
)

// TimeForExpire - время протухания токена.
var TimeForExpire int

// GetTimeForExpire пересохраняет время протухания в глобальную переменную.
func GetTimeForExpire(timeForExpire int) {
	TimeForExpire = timeForExpire
}

// JWT в куку
func setTokenCookie(w http.ResponseWriter, token string, duration time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(duration),
	})
}

// GetTokenFromCookie забирает токен из куков HTTP-запроса.
func GetTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		log.Println("!!! Ошибка: не найден session_token в куках")
		return "", err
	}

	return cookie.Value, nil
}

// CreateCookie пишет токен в куки.
func CreateCookie(w http.ResponseWriter, secret string) (string, error) {
	userID := utils.GenUUID()
	token, err := buildJWTString(secret, userID, time.Hour*time.Duration(TimeForExpire))
	if err != nil {
		return "", err
	}

	setTokenCookie(w, token, time.Hour*time.Duration(TimeForExpire))

	return userID, nil
}
