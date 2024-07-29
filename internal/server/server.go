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
}

func Usage() {
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
	return address, baseURL
}

func InitConfig() (*Config, error) {
	flag.Usage = Usage

	address, baseURL := getAddressAndBaseURL()

	if address == "" || baseURL == "" {
		flag.Usage()
		return nil, errors.New("the address or baseURL is empty")
	}

	return &Config{
		Address: address,
		BaseURL: baseURL,
	}, nil

}
