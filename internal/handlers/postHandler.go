package handlers

import (
	"compress/gzip"

	"io"
	"net/http"
)

// readRequestBody читает тело запроса с поддержкой GZIP.
func readRequestBody(req *http.Request) ([]byte, error) {
	var body []byte
	var err error

	if req.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(req.Body)
		if err != nil {
			return nil, err
		}
		defer gz.Close()
		body, err = io.ReadAll(gz)
	} else {
		body, err = io.ReadAll(req.Body)
	}

	return body, err
}

// handleConflictResponse отправляет короткий URL.
func handleConflictResponse(w http.ResponseWriter, baseURL, existingShortURL string) {
	if baseURL == "" || existingShortURL == "" {
		http.Error(w, "Invalid base URL or short URL", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusConflict)
	fullShortURL := generateShortURL(baseURL, existingShortURL)
	_, err := w.Write([]byte(fullShortURL))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
