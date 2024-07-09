package config

import (
	"flag"
	"fmt"
	"os"
)

// Добавил export SERVER_ADDRESS=localhost:8080
// Добавил export BASE_URL=http://localhost:8080
type Config struct {
	Address string
	B_URL   string
}

func Usage() {
	// Версия
	var version = "0.0.1"

	fmt.Fprintf(flag.CommandLine.Output(), "Use: %s\n\n\r ", os.Args[0])
	fmt.Fprintf(flag.CommandLine.Output(), "Version: %s\n\n ", version)
	flag.PrintDefaults()
}

func getAddressAndBaseURL() (string, string) {
	address := os.Getenv("SERVER_ADDRESS")
	baseURL := os.Getenv("BASE_URL")

	tempAddress := flag.String("a", "localhost:8080", "address to start the HTTP server")
	tempBaseURL := flag.String("b", "http://localhost:8080", "the URL for the shortURL")

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
	return address, baseURL
}

func InitConfig() *Config {

	// Моя Usage
	flag.Usage = Usage

	address, baseURL := getAddressAndBaseURL()
	flag.Parse()

	if address == "" || baseURL == "" {
		flag.Usage()
		os.Exit(1)
	}

	return &Config{
		Address: address,
		B_URL:   baseURL,
	}

}
