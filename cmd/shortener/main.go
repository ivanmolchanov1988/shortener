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
	//srv := server.NewServer(handler)

	// err := srv.ListenAndServe()
	// if err != nil {
	// 	panic(err)
	// }

	r := chi.NewRouter()
	r.Post("/", handler.PostUrl)
	r.Get("/{id}", handler.GetUrl)

	if err := http.ListenAndServe(cfg.Address, nil); err != nil {
		fmt.Printf("start with error: %v\n", err)
	}
}
