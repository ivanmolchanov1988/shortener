// Package handlers определяет основной HTTP-хендлер, управляющий запросами к сервису сокращения ссылок.
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/ivanmolchanov1988/shortener/internal/server"
	"github.com/ivanmolchanov1988/shortener/internal/storage"
	"github.com/ivanmolchanov1988/shortener/pkg/utils"
)

// Handler отвечает за обработку HTTP-запросов.
type Handler struct {
	storage storage.Storage
	config  *server.Config
}

// NewHandler -  Создание нового Handler.
func NewHandler(s storage.Storage, cfg *server.Config) *Handler {
	if cfg == nil {
		panic("can't be nil for cfg")
	}
	if s == nil {
		panic("can't be nil for storage")
	}

	return &Handler{
		storage: s,
		config:  cfg,
	}
}

// //////// POST //////////

// PostURL обрабатывает запрос POST для создания нового сокращенного URL.
// Принимает URL в теле запроса и возвращает сокращенную версию.
func (h *Handler) PostURL(res http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") && !strings.Contains(contentType, "application/x-gzip") {
		writeErrorResponse(res, http.StatusBadRequest, "Content-Type must be text/plain or application/x-gzip")
		return
	}

	body, err := readRequestBody(req)
	if err != nil {
		writeErrorResponse(res, http.StatusBadRequest, "Failed to read body")
		return
	}
	defer req.Body.Close()

	urlStr := string(body)
	_, err = url.ParseRequestURI(urlStr)
	if err != nil {
		writeErrorResponse(res, http.StatusBadRequest, "Invalid URL")
		return
	}

	// Забираем рандомную строку для ссылки.
	shortURL, err := utils.RandStr(8)
	if err != nil {
		writeErrorResponse(res, http.StatusInternalServerError, "Failed to generate short URL")
		return
	}
	// Сохраним URL.
	id := utils.GenUUID()

	userID, err := getUserIDFromCookie(res, req, h.config.Secret, true)
	if err != nil {
		writeErrorResponse(res, http.StatusInternalServerError, "Error fetching user ID")
		return
	}

	existingShortURL, err := h.storage.SaveURL(id, shortURL, urlStr, userID)
	if err != nil {
		if errors.Is(err, storage.ErrURLAlreadyExists) {
			handleConflictResponse(res, h.config.BaseURL, existingShortURL)
			return
		}
		writeErrorResponse(res, http.StatusInternalServerError, "Error saving URL")
		return
	}
	log.Printf("Successfully saved URL: shortURL=%s, originalURL=%s", shortURL, urlStr)

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(generateShortURL(h.config.BaseURL, shortURL)))

}

// ///////// POST BATCH ////////

