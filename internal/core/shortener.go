// Package core содержит бизнес-логику сервиса сокращения ссылок.
package core

import (
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/ivanmolchanov1988/shortener/internal/storage"
	"github.com/ivanmolchanov1988/shortener/pkg/utils"
)

type Shortener interface {
	PostURL(userID, originalURL string) (string, error)
	BatchURL(batchItems []BatchRequestItem, userID string) ([]BatchResponseItem, error)
	Shorten(req ShortenRequest, userID string) (ShortenResponse, error)
	GetURL(req GetURLRequest) (GetURLResponse, error)
	GetUserURLs(req GetUserURLsRequest) (GetUserURLsResponse, error)
	DeleteURLs(req DeleteURLsRequest) (DeleteURLsResponse, error)
	GetStats(req GetStatsRequest) (GetStatsResponse, error)
}

type ShortenerService struct {
	storage storage.Storage
	baseURL string
}

func NewShortener(storage storage.Storage, baseURL string) Shortener {
	return &ShortenerService{
		storage: storage,
		baseURL: baseURL,
	}
}

///////// реализация интерфейса /////////

// //////// POST //////////

// PostURL Принимает URL в теле запроса и возвращает сокращенную версию.
func (s *ShortenerService) PostURL(userID, originalURL string) (string, error) {
	shortURL, err := utils.RandStr(8)
	if err != nil {
		return "", err
	}
	id := utils.GenUUID()

	// Сохранение ссылки в хранилище
	existingShortURL, err := s.storage.SaveURL(id, shortURL, originalURL, userID)
	if err != nil {
		if errors.Is(err, ErrURLAlreadyExists) {
			return existingShortURL, ErrURLAlreadyExists
		}
		return "", err
	}

	result := shortURL
	return result, nil
}

// ///////// POST BATCH ////////

// BatchURL обрабатывает запрос POST для создания нового списка сокращенных URLs.
func (s *ShortenerService) BatchURL(batchItems []BatchRequestItem, userID string) ([]BatchResponseItem, error) {
	// Открываем транзакцию для записи
	tx, err := s.storage.BeginTransaction()
	if err != nil {
		//http.Error(res, "Failed to start transaction", http.StatusInternalServerError)
		return nil, ErrorTransaction
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var results []BatchResponseItem

	// Обрабатываем каждый URL
	for _, item := range batchItems {
		responseItem, err := s.processBatchItem(tx, item, userID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		results = append(results, responseItem)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return results, nil
}

// //////// SHORTEN //////////

// Shorten сокращает URL и сохраняет его в хранилище.
func (s *ShortenerService) Shorten(req ShortenRequest, userID string) (ShortenResponse, error) {
	// Проверка валидности URL
	if _, err := url.ParseRequestURI(req.OriginalURL); err != nil {
		return ShortenResponse{}, fmt.Errorf("invalid URL: %w", err)
	}

	// Генерируем короткую ссылку
	shortURL, err := utils.RandStr(8)
	if err != nil {
		return ShortenResponse{}, fmt.Errorf("failed to generate short URL: %w", err)
	}

	// Генерируем UUID
	id := utils.GenUUID()

	// Сохраняем ссылку
	existingShortURL, err := s.storage.SaveURL(id, shortURL, req.OriginalURL, userID)
	if err != nil {
		if errors.Is(err, ErrURLAlreadyExists) {
			return ShortenResponse{ShortURL: existingShortURL}, ErrURLAlreadyExists
		}
		return ShortenResponse{}, fmt.Errorf("error saving URL: %w", err)
	}

	result := ShortenResponse{ShortURL: shortURL}
	return result, nil
}

// ///////// GET //////////

// GetURL ищет оригинальный URL по сокращённому идентификатору.
func (s *ShortenerService) GetURL(req GetURLRequest) (GetURLResponse, error) {
	originalURL, err := s.storage.GetURL(req.ShortURL)
	if err != nil {
		if errors.Is(err, ErrURLNotFound) {
			return GetURLResponse{}, fmt.Errorf("URL not found: %w", err)
		}
		return GetURLResponse{}, fmt.Errorf("failed to get URL: %w", err)
	}

	result := GetURLResponse{OriginalURL: originalURL}
	return result, nil
}

// ///////// GET USER URLS //////////

// GetUserURLs возвращает все ссылки пользователя.
func (s *ShortenerService) GetUserURLs(req GetUserURLsRequest) (GetUserURLsResponse, error) {
	urls, err := s.storage.GetUserURLS(req.UserID)
	if err != nil {
		return GetUserURLsResponse{}, fmt.Errorf("failed to get user URLs: %w", err)
	}

	if len(urls) == 0 {
		return GetUserURLsResponse{}, nil // Возвращаем пустой массив, обработка в хендлере
	}

	var response GetUserURLsResponse
	for _, urlItem := range urls {
		response.URLs = append(response.URLs, UserURL{
			ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, urlItem.ShortURL),
			OriginalURL: urlItem.OriginalURL,
		})
	}

	result := response
	return result, nil
}

// ///////// DELETE USER URLS ///////////

// DeleteURLs удаляет переданные пользователем ссылки.
func (s *ShortenerService) DeleteURLs(req DeleteURLsRequest) (DeleteURLsResponse, error) {
	if len(req.ShortURLs) == 0 {
		return DeleteURLsResponse{}, fmt.Errorf("no URLs provided for deletion")
	}

	// Удаление ссылок в фоновом режиме (как в исходном коде)
	go func() {
		if err := s.storage.DeleteURLS(req.UserID, req.ShortURLs); err != nil {
			fmt.Printf("Failed to delete URLs for user %s: %v\n", req.UserID, err)
		}
	}()

	result := DeleteURLsResponse{Status: "accepted"}
	return result, nil
}

// ///////// GET STATS ///////////

// GetStats возвращает статистику по количеству ссылок и пользователей.
func (s *ShortenerService) GetStats(req GetStatsRequest) (GetStatsResponse, error) {
	// Проверяем, установлена ли доверенная подсеть
	if req.Subnet == "" {
		return GetStatsResponse{}, errors.New("access denied: no trusted subnet configured")
	}

	// Разбираем подсеть
	_, trustedNet, err := net.ParseCIDR(req.Subnet)
	if err != nil {
		return GetStatsResponse{}, fmt.Errorf("invalid trusted subnet configuration: %w", err)
	}

	// Проверяем, входит ли клиентский IP в доверенную подсеть
	clientIP := net.ParseIP(req.ClientIP)
	if clientIP == nil || !trustedNet.Contains(clientIP) {
		return GetStatsResponse{}, errors.New("access denied: unauthorized subnet")
	}

	// Получаем статистику из хранилища
	stats, err := s.storage.GetStats()
	if err != nil {
		return GetStatsResponse{}, fmt.Errorf("failed to retrieve statistics: %w", err)
	}

	return GetStatsResponse{
		URLs:  stats.URLs,
		Users: stats.Users,
	}, nil
}
