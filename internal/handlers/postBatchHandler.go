// Package handlers реализует обработку пакетных (batch) HTTP-запросов для массового сокращения ссылок.
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ivanmolchanov1988/shortener/internal/storage"
	"github.com/ivanmolchanov1988/shortener/pkg/utils"
)

// decodeRequestBody декодирует JSON запроса.
func decodeRequestBody[T any](req *http.Request) (T, error) {
	var data T
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&data)
	return data, err
}

// writeJSONResponse отправляет JSON ответ.
func writeJSONResponse(res http.ResponseWriter, status int, data interface{}) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	if err := json.NewEncoder(res).Encode(data); err != nil {
		http.Error(res, "Error encoding response", http.StatusInternalServerError)
	}
}

// обработка Batch
func (h *Handler) processBatch(requestData []struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}, tx storage.TransactionStorage, res http.ResponseWriter, req *http.Request) ([]struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}, error) {
	var responseData []struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	for _, item := range requestData {
		// Валидируем URL
		if _, err := url.ParseRequestURI(item.OriginalURL); err != nil {
			return nil, fmt.Errorf("invalid URL: %s", item.OriginalURL)
		}

		// Генерируем уникальную короткую ссылку
		shortURL, err := utils.RandStr(8)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short URL")
		}

		// Получаем userID из куки
		userID, err := getUserIDFromCookie(res, req, h.config.Secret, true)
		if err != nil {
			return nil, fmt.Errorf("error fetching user ID")
		}

		// Сохраняем URL
		id := utils.GenUUID()
		if _, err := tx.SaveURLTx(id, shortURL, item.OriginalURL, userID); err != nil {
			return nil, fmt.Errorf("error saving URL: %v", err)
		}

		// Формируем ответ
		responseData = append(responseData, struct {
			CorrelationID string `json:"correlation_id"`
			ShortURL      string `json:"short_url"`
		}{
			CorrelationID: item.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", h.config.BaseURL, shortURL),
		})
	}

	return responseData, nil
}
