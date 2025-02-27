package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/ivanmolchanov1988/shortener/internal/handlers"
	"github.com/ivanmolchanov1988/shortener/internal/memory"
	"github.com/ivanmolchanov1988/shortener/internal/server"
)

func ExampleHandler_Shorten() {
	cfg := &server.Config{
		Address: "localhost:8080",
		BaseURL: "http://localhost:8080",
	}
	storage := memory.NewMemoryStorage()
	handler := handlers.NewHandler(storage, cfg)

	requestData := map[string]string{"url": "https://example.com"}
	requestBody, _ := json.Marshal(requestData)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.Shorten(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	var response map[string]string
	json.NewDecoder(res.Body).Decode(&response)

	fmt.Println("Status code:", res.StatusCode)
	if strings.HasPrefix(response["result"], "http://localhost:8080/") {
		fmt.Println("Shorten response valid")
	} else {
		fmt.Println("Shorten response invalid")
	}

	// Output:
	// Status code: 201
	// Shorten response valid
}
