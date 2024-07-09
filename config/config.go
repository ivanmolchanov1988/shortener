package config

import (
	"flag"
	"fmt"
	"os"
)

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

func InitConfig() *Config {

	// Моя Usage
	flag.Usage = Usage

	address := flag.String("a", "localhost:8080", "address to start the HTTP server")
	bURL := flag.String("b", "http://localhost:8080", "the URL for the shortURL")
	flag.Parse()

	if *address == "" || *bURL == "" {
		flag.Usage()
		os.Exit(1)
	}

	return &Config{
		Address: *address,
		B_URL:   *bURL,
	}

}
