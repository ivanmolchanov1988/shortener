package server

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ivanmolchanov1988/shortener/internal/filestore"
	"github.com/ivanmolchanov1988/shortener/internal/memory"
	postgr "github.com/ivanmolchanov1988/shortener/internal/postgres"
	"github.com/ivanmolchanov1988/shortener/internal/storage"

	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Config struct {
	Address         string
	BaseURL         string
	Logging         string
	FileStoragePath string
	//db
	DatabaseDsn string
}

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

func Usage() {
	var version = "0.0.1"

	fmt.Fprintf(flag.CommandLine.Output(), "Use: %s\n\n\r ", os.Args[0])
	fmt.Fprintf(flag.CommandLine.Output(), "Version: %s\n\n ", version)
	flag.PrintDefaults()
}

func getFlags() FlagsConfig {
	tempAddress := flag.String("a", "localhost:8080", "address to start the HTTP server")
	tempBaseURL := flag.String("b", "http://localhost:8080", "the URL for the shortURL")
	tempLogging := flag.String("log-level", "info", "logging for INFO lvl")
	tempFilePath := flag.String("f", getDefaultFilePath(), "file for urls data")
	//db
	dsn4flag := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		baseDSN.host, baseDSN.port, baseDSN.user, baseDSN.password, baseDSN.dbname, baseDSN.sslmode)
	tempDB := flag.String("d", dsn4flag, "PostgreSQL DSN")
	//OR
	//tempDB := flag.String("d", "", "PostgreSQL DSN (Data Source Name)")

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

func InitConfigAndPrepareStorage() (*Config, storage.Storage, error) {
	cfg, err := InitConfig()
	if err != nil {
		log.Println("Config is failed: %v\n", err)
		return nil, nil, fmt.Errorf("config initialization failed: %w", err)
	}
	if cfg == nil {
		log.Println("Config is nil")
	}

	var store storage.Storage

	// Определение типа хранилища
	switch {
	case cfg.DatabaseDsn != "":
		db, err := initializeDatabase(cfg.DatabaseDsn)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to initialize database: %w", err)
		}
		store, err = postgr.NewPostgresStorage(db)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create NewPostgresStorage: %v", err)
		}
		log.Println("Storage initialized with PostgreSQL")
	case cfg.FileStoragePath != "":
		store, err = filestore.NewFileStorage(cfg.FileStoragePath)
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

func creteDBconnection(dbDSN string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database: %v", err)
	}
	// Проверка подключения
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to connect to database: %v", err)
	}
	return db, nil

}

// func initializeDatabase(db *sql.DB) error {
func initializeDatabase(dbDSN string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbDSN)
	if db == nil {
		return nil, errors.New("DB connection is nil")
	}

	// миграции
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("Failed to create migrate: %v", err)
	}
	rootPath := getShortRoot()
	files, err := os.ReadDir(rootPath + "/internal/migrations")
	if err != nil {
		return nil, fmt.Errorf("Failed to read migrations directory: %v", err)
	}
	for _, file := range files {
		log.Printf("Found migration file: %s", file.Name())
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+rootPath+"/internal/migrations",
		baseDSN.dbname,
		driver)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize migrate: %v", err)
	}
	//запуск миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("Failed to apply migrate: %v", err)
	}
	log.Println("Migration is complete!")

	return db, nil
}

func initializeFileStorage(filePath string) (*filestore.FileStorage, error) {
	if err := CreateDirectories(filePath); err != nil {
		return nil, err
	}

	if err := CreateFileIfNotExist(filePath); err != nil {
		return nil, err
	}

	store, err := filestore.NewFileStorage(filePath)
	if err != nil {
		return nil, err
	}

	return store, nil
}

func InitConfig() (*Config, error) {
	flag.Usage = Usage

	flags := getFlags()

	if flags.Address == "" || flags.BaseURL == "" {
		flag.Usage()
		return nil, errors.New("the address or baseURL is empty")
	}

	// Логирование для отладки (можете убрать после проверки)
	log.Printf("Flags:\nAddress: %s\nBaseURL: %s\nFilePath: %s\nLogging: %s\nDatabaseDsn: %s\n",
		flags.Address, flags.BaseURL, flags.FilePath, flags.Logging, flags.DatabaseDsn)
	///

	return &Config{
		Address:         flags.Address,
		BaseURL:         flags.BaseURL,
		Logging:         flags.Logging,
		FileStoragePath: flags.FilePath,
		//db
		DatabaseDsn: flags.DatabaseDsn,
	}, nil

}

func CreateDirectories(filePath string) error {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Directory does not exist, creating: %v\n", dir)
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
	}
	return nil
}

// Проверяем наличие файла и создаем его, если он отсутствует
func CreateFileIfNotExist(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("File does not exist, creating: %v\n", filePath)
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}
		file.Close()
	} else if err != nil {
		return fmt.Errorf("error checking file: %w", err)
	} else {
		//log.Printf("File exists: %v\n", filePath)
	}
	return nil
}

/// help funcs - вынести из server

func getProjectRoot() string {
	// Используем текущий рабочий каталог как корневой каталог
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return ""
	}
	return dir
}
func getDefaultFilePath() string {
	projectRoot := getProjectRoot()
	newPath := filepath.Join(projectRoot, "urls.json")
	return newPath
}
func getShortRoot() string {
	var fullRoot = getProjectRoot()
	index := strings.Index(fullRoot, "/cmd/shortener")
	if index == -1 {
		fmt.Println("Failed to find ShortRoot")
	}
	return fullRoot[:index]
}
