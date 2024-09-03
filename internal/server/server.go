package server

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Абсолютный путь к каталогу data в корне проекта и файлу urls.json
// var DefaultFilePath = getDefaultFilePath()

// Имя файла-хранилища
//var dataFileName = "urls.json"

type Config struct {
	Address         string
	BaseURL         string
	Logging         string
	FileStoragePath string
}

func Usage() {
	var version = "0.0.1"

	fmt.Fprintf(flag.CommandLine.Output(), "Use: %s\n\n\r ", os.Args[0])
	fmt.Fprintf(flag.CommandLine.Output(), "Version: %s\n\n ", version)
	flag.PrintDefaults()
}

func getFlags() (string, string, string, string) {
	tempAddress := flag.String("a", "localhost:8080", "address to start the HTTP server")
	tempBaseURL := flag.String("b", "http://localhost:8080", "the URL for the shortURL")
	tempLogging := flag.String("log-level", "info", "logging for INFO lvl")
	tempFilePath := flag.String("f", getDefaultFilePath(), "file for urls data")

	flag.Parse()

	address := os.Getenv("SERVER_ADDRESS")
	baseURL := os.Getenv("BASE_URL")
	logging := os.Getenv("LOG_LVL")
	filePath := os.Getenv("FILE_STORAGE_PATH")
	//filePath = filePath + dataFileName

	if address == "" {
		address = *tempAddress
	} else {
		fmt.Printf("Using ENV for address: %s\n", address)
	}
	if baseURL == "" {
		baseURL = *tempBaseURL
	} else {
		fmt.Printf("Using ENV for baseURL: %s\n", baseURL)
	}
	if filePath == "" {
		filePath = *tempFilePath
	} else {
		fmt.Printf("Using ENV for file path: %s\n", filePath)
	}
	if logging == "" {
		logging = *tempLogging
	} // добать остальные уровни логирования...
	return address, baseURL, filePath, logging
}

func InitConfig() (*Config, error) {
	flag.Usage = Usage

	address, baseURL, filePath, logging := getFlags()

	if address == "" || baseURL == "" {
		flag.Usage()
		return nil, errors.New("the address or baseURL is empty")
	}

	return &Config{
		Address:         address,
		BaseURL:         baseURL,
		Logging:         logging,
		FileStoragePath: filePath,
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
	newPath := strings.Replace(projectRoot, "cmd/shortener", "", 1)
	return newPath
}
