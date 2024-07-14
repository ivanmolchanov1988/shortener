package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ivanmolchanov1988/shortener/internal/handlers"
	"github.com/ivanmolchanov1988/shortener/internal/memory"

	"github.com/ivanmolchanov1988/shortener/internal/server"
)

func main() {
	// Конфигурация флагов
	cfg := server.InitConfig()

	memStore := memory.NewMemoryStorage()
	handler := handlers.NewHandler(memStore, cfg)

	r := chi.NewRouter()
	r.Post("/", handler.PostUrl)
	r.Get("/{id}", handler.GetUrl)

	fmt.Printf("Server start: => %s\n\r", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, r); err != nil {
		fmt.Printf("Start with error: %v\n", err)
	}
}
