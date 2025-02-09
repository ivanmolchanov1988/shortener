// Package handlers содержит обработчики HTTP-запросов для работы с сервисом сокращения ссылок.
package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ivanmolchanov1988/shortener/internal/storage"
)

// extractIDFromPath удаляет префикс "/" из URL и возвращает ID.
func extractIDFromPath(path string) string {
	return strings.TrimPrefix(path, "/")
}

// writeRedirectResponse отправляет HTTP-ответ с редиректом.
func writeRedirectResponse(res http.ResponseWriter, location string, statusCode int) {
	res.Header().Set("Location", location)
	res.WriteHeader(statusCode)
}

// Обработка ошибок
func (h *Handler) handleStorageError(res http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrURLIsGone):
		writeErrorResponse(res, http.StatusGone, "URL is gone")
	case errors.Is(err, storage.ErrURLNotFound):
		writeErrorResponse(res, http.StatusNotFound, "URL not found")
	default:
		writeErrorResponse(res, http.StatusInternalServerError, "Internal server error")
	}
}
