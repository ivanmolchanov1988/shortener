package core

import (
	"fmt"
	"net/url"

	"github.com/ivanmolchanov1988/shortener/internal/storage"
	"github.com/ivanmolchanov1988/shortener/pkg/utils"
)

// BatchRequestItem представляет элемент запроса при пакетном сокращении ссылок.
type BatchRequestItem struct {
	CorrelationID string
	OriginalURL   string
}

// BatchResponseItem представляет элемент ответа при пакетном сокращении ссылок.
type BatchResponseItem struct {
	CorrelationID string
	ShortURL      string
}

// processBatchItem обрабатывает один элемент batch-запроса
func (s *ShortenerService) processBatchItem(tx storage.TransactionStorage, item BatchRequestItem, userID string) (BatchResponseItem, error) {
	// Проверяем валидность URL
	if _, err := url.ParseRequestURI(item.OriginalURL); err != nil {
		return BatchResponseItem{}, fmt.Errorf("invalid URL: %s", item.OriginalURL)
	}

	// Генерируем короткий URL
	shortURL, err := utils.RandStr(8)
	if err != nil {
		return BatchResponseItem{}, fmt.Errorf("failed to generate short URL: %w", err)
	}

	// Генерируем ID
	id := utils.GenUUID()

	// Сохраняем в транзакции
	existingShortURL, err := tx.SaveURLTx(id, shortURL, item.OriginalURL, userID)
	if err != nil {
		return BatchResponseItem{}, fmt.Errorf("error saving URL: %w", err)
	}

	// Формируем ответ
	return BatchResponseItem{
		CorrelationID: item.CorrelationID,
		ShortURL:      fmt.Sprintf("%s/%s", s.baseURL, existingShortURL),
	}, nil
}
