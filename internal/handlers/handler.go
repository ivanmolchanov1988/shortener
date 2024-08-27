package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/ivanmolchanov1988/shortener/internal/server"
	"github.com/ivanmolchanov1988/shortener/pkg/utils"
)

// interfaces
type Storage interface {
	SaveURL(shortURL, originalURL string) error
	GetURL(shortURL string) (string, error)
}

//

type Handler struct {
	storage Storage
	config  *server.Config
}

func NewHandler(s Storage, cfg *server.Config) *Handler {
	return &Handler{
		storage: s,
		config:  cfg,
	}
}

// //////// POST //////////
func (h *Handler) PostURL(res http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		http.Error(res, "Content-Type must be text/plain", http.StatusBadRequest)
		return
	}
	// #4.2 Сервер принимает в теле запроса строку URL
	body, err := io.ReadAll(req.Body)
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
	h.storage.SaveURL(shortURL, urlStr)

	// #2 Header Content-Type = text/plain
	res.Header().Set("Content-Type", "text/plain")
	// #3 res code = 201
	res.WriteHeader(http.StatusCreated)
	// #5 возвращает ответ с сокращённым URL
	fullShortURL := fmt.Sprintf("%s/%s", h.config.BaseURL, shortURL)
	res.Write([]byte(fullShortURL))

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
	err = h.storage.SaveURL(shortURL, requestData.URL)
	if err != nil {
		http.Error(res, "Error saving URL", http.StatusInternalServerError)
		return
	}

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
