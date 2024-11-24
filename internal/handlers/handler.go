package handlers

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/ivanmolchanov1988/shortener/internal/auth"
	"github.com/ivanmolchanov1988/shortener/internal/server"
	"github.com/ivanmolchanov1988/shortener/internal/storage"
	"github.com/ivanmolchanov1988/shortener/pkg/utils"
)

type Handler struct {
	storage   storage.Storage
	txStorage storage.TransactionStorage
	config    *server.Config
}

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
func (h *Handler) PostURL(res http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") && !strings.Contains(contentType, "application/x-gzip") {
		http.Error(res, "Content-Type must be text/plain or application/x-gzip", http.StatusBadRequest)
		return
	}

	var body []byte
	var err error

	if req.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(req.Body)
		if err != nil {
			http.Error(res, "Failed to decompress gzip body", http.StatusBadRequest)
			return
		}
		defer gz.Close()
		body, err = io.ReadAll(gz)
		if err != nil {
			http.Error(res, "Unable to read body", http.StatusBadRequest)
			return
		}
	} else {
		body, err = io.ReadAll(req.Body)
	}

	// #4.2 Сервер принимает в теле запроса строку URL
	if err != nil {
		http.Error(res, "Unable to read body", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	urlStr := string(body)
	_, err = url.ParseRequestURI(urlStr)
	if err != nil {
		http.Error(res, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Забираем рандомную строку для ссылки
	shortURL, err := utils.RandStr(8)
	if err != nil {
		http.Error(res, "Unable to generate short URL", http.StatusBadRequest)
		return
	}
	// Сохраним URL
	id := utils.GenUUID()

	userID, err := GetUserIDFromCookie(res, req, h.config.Secret)
	if err != nil {
		http.Error(res, "Failed to get user ID", http.StatusUnauthorized)
		return
	}

	existingShortURL, err := h.storage.SaveURL(id, shortURL, urlStr, userID)
	if err != nil {
		if errors.Is(err, storage.ErrURLAlreadyExists) {
			// Возвращаем HTTP 409 Conflict и существующий shortURL
			res.Header().Set("Content-Type", "text/plain")
			res.WriteHeader(http.StatusConflict)
			fullShortURL := fmt.Sprintf("%s/%s", h.config.BaseURL, existingShortURL)
			res.Write([]byte(fullShortURL))
			return
		}
		log.Printf("Failed to save URL: %v", err)
		http.Error(res, "Error saving URL", http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully saved URL: shortURL=%s, originalURL=%s", shortURL, urlStr)

	// #2 Header Content-Type = text/plain
	res.Header().Set("Content-Type", "text/plain")
	// #3 res code = 201
	res.WriteHeader(http.StatusCreated)
	// #5 возвращает ответ с сокращённым URL
	fullShortURL := fmt.Sprintf("%s/%s", h.config.BaseURL, shortURL)
	res.Write([]byte(fullShortURL))

}

// ///////// POST BATCH ////////
func (h *Handler) Batch(res http.ResponseWriter, req *http.Request) {
	// Проверка Content-Type
	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	// Структура для входящего запроса
	var requestData []struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}
	// ... исходящего
	var responseData []struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	// Декодер JSON
	err := json.NewDecoder(req.Body).Decode(&requestData)
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
	defer tx.Rollback()

	for _, item := range requestData {
		// Валидируем URL
		_, err := url.ParseRequestURI(item.OriginalURL)
		if err != nil {
			http.Error(res, fmt.Sprintf("Invalid URL: %s", item.OriginalURL), http.StatusBadRequest)
			return
		}

		// Генерируем уникальную короткую ссылку
		shortURL, err := utils.RandStr(8)
		if err != nil {
			http.Error(res, "Failed to generate short URL", http.StatusInternalServerError)
			return
		}

		// Сохраняем URL в рамках транзакции
		id := utils.GenUUID()
		//_, err = h.txStorage.SaveURLTx(tx, id, shortURL, item.OriginalURL)
		userID, err := GetUserIDFromCookie(res, req, h.config.Secret)
		if err != nil {
			http.Error(res, "Failed to get user ID", http.StatusUnauthorized)
			return
		}

		_, err = tx.SaveURLTx(id, shortURL, item.OriginalURL, userID)
		if err != nil {
			http.Error(res, "Error saving URL", http.StatusInternalServerError)
			return
		}

		// Заполняем ответ
		responseData = append(responseData, struct {
			CorrelationID string `json:"correlation_id"`
			ShortURL      string `json:"short_url"`
		}{
			CorrelationID: item.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", h.config.BaseURL, shortURL),
		})
	}

	// Завершаем транзакцию
	if err := tx.Commit(); err != nil {
		http.Error(res, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	// Устанавливаем Content-Type и код состояния
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	// Кодируем и отправляем JSON ответ
	err = json.NewEncoder(res).Encode(responseData)
	if err != nil {
		http.Error(res, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// //////// SHORTEN //////////
func (h *Handler) Shorten(res http.ResponseWriter, req *http.Request) {
	// Content-Type - application/json
	if req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	// Структура для входящего запроса
	var requestData struct {
		URL string `json:"url"`
	}

	// Декодируем JSON
	err := json.NewDecoder(req.Body).Decode(&requestData)
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

	// Сохраняем URL
	id := utils.GenUUID()

	userID, err := GetUserIDFromCookie(res, req, h.config.Secret)
	if err != nil {
		http.Error(res, "Failed to get user ID", http.StatusUnauthorized)
		return
	}

	existingShortURL, err := h.storage.SaveURL(id, shortURL, requestData.URL, userID)
	if err != nil {
		if errors.Is(err, storage.ErrURLAlreadyExists) {
			// Возвращаем HTTP 409 Conflict и уже существующий shortURL
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusConflict)
			responseData := struct {
				Result string `json:"result"`
			}{
				Result: fmt.Sprintf("%s/%s", h.config.BaseURL, existingShortURL),
			}
			err = json.NewEncoder(res).Encode(responseData)
			if err != nil {
				http.Error(res, "Error encoding response", http.StatusInternalServerError)
			}
			return
		}
		http.Error(res, "Error saving URL", http.StatusInternalServerError)
		return
	}

	// err = h.storage.SaveURL(id, shortURL, requestData.URL)
	// if err != nil {
	// 	http.Error(res, "Error saving URL", http.StatusInternalServerError)
	// 	return
	// }

	// Структуру ответа
	responseData := struct {
		Result string `json:"result"`
	}{
		Result: fmt.Sprintf("%s/%s", h.config.BaseURL, shortURL),
	}

	// Заголовок Content-Type для ответа
	res.Header().Set("Content-Type", "application/json")
	// 201 Created
	res.WriteHeader(http.StatusCreated)
	// responseData в JSON
	err = json.NewEncoder(res).Encode(responseData)
	if err != nil {
		http.Error(res, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// ///////// GET //////////
func (h *Handler) GetURL(res http.ResponseWriter, req *http.Request) {
	// #7 парсинг ссылки
	idLink := strings.TrimPrefix(req.URL.Path, "/")
	if idLink == "" {
		http.Error(res, "Invalid or empty ID", http.StatusNotFound)
		return
	}
	// #8 возвращение исходной ссылки и 307 в HTTP-заголовке Location
	// 404, если не найден
	originURL, err := h.storage.GetURL(idLink)
	if err != nil {
		http.Error(res, "URL not found", http.StatusNotFound)
		return
	}
	res.Header().Set("Location", originURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
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

// UserID From Cookie
func GetUserIDFromCookie(w http.ResponseWriter, r *http.Request, secret string) (string, error) {
	tokenString, err := auth.GetTokenFromCookie(r)
	if err != nil {
		log.Printf("Error token from cookie: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return "", err
	}

	userID, err := auth.GetUserID(secret, tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return "", err
	}

	return userID, nil
}

// func CreateCookie(w http.ResponseWriter, secret string, h *Handler) (string, error) {
// 	userID := utils.GenUUID()
// 	token, err := auth.BuildJWTString(secret, userID, time.Hour * time.Duration(h.config.TimeToExpire))
// 	if err != nil {
// 		return "", err
// 	}

// 	auth.SetTokenCookie(w, token, time.Hour * time.Duration(h.config.TimeToExpire))

// 	return userID, nil
// }

// GET USER URLS
func (h *Handler) GetUserURLS(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserIDFromCookie(w, r, h.config.Secret)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	urls, err := h.storage.GetUserURLS(userID)
	if err != nil {
		http.Error(w, "Failed to get user URLs", http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(urls); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	//json.NewEncoder(w).Encode(urls)
}
