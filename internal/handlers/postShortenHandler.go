// Package handlers реализует хендлер для сокращения URL с использованием JSON API.
package handlers

import (
	"encoding/json"
	"net/http"
)

// handleConflict отправляет ответ с HTTP 409 Conflict.
func handleConflict(res http.ResponseWriter, existingShortURL string) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusConflict)
	responseData := struct {
		Result string `json:"result"`
	}{
		Result: existingShortURL,
	}
	if err := json.NewEncoder(res).Encode(responseData); err != nil {
		http.Error(res, "Error encoding response", http.StatusInternalServerError)
	}
}