// Batch обрабатывает запрос POST для создания нового списка сокращенных URLs.
// Принимает массив объектов из correlation_id и original_url в теле запроса.
// Возвращает массив объктов из correlation_id и short_url.
func (h *Handler) Batch(res http.ResponseWriter, req *http.Request) {
	// Проверка Content-Type
	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	// Читаем тело запроса и декодируем JSON.
	requestData, err := decodeRequestBody[[]struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}](req)
	if err != nil {
		http.Error(res, "Error decoding request body", http.StatusBadRequest)
		return
	}

	// Открываем транзакцию для записи
	tx, err := h.storage.BeginTransaction()
	if err != nil {
		http.Error(res, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Обрабатываем каждый URL
	responseData, err := h.processBatch(requestData, tx, res, req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// Завершаем транзакцию
	if err := tx.Commit(); err != nil {
		http.Error(res, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ
	writeJSONResponse(res, http.StatusCreated, responseData)
}

// //////// SHORTEN //////////

// Shorten обрабатывает запрос POST для создания нового сокращенного URL.
// Принимает URL как json с параметром url в теле запроса и возвращает сокращенную версию.
func (h *Handler) Shorten(res http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	// Декодируем тело запроса
	requestData, err := decodeRequestBody[struct {
		URL string `json:"url"`
	}](req)
	if err != nil {
		http.Error(res, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Валидность URL
	_, err = url.ParseRequestURI(requestData.URL)
	if err != nil {
		http.Error(res, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Генерируем shortLink
	shortURL, err := utils.RandStr(8)
	if err != nil {
		http.Error(res, "Failed to generate short URL", http.StatusInternalServerError)
		return
	}

	// Получаем userID из куки
	userID, err := getUserIDFromCookie(res, req, h.config.Secret, true)
	if err != nil {
		http.Error(res, "Error fetching user ID", http.StatusInternalServerError)
		return
	}

	// Сохраняем URL
	existingShortURL, err := h.storage.SaveURL(utils.GenUUID(), shortURL, requestData.URL, userID)
	if err != nil {
		if errors.Is(err, storage.ErrURLAlreadyExists) {
			// Возвращаем HTTP 409 Conflict и уже существующий shortURL
			handleConflict(res, fmt.Sprintf("%s/%s", h.config.BaseURL, existingShortURL))
			return
		}
		http.Error(res, "Error saving URL", http.StatusInternalServerError)
		return
	}

	// Формируем и отправляем успешный ответ
	writeJSONResponse(res, http.StatusCreated, struct {
		Result string `json:"result"`
	}{
		Result: fmt.Sprintf("%s/%s", h.config.BaseURL, shortURL),
	})
}

// ///////// GET //////////

// GetURL обрабатывает запрос GET вида http://{server}/{shortURL} и возвращает полный URL.
func (h *Handler) GetURL(res http.ResponseWriter, req *http.Request) {
	idLink := extractIDFromPath(req.URL.Path)
	if idLink == "" {
		writeErrorResponse(res, http.StatusNotFound, "Invalid or empty ID")
		return
	}

	// Получаем оригинальный URL
	originURL, err := h.storage.GetURL(idLink)
	if err != nil {
		h.handleStorageError(res, err)
		return
	}

	writeRedirectResponse(res, originURL, http.StatusTemporaryRedirect)
}

// ///////// GET USER URLS //////////

// GetUserURLS обрабатывает запрос GET вида http://{server}/api/user/urls и возвращает массив с short и original urls.
func (h *Handler) GetUserURLS(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromCookie(w, r, h.config.Secret, false)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Получение URLs пользователя
	urls, err := h.storage.GetUserURLS(userID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user URLs")
		return
	}

	// Если URLs нет, отправляем HTTP 204
	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	for i := range urls {
		urls[i].ShortURL = fmt.Sprintf("%s/%s", h.config.BaseURL, urls[i].ShortURL)
	}

	writeJSONResponse(w, http.StatusOK, urls)
}

// ///////// DELETE USER URLS ///////////

// DeleteURLS обрабатывает запрос DELETE вида http://{server}/api/user/urls.
// Принимает в теле массив из строк с shortURLs и удаляет их.
func (h *Handler) DeleteURLS(res http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Content-Type") != "application/json" {
		writeErrorResponse(res, http.StatusBadRequest, "Content-Type must be application/json")
		return
	}

	userID, err := getUserIDFromCookie(res, req, h.config.Secret, false)
	if err != nil {
		writeErrorResponse(res, http.StatusUnauthorized, "Unauthorized for delete")
		return
	}

	var shortURLs4Delete []string
	err = json.NewDecoder(req.Body).Decode(&shortURLs4Delete)
	if err != nil {
		writeErrorResponse(res, http.StatusBadRequest, "Error reading request body for delete")
		return
	}

	if len(shortURLs4Delete) == 0 {
		writeErrorResponse(res, http.StatusBadRequest, "There are no URLs to delete")
		return
	}

	go func() {
		if err := h.storage.DeleteURLS(userID, shortURLs4Delete); err != nil {
			log.Printf("Failed to delete URLs for user %s: %v", userID, err)
		}
	}()
	// --- Стоит добавить таймаут для асинхронных операций --- НЕ ПРОХОДИТ ТЕСТЫ iter15 !!!
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	// go func(ctx context.Context) {
	// 	select {
	// 	case <-ctx.Done():
	// 		log.Printf("Timeout reached for deleting URLs for user %s", userID)
	// 		return
	// 	default:
	// 		if err := h.storage.DeleteURLS(userID, shortURLs4Delete); err != nil {
	// 			log.Printf("Failed to delete URLs for user %s: %v", userID, err)
	// 		}
	// 	}
	// }(ctx)

	res.WriteHeader(http.StatusAccepted)
}
