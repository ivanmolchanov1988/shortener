package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/ivanmolchanov1988/shortener/internal/compress"
	"github.com/ivanmolchanov1988/shortener/internal/handlers"
	"github.com/ivanmolchanov1988/shortener/internal/memory"

	"github.com/ivanmolchanov1988/shortener/internal/file"
	"github.com/ivanmolchanov1988/shortener/internal/logger"
	"github.com/ivanmolchanov1988/shortener/internal/server"
)

func main() {
	// Конфигурация флагов
	cfg, err := server.InitConfig()
	if err != nil {
		fmt.Printf("InitConfig() with error: %v\n", err)
		os.Exit(1)
	}

	// Инициализация логгера
	if err := logger.Initialize(cfg.Logging); err != nil {
		fmt.Printf("Logger initialization failed: %v\n", err)
		os.Exit(1)
	}

	// Файл для хранения
	filePath := filepath.Clean(cfg.FileStoragePath)

	// Проверяем, есть ли файл
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("File does not exist, creating: %v\n", filePath)
		file, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("Error creating file: %v\n", err)
			os.Exit(1)
		}
		file.Close()
	} else if err != nil {
		fmt.Printf("Error checking file: %v\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("File exists: %v\n", filePath)
	}

	fmt.Printf("File path: %v\n", filePath)
	fileStore := file.NewFileStorage(filePath)
	memStore, err := memory.NewMemoryStorage(fileStore)
	if err != nil {
		fmt.Printf("Error for memStore: %v\n", err)
		os.Exit(1)
	}
	handler := handlers.NewHandler(memStore, cfg)
	//compressHandler := compress.NewCompressHandler(handler) // chi тут решает

	r := chi.NewRouter()

	// Добавляем middleware логирования к каждому запросу
	r.Use(logger.RequestLogger)

	// Применяем middleware сжатия
	r.Use(compress.NewCompressHandler)
	r.Use(compress.DecompressHandler)

	r.Post("/", handler.PostURL)
	r.Post("/api/shorten", handler.Shorten)
	r.Get("/{id}", handler.GetURL)

	fmt.Printf("Server start: => %s\n\r", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, r); err != nil {
		fmt.Printf("Start with error: %v\n", err)
		os.Exit(1)
	}

}
