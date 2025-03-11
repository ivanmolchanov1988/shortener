// Package handlers содержит вспомогательные функции для обработки HTTP-запросов и ответов.
package handlers

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/ivanmolchanov1988/shortener/internal/auth"
)

////////// 4 ALL //////////

// generateShortURL генерирует короткую ссылку.
func generateShortURL(baseURL string, shortID string) string {
	return fmt.Sprintf("%s/%s", baseURL, shortID)
}

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

// GetPingDB -  DB ping
func (h *Handler) GetPingDB(res http.ResponseWriter, req *http.Request) {
	dbDSN := h.config.DatabaseDsn
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		http.Error(res, "Failed to connect to database", http.StatusInternalServerError)
	}
	defer db.Close()

	res.WriteHeader(http.StatusOK)

}

// writeJSONResponse отправляет JSON ответ.
func writeJSONResponse(res http.ResponseWriter, status int, data interface{}) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	if err := json.NewEncoder(res).Encode(data); err != nil {
		http.Error(res, "Error encoding response", http.StatusInternalServerError)
	}
}

// extractIDFromPath удаляет префикс "/" из URL и возвращает ID.
func extractIDFromPath(path string) string {
	return strings.TrimPrefix(path, "/")
}

// writeRedirectResponse отправляет HTTP-ответ с редиректом.
func writeRedirectResponse(res http.ResponseWriter, location string, statusCode int) {
	res.Header().Set("Location", location)
	res.WriteHeader(statusCode)
}

// readRequestBody читает тело запроса с поддержкой GZIP.
func readRequestBody(req *http.Request) ([]byte, error) {
	// var body []byte
	// var err error

	if req.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(req.Body)
		if err != nil {
			return nil, err
		}
		defer gz.Close()

		body, err := io.ReadAll(gz)
		if err != nil {
			return nil, err
		}
		return body, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	return body, err
}

// handleConflictResponse отправляет короткий URL.
func handleConflictResponse(w http.ResponseWriter, baseURL, existingShortURL string) {
	if baseURL == "" || existingShortURL == "" {
		http.Error(w, "Invalid base URL or short URL", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusConflict)
	fullShortURL := generateShortURL(baseURL, existingShortURL)
	_, err := w.Write([]byte(fullShortURL))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// resolveClientIP получает IP клиента из X-Real-IP, X-Forwarded-For или RemoteAddr.
func resolveClientIP(r *http.Request) (net.IP, error) {
	// 1. Проверяем заголовок X-Real-IP
	ipStr := r.Header.Get("X-Real-IP")
	ip := net.ParseIP(ipStr)
	if ip != nil {
		return ip, nil
	}

	// 2. Проверяем заголовок X-Forwarded-For (цепочка IP через запятую)
	ips := r.Header.Get("X-Forwarded-For")
	if ips != "" {
		ipList := net.ParseIP(ips)
		if ipList != nil {
			return ipList, nil
		}
	}

	// 3. Если заголовки отсутствуют, берём IP из RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, err
	}
	return net.ParseIP(host), nil
}
