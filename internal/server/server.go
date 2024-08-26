package server

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

type Config struct {
	Address string
	BaseURL string
	Logging string
}

func Usage() {
	var version = "0.0.1"

	fmt.Fprintf(flag.CommandLine.Output(), "Use: %s\n\n\r ", os.Args[0])
	fmt.Fprintf(flag.CommandLine.Output(), "Version: %s\n\n ", version)
	flag.PrintDefaults()
}

func getAddressAndBaseURL() (string, string, string) {
	address := os.Getenv("SERVER_ADDRESS")
	baseURL := os.Getenv("BASE_URL")
	logging := os.Getenv("LOG_LVL")

	tempAddress := flag.String("a", "localhost:8080", "address to start the HTTP server")
	tempBaseURL := flag.String("b", "http://localhost:8080", "the URL for the shortURL")
	tempLogging := flag.String("log-level", "info", "logging for INFO lvl")

	flag.Parse()

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
	if logging == "" {
		logging = *tempLogging
	} // добать остальные уровни логирования...
	return address, baseURL, logging
}

func InitConfig() (*Config, error) {
	flag.Usage = Usage

	address, baseURL, logging := getAddressAndBaseURL()

	if address == "" || baseURL == "" {
		flag.Usage()
		return nil, errors.New("the address or baseURL is empty")
	}

	return &Config{
		Address: address,
		BaseURL: baseURL,
		Logging: logging,
	}, nil

}
