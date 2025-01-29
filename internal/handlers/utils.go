package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/ivanmolchanov1988/shortener/internal/auth"
)

////////// 4 ALL //////////

// generateShortURL генерирует короткую ссылку.
func generateShortURL(baseURL string, shortID string) string {
	return fmt.Sprintf("%s/%s", baseURL, shortID)
}

// validateURL проверяет, является ли строка допустимым URL.
// func validateURL(urlStr string) error {
// 	_, err := url.ParseRequestURI(urlStr)
// 	return err
// }

// writeErrorResponse отправляет ошибку.
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	http.Error(w, message, statusCode)
}

// getUserIDFromCookie получает UserID из куков
func getUserIDFromCookie(w http.ResponseWriter, r *http.Request, secret string, createIfMissing bool) (string, error) {
	tokenString, err := auth.GetTokenFromCookie(r)
	if err != nil {
		if createIfMissing {
			return auth.CreateCookie(w, secret)
		}

		return "", err
	}

	userID, err := auth.GetUserID(secret, tokenString)
	if err != nil {
		if createIfMissing {
			return auth.CreateCookie(w, secret)
		}

		return "", err
	}

	return userID, nil
}

// DB ping
func (h *Handler) GetPingDB(res http.ResponseWriter, req *http.Request) {
	dbDSN := h.config.DatabaseDsn
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		http.Error(res, "Failed to connect to database", http.StatusInternalServerError)
	}
	defer db.Close()

	res.WriteHeader(http.StatusOK)

}
