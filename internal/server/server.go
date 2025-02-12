// Package server управляет инициализацией конфигурации и запуском HTTP-сервера.
package server

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ivanmolchanov1988/shortener/internal/memory"
	postgr "github.com/ivanmolchanov1988/shortener/internal/postgres"
	"github.com/ivanmolchanov1988/shortener/internal/storage"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// Переменные для stdout
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

// Config - конфиг.
type Config struct {
	Address         string
	BaseURL         string
	Logging         string
	FileStoragePath string
	//db
	DatabaseDsn string
	//user id
	Secret       string
	TimeToExpire int
}

// FlagsConfig - флаги
type FlagsConfig struct {
	Address  string
	BaseURL  string
	FilePath string
	Logging  string
	//db
	DatabaseDsn string
}

var baseDSN = struct {
	host     string
	port     string
	user     string
	password string
	dbname   string
	sslmode  string
}{
	host:     "localhost",
	port:     "5432",
	user:     "postgres",
	password: "password",
	dbname:   "shortener",
	sslmode:  "disable",
}

// Usage - начальное логирование.
func Usage() {
	//fmt.Fprintf(flag.CommandLine.Output(), "Use: %s\n\n\r ", os.Args[0])

	// Для примера
	// go run -ldflags="-X 'github.com/ivanmolchanov1988/shortener/internal/server.buildVersion=1.2.3' -X 'github.com/ivanmolchanov1988/shortener/internal/server.buildDate=2025-02-10' -X 'github.com/ivanmolchanov1988/shortener/internal/server.buildCommit=abcdefg'" main.go
	setDefaultNA(&buildVersion, "N/A")
	setDefaultNA(&buildDate, "N/A")
	setDefaultNA(&buildCommit, "N/A")
	fmt.Fprintf(flag.CommandLine.Output(), "Build version: %s\n\r", buildVersion)
	fmt.Fprintf(flag.CommandLine.Output(), "Build date: %s\n\r", buildDate)
	fmt.Fprintf(flag.CommandLine.Output(), "Build commit: %s\n\n\r", buildCommit)

	flag.PrintDefaults()
}

func getFlags() FlagsConfig {
	tempAddress := flag.String("a", "localhost:8080", "address to start the HTTP server")
	tempBaseURL := flag.String("b", "http://localhost:8080", "the URL for the shortURL")
	tempLogging := flag.String("log-level", "info", "logging for INFO lvl")
	tempFilePath := flag.String("f", getDefaultFilePath(), "file for urls data")
	tempDB := flag.String("d", "", "Postgre DSN (Data Source Name)")

	flag.Parse()

	address := os.Getenv("SERVER_ADDRESS")
	baseURL := os.Getenv("BASE_URL")
	logging := os.Getenv("LOG_LVL")
	filePath := os.Getenv("FILE_STORAGE_PATH")
	dbDSN := os.Getenv("DATABASE_DSN")

	if address == "" {
		address = *tempAddress
	} else {
		fmt.Printf("Using ENV(SERVER_ADDRESS) for address: %s\n", address)
	}
	if baseURL == "" {
		baseURL = *tempBaseURL
	} else {
		fmt.Printf("Using ENV(BASE_URL) for baseURL: %s\n", baseURL)
	}
	if filePath == "" {
		filePath = *tempFilePath
	} else {
		fmt.Printf("Using ENV(FILE_STORAGE_PATH) for file path: %s\n", filePath)
	}
	if dbDSN == "" {
		dbDSN = *tempDB
	} else {
		fmt.Printf("Using ENV(DATABASE_DSN) for addressDB: %s\n", dbDSN)
	}
	if logging == "" {
		logging = *tempLogging
	} // добать остальные уровни логирования...

	return FlagsConfig{
		Address:     address,
		BaseURL:     baseURL,
		FilePath:    filePath,
		Logging:     logging,
		DatabaseDsn: dbDSN,
	}
}

// InitConfigAndPrepareStorage подготавливает конфигурацию и хранилище.
func InitConfigAndPrepareStorage() (*Config, storage.Storage, error) {
	fmt.Println("Initializing configuration and preparing storage...")
	cfg, err := InitConfig()
	if err != nil {
		log.Printf("Config is failed: %v\n", err)
		return nil, nil, fmt.Errorf("config initialization failed: %w", err)
	}
	fmt.Printf("Loaded configuration: %+v\n", cfg)
	if cfg == nil {
		log.Println("Config is nil")
	}

	var store storage.Storage

	// Определение типа хранилища
	if cfg != nil {
		switch {
		case cfg.DatabaseDsn != "":
			db, err := initializeDatabase(cfg.DatabaseDsn)
			if err != nil {
				log.Printf("Database initialization failed, switching to memory storage: %v", err)
				store = memory.NewMemoryStorage()
			} else {
				store, err = postgr.NewPostgresStorage(db)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to create NewPostgresStorage: %v", err)
				}
				log.Println("Storage initialized with Postgre")
			}
		case cfg.FileStoragePath != "":
			store, err = initializeFileStorage(cfg.FileStoragePath)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create file storage: %w", err)
			}
			log.Println("Storage initialized with file storage")
		default:
			store = memory.NewMemoryStorage()
			log.Println("Using mem storage")
		}

		return cfg, store, nil
	} else {
		return nil, nil, fmt.Errorf("failed from config: %w", err)
	}
}

// InitConfig подготавливает конфиг.
func InitConfig() (*Config, error) {
	flag.Usage = Usage

	flags := getFlags()

	if flags.Address == "" || flags.BaseURL == "" {
		flag.Usage()
		return nil, errors.New("the address or baseURL is empty")
	}

	// Логирование для отладки
	log.Printf("Flags:\nAddress: %s\nBaseURL: %s\nFilePath: %s\nLogging: %s\nDatabaseDsn: %s\n",
		flags.Address, flags.BaseURL, flags.FilePath, flags.Logging, flags.DatabaseDsn)
	///

	return &Config{
		Address:         flags.Address,
		BaseURL:         flags.BaseURL,
		Logging:         flags.Logging,
		FileStoragePath: flags.FilePath,
		//db
		DatabaseDsn:  flags.DatabaseDsn,
		Secret:       "secret",
		TimeToExpire: 3,
	}, nil

}
