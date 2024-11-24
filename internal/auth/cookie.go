package auth

import (
	"net/http"
	"time"

	"github.com/ivanmolchanov1988/shortener/pkg/utils"
)

var TimeForExpire int

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

// JWT из куки
func GetTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

func CreateCookie(w http.ResponseWriter, secret string) (string, error) {
	userID := utils.GenUUID()
	token, err := buildJWTString(secret, userID, time.Hour*time.Duration(TimeForExpire))
	if err != nil {
		return "", err
	}

	setTokenCookie(w, token, time.Hour*time.Duration(TimeForExpire))

	return userID, nil
}
