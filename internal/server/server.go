package server

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

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
	tempFilePath := flag.String("f", "../../data/urls.json", "file for urls data")

	flag.Parse()

	address := os.Getenv("SERVER_ADDRESS")
	baseURL := os.Getenv("BASE_URL")
	logging := os.Getenv("LOG_LVL")
	filePath := os.Getenv("FILE_STORAGE_PATH")

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

	//проверка файла
	if err := checkAndCreateFile(filePath); err != nil {
		return nil, fmt.Errorf("error for fileData: %w", err)
	}

	return &Config{
		Address:         address,
		BaseURL:         baseURL,
		Logging:         logging,
		FileStoragePath: filePath,
	}, nil

}

func checkAndCreateFile(filePath string) error {
	// Проверка существования файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("Creating new file: %s\n", filePath)
		return createNewFile(filePath)
	} else if err != nil {
		return err
	}

	return nil
}

func createNewFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Println("File created:", filePath)
	return nil
}
