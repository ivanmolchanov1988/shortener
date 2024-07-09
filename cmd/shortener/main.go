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
	// Проверка параметров
	// if len(os.Args) < 2 {
	// 	config.Usage()
	// 	os.Exit(1)
	// }

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

	fmt.Printf("Server start: => %s\n\r", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, r); err != nil {
		fmt.Printf("Start with error: %v\n", err)
	}
}
