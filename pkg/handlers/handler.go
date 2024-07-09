package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/ivanmolchanov1988/shortener/config"
	"github.com/ivanmolchanov1988/shortener/pkg/storage"
	"github.com/ivanmolchanov1988/shortener/pkg/utils"
)

type Handler struct {
	storage storage.Storage
	config  *config.Config
}

func NewHandler(s storage.Storage, cfg *config.Config) *Handler {
	return &Handler{
		storage: s,
		config:  cfg,
	}
}

// //////// POST //////////
func (h *Handler) PostUrl(res http.ResponseWriter, req *http.Request) {
	// #1 проверка на POST
	if req.Method != http.MethodPost {
		http.Error(res, "Only the POST method is available", http.StatusBadRequest)
		return
	}
	// #4.1 проверка URL как text/plain
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
	urlStr := string(body)
	_, err = url.ParseRequestURI(urlStr)
	if err != nil {
		http.Error(res, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Забираем рандомную строку для ссылки
	shortUrl, err := utils.RandStr(8)
	if err != nil {
		http.Error(res, "Unable to generate short URL", http.StatusBadRequest)
		return
	}
	// Сохраним URL
	h.storage.SaveURL(shortUrl, urlStr)

	// #2 Header Content-Type = text/plain
	res.Header().Set("Content-Type", "text/plain")
	// #3 res code = 201
	res.WriteHeader(http.StatusCreated)
	// #5 возвращает ответ с сокращённым URL
	//fullShortUrl := fmt.Sprintf("http://%s/%s", req.Host, shortUrl)
	fullShortUrl := fmt.Sprintf("%s/%s", h.config.B_URL, shortUrl)
	res.Write([]byte(fullShortUrl))

	// res.Write([]byte("Тут будет POST '/' URL = text/plain и ответ = 201 + сокращённый URL как text/plain"))
}

// ///////// GET //////////
func (h *Handler) GetUrl(res http.ResponseWriter, req *http.Request) {
	// #6 проверка на GET
	if req.Method != http.MethodGet {
		http.Error(res, "Only the GET method is available", http.StatusBadRequest)
		return
	}

	// #7 парсинг ссылки
	idLink := strings.TrimPrefix(req.URL.Path, "/")
	if idLink == "" {
		http.Error(res, "This ID does not exist", http.StatusBadRequest)
		// res.Header().Set("Content-Type", "text/plain")
		// res.Write([]byte("Shortener"))
		return
	}
	// #8 возвращение исходной ссылки и 307 в HTTP-заголовке Location
	//originUrl, ok := memory.TempUrlStore[idLink]
	originUrl, err := h.storage.GetURL(idLink)
	if err != nil {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", originUrl)
	res.WriteHeader(http.StatusTemporaryRedirect)

	// res.Write([]byte("Тут будет GET '/{id}/' id = идентификатор сокращённого URL и ответ = 307 + оригинальным URL в HTTP-заголовке Location"))
}
