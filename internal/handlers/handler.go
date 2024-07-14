package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/ivanmolchanov1988/shortener/internal/server"
	"github.com/ivanmolchanov1988/shortener/pkg/utils"
)

// interfases
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
func (h *Handler) PostUrl(res http.ResponseWriter, req *http.Request) {
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
	fullShortUrl := fmt.Sprintf("%s/%s", h.config.BaseURL, shortURL)
	res.Write([]byte(fullShortUrl))

}

// ///////// GET //////////
func (h *Handler) GetUrl(res http.ResponseWriter, req *http.Request) {
	// #7 парсинг ссылки
	idLink := strings.TrimPrefix(req.URL.Path, "/")
	if idLink == "" {
		http.Error(res, "This ID does not exist", http.StatusBadRequest)
		return
	}
	// #8 возвращение исходной ссылки и 307 в HTTP-заголовке Location
	originUrl, err := h.storage.GetURL(idLink)
	if err != nil {
		http.Error(res, "URL not found", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", originUrl)
	res.WriteHeader(http.StatusTemporaryRedirect)
}