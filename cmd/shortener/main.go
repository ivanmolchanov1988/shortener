package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/ivanmolchanov1988/shortener/internal/compress"
	"github.com/ivanmolchanov1988/shortener/internal/handlers"
	"github.com/ivanmolchanov1988/shortener/internal/memory"

	"github.com/ivanmolchanov1988/shortener/internal/filestore"
	"github.com/ivanmolchanov1988/shortener/internal/logger"
	"github.com/ivanmolchanov1988/shortener/internal/server"

	"database/sql"

	_ "github.com/lib/pq"
)

func main() {
	// Конфигурация флагов и хранилища
	cfg, err := server.InitConfigAndPrepareStorage()
	if err != nil {
		fmt.Printf("InitConfig() with error: %v\n", err)
		os.Exit(1)
	}

	// Инициализация логгера
	if err := logger.Initialize(cfg.Logging); err != nil {
		fmt.Printf("Logger initialization failed: %v\n", err)
		os.Exit(1)
	}

	// Файловое хранилище
	fStore := filestore.NewFileStorage(cfg.FileStoragePath)
	memStore, err := memory.NewStorage(fStore)
	if err != nil {
		fmt.Printf("Error for memStore: %v\n", err)
		os.Exit(1)
	}

	//DATA BASE
	dbDSN := cfg.DatabaseDsn
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	//хэндлеры
	handler := handlers.NewHandler(memStore, cfg)

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

	fmt.Printf("Server start: => %s\n\r", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, r); err != nil {
		fmt.Printf("Start with error: %v\n", err)
		os.Exit(1)
	}

}
