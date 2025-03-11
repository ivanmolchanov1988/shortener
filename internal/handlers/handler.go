// Package handlers определяет основной HTTP-хендлер, управляющий запросами к сервису сокращения ссылок.
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ivanmolchanov1988/shortener/internal/core"
	"github.com/ivanmolchanov1988/shortener/internal/server"
)

// Handler отвечает за обработку HTTP-запросов.
type Handler struct {
	//storage storage.Storage
	shortener core.Shortener
	config    *server.Config
}

// NewHandler -  Создание нового Handler.
func NewHandler(s core.Shortener, cfg *server.Config) *Handler {
	if cfg == nil {
		panic("can't be nil for cfg")
	}
	if s == nil {
		panic("can't be nil for shortener")
	}

	return &Handler{
		shortener: s,
		config:    cfg,
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

	userID, err := getUserIDFromCookie(res, req, h.config.Secret, true)
	if err != nil {
		writeErrorResponse(res, http.StatusInternalServerError, "Error fetching user ID")
		return
	}

	shortURL, err := h.shortener.PostURL(userID, string(body))
	if err != nil {
		if errors.Is(err, core.ErrURLAlreadyExists) {
			handleConflictResponse(res, h.config.BaseURL, shortURL)
			return
		}
		writeErrorResponse(res, http.StatusInternalServerError, "Error saving URL")
		return
	}

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

	var requestData []core.BatchRequestItem
	if err := json.NewDecoder(req.Body).Decode(&requestData); err != nil {
		http.Error(res, "Error decoding request body", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromCookie(res, req, h.config.Secret, true)
	if err != nil {
		http.Error(res, "Unauthorized", http.StatusUnauthorized)
		return
	}

	results, err := h.shortener.BatchURL(requestData, userID)
	if err != nil {
		http.Error(res, "Error processing batch request", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ
	writeJSONResponse(res, http.StatusCreated, results)
}

// //////// SHORTEN //////////

// Shorten обрабатывает запрос POST для создания нового сокращенного URL.
// Принимает URL как json с параметром url в теле запроса и возвращает сокращенную версию.
func (h *Handler) Shorten(res http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var requestData core.ShortenRequest
	if err := json.NewDecoder(req.Body).Decode(&requestData); err != nil {
		http.Error(res, "Error reading request body", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromCookie(res, req, h.config.Secret, true)
	if err != nil {
		http.Error(res, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response, err := h.shortener.Shorten(requestData, userID)
	if err != nil {
		if errors.Is(err, core.ErrURLAlreadyExists) {
			handleConflictResponse(res, h.config.BaseURL, response.ShortURL)
			return
		}
		http.Error(res, "Error saving URL", http.StatusInternalServerError)
		return
	}

	// Формируем и отправляем успешный ответ
	writeJSONResponse(res, http.StatusCreated, struct {
		Result string `json:"result"`
	}{
		Result: fmt.Sprintf("%s/%s", h.config.BaseURL, response),
	})
}

// ///////// GET //////////

// GetURL обрабатывает запрос GET вида http://{server}/{shortURL} и возвращает полный URL.
func (h *Handler) GetURL(res http.ResponseWriter, req *http.Request) {
	shortURL := extractIDFromPath(req.URL.Path)
	if shortURL == "" {
		writeErrorResponse(res, http.StatusNotFound, "Invalid or empty ID")
		return
	}

	response, err := h.shortener.GetURL(core.GetURLRequest{ShortURL: shortURL})
	if err != nil {
		writeErrorResponse(res, http.StatusNotFound, "URL not found")
		return
	}

	writeRedirectResponse(res, response.OriginalURL, http.StatusTemporaryRedirect)
}

// ///////// GET USER URLS //////////

// GetUserURLS обрабатывает запрос GET вида http://{server}/api/user/urls и возвращает массив с short и original urls.
func (h *Handler) GetUserURLS(w http.ResponseWriter, r *http.Request) {
	var userID string
	userID, err := getUserIDFromCookie(w, r, h.config.Secret, false)
	if err != nil {
		// Создаём новую куку для пользователя
		userID, err = getUserIDFromCookie(w, r, h.config.Secret, true)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to create session")
			return
		}

		// PASS userID
		fmt.Printf("New user created: %s\n", userID)

		// Возвращаем 204 No Content, так как это новый пользователь без ссылок
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Получение URLs пользователя
	response, err := h.shortener.GetUserURLs(core.GetUserURLsRequest{UserID: userID})
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get user URLs")
		return
	}

	// Если URLs нет, отправляем HTTP 204
	if len(response.URLs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	writeJSONResponse(w, http.StatusOK, response.URLs)
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

	var requestData core.DeleteURLsRequest
	if err := json.NewDecoder(req.Body).Decode(&requestData.ShortURLs); err != nil {
		writeErrorResponse(res, http.StatusBadRequest, "Error reading request body for delete")
		return
	}

	requestData.UserID = userID

	_, err = h.shortener.DeleteURLs(requestData)
	if err != nil {
		writeErrorResponse(res, http.StatusInternalServerError, "Failed to delete URLs")
		return
	}

	res.WriteHeader(http.StatusAccepted)
}

// ///////// GET STATS ///////////

// GetStats проверяем X-Real-IP и возвращает статистику по urls и users.
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	// Получаем IP клиента
	clientIP, err := resolveClientIP(r)
	if err != nil {
		http.Error(w, "Failed to resolve client IP", http.StatusInternalServerError)
		return
	}

	response, err := h.shortener.GetStats(core.GetStatsRequest{
		ClientIP: clientIP.String(),
		Subnet:   h.config.TrustedSubnet,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	writeJSONResponse(w, http.StatusOK, response)
}
