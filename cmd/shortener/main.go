package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	//"github.com/go-chi/chi/v5/middleware"
	"github.com/ivanmolchanov1988/shortener/pkg/handlers"
	"github.com/ivanmolchanov1988/shortener/pkg/memory"

	//"github.com/ivanmolchanov1988/shortener/pkg/server"
	"github.com/ivanmolchanov1988/shortener/config"
)

func main() {
	// Конфигурация флагов
	cfg := config.InitConfig()

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
