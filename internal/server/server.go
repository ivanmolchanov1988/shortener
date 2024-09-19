package server

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
		fmt.Printf("File exists: %v\n", filePath)
	}
	return nil
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
	tempDB := flag.String("d", "host=localhost port=5432 user=postgres password=password dbname=shortener sslmode=disable", "PostgreSQL DSN (Data Source Name)")

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
		fmt.Printf("Using ENV(DATABASE_DSN) for addressDB: %s\n", filePath)
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

func InitConfigAndPrepareStorage() (*Config, error) {
	cfg, err := InitConfig()
	if err != nil {
		return nil, err
	}

	// Создание директорий и файлов
	if err := CreateDirectories(cfg.FileStoragePath); err != nil {
		return nil, err
	}

	if err := CreateFileIfNotExist(cfg.FileStoragePath); err != nil {
		return nil, err
	}

	return cfg, nil
}

func InitConfig() (*Config, error) {
	flag.Usage = Usage

	flags := getFlags()

	if flags.Address == "" || flags.BaseURL == "" {
		flag.Usage()
		return nil, errors.New("the address or baseURL is empty")
	}

	return &Config{
		Address:         flags.Address,
		BaseURL:         flags.BaseURL,
		Logging:         flags.Logging,
		FileStoragePath: flags.FilePath,
		//db
		DatabaseDsn: flags.DatabaseDsn,
	}, nil

}

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
