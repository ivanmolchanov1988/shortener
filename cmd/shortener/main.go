package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/ivanmolchanov1988/shortener/internal/compress"
	"github.com/ivanmolchanov1988/shortener/internal/handlers"
	"github.com/ivanmolchanov1988/shortener/internal/storage"

	"github.com/ivanmolchanov1988/shortener/internal/logger"
	"github.com/ivanmolchanov1988/shortener/internal/server"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	cfg, store, err := server.InitConfigAndPrepareStorage()
	if err != nil {
		log.Fatalf("INIT is failed: %v\n", err)
	}
	// if store == nil {
	// 	log.Fatalf("CFG is failed: %v\n", store)
	// }
	// if cfg == nil {
	// 	log.Fatalf("CFG is failed: %v\n", cfg)
	// }

	//defer store.Close()

	// Логгер
	if err := logger.Initialize(cfg.Logging); err != nil {
		log.Fatalf("Logger initialization failed: %v\n", err)
	}

	// Хендлеры
	r := setupHandlers(store, cfg)

	// Старт
	fmt.Printf("Server start: => %s\n\r", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, r); err != nil {
		fmt.Printf("Start with error: %v\n", err)
		os.Exit(1)
	}

}

func setupHandlers(store storage.Storage, cfg *server.Config) http.Handler {
	//хэндлеры
	handler := handlers.NewHandler(store, cfg)
	r := chi.NewRouter()
	// Добавляем middleware логирования к каждому запросу
	r.Use(logger.RequestLogger)
	// Применяем middleware сжатия
	r.Use(compress.NewCompressHandler)
	r.Use(compress.DecompressHandler)
	r.Post("/", handler.PostURL)
	r.Post("/api/shorten", handler.Shorten)
	r.Get("/{id}", handler.GetURL)
	r.Get("/ping", handler.GetPingDB)

	return r
}
