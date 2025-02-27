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
	EnableHTTPS  bool
}

// FlagsConfig - флаги
type FlagsConfig struct {
	Address  string
	BaseURL  string
	FilePath string
	Logging  string
	//db
	DatabaseDsn string
	//https
	EnableHTTPS    bool
	ConfigFilePath string
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

func getFlags() (string, FlagsConfig) {
	tempAddress := flag.String("a", "localhost:8080", "address to start the HTTP server")
	tempBaseURL := flag.String("b", "http://localhost:8080", "the URL for the shortURL")
	tempLogging := flag.String("log-level", "info", "logging for INFO lvl")
	tempFilePath := flag.String("f", getDefaultFilePath(), "file for urls data")
	tempDB := flag.String("d", "", "Postgre DSN (Data Source Name)")
	tempEnableHTTPS := flag.Bool("s", false, "enable HTTPS (true/false)")
	tempConfigPath := flag.String("c", "", "path to config file in JSON format")

	flag.Parse()

	// address := os.Getenv("SERVER_ADDRESS")
	// baseURL := os.Getenv("BASE_URL")
	// logging := os.Getenv("LOG_LVL")
	// filePath := os.Getenv("FILE_STORAGE_PATH")
	// dbDSN := os.Getenv("DATABASE_DSN")
	// envHTTPS := os.Getenv("ENABLE_HTTPS")
	// var enableHTTPS bool
	// envConfigPath := os.Getenv("CONFIG")

	// if address == "" {
	// 	address = *tempAddress
	// } else {
	// 	fmt.Printf("Using ENV(SERVER_ADDRESS) for address: %s\n", address)
	// }
	// if baseURL == "" {
	// 	baseURL = *tempBaseURL
	// } else {
	// 	fmt.Printf("Using ENV(BASE_URL) for baseURL: %s\n", baseURL)
	// }
	// if filePath == "" {
	// 	filePath = *tempFilePath
	// } else {
	// 	fmt.Printf("Using ENV(FILE_STORAGE_PATH) for file path: %s\n", filePath)
	// }
	// if dbDSN == "" {
	// 	dbDSN = *tempDB
	// } else {
	// 	fmt.Printf("Using ENV(DATABASE_DSN) for addressDB: %s\n", dbDSN)
	// }
	// if logging == "" {
	// 	logging = *tempLogging
	// } // добать остальные уровни логирования...
	// if envHTTPS == "" {
	// 	enableHTTPS = *tempEnableHTTPS
	// } else {
	// 	enableHTTPS = envHTTPS == "true"
	// 	fmt.Printf("Using ENV(ENABLE_HTTPS) for HTTPS: %t\n", enableHTTPS)
	// }
	// if envConfigPath == "" {

	// }

	// return FlagsConfig{
	// 	Address:     address,
	// 	BaseURL:     baseURL,
	// 	FilePath:    filePath,
	// 	Logging:     logging,
	// 	DatabaseDsn: dbDSN,
	// 	EnableHTTPS: enableHTTPS,
	// 	ConfigFilePath: *tempConfigPath,
	// }

	return *tempConfigPath, FlagsConfig{
		Address:     *tempAddress,
		BaseURL:     *tempBaseURL,
		FilePath:    *tempFilePath,
		Logging:     *tempLogging,
		DatabaseDsn: *tempDB,
		EnableHTTPS: *tempEnableHTTPS,
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
		// log.Println("Config is nil")
		return nil, nil, errors.New("failed to initialize config")
	}

	var store storage.Storage

	// Определение типа хранилища
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
}

// InitConfig подготавливает конфиг.
func InitConfig() (*Config, error) {
	flag.Usage = Usage

	// flags := getFlags()
	configPath, flags := getFlags()

	// если flag не указан, пробуем ENV
	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	// теперь пробуем загрузить из файла
	fileConfig, fileLoaded, err := loadConfig(configPath)
	if err != nil {
		log.Printf("Failed to load config from file: %v\n", err)
		fileConfig = &Config{} // Если ошибка, используем пустой конфиг
	}

	// Flags > ENV > JSON
	address := firstNonEmpty(flags.Address, os.Getenv("SERVER_ADDRESS"), fileConfig.Address)
	baseURL := firstNonEmpty(flags.BaseURL, os.Getenv("BASE_URL"), fileConfig.BaseURL)
	filePath := firstNonEmpty(flags.FilePath, os.Getenv("FILE_STORAGE_PATH"), fileConfig.FileStoragePath)
	logging := firstNonEmpty(flags.Logging, os.Getenv("LOG_LVL"), fileConfig.Logging)
	dbDSN := firstNonEmpty(flags.DatabaseDsn, os.Getenv("DATABASE_DSN"), fileConfig.DatabaseDsn)
	enableHTTPS := firstNonEmptyBool(flags.EnableHTTPS, parseBool(os.Getenv("ENABLE_HTTPS")), fileConfig.EnableHTTPS)

	if flags.Address == "" || flags.BaseURL == "" {
		Usage()
		return nil, errors.New("the address or baseURL is empty")
	}

	// Логирование для отладки
	log.Printf("Config (JSON loaded: %t):\nAddress: %s\nBaseURL: %s\nFilePath: %s\nLogging: %s\nDatabaseDsn: %s\nEnableHTTPS: %t\n",
		fileLoaded, address, baseURL, filePath, logging, dbDSN, enableHTTPS)
	///

	return &Config{
		Address:         address,
		BaseURL:         baseURL,
		Logging:         logging,
		FileStoragePath: filePath,
		//db
		DatabaseDsn:  dbDSN,
		Secret:       "secret",
		TimeToExpire: 3,
		EnableHTTPS:  enableHTTPS,
	}, nil

}
