package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/ivanmolchanov1988/shortener/internal/handlers"
	"github.com/ivanmolchanov1988/shortener/internal/memory"

	"github.com/ivanmolchanov1988/shortener/internal/server"
)

func main() {
	// Конфигурация флагов
	cfg, err := server.InitConfig()
	if err != nil {
		memStore := memory.NewMemoryStorage()
		handler := handlers.NewHandler(memStore, cfg)

		r := chi.NewRouter()
		r.Post("/", handler.PostURL)
		r.Get("/{id}", handler.GetURL)

		fmt.Printf("Server start: => %s\n\r", cfg.Address)
		if err := http.ListenAndServe(cfg.Address, r); err != nil {
			fmt.Printf("Start with error: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("InitConfig() with error: %v\n", err)
		os.Exit(1)
	}

}
