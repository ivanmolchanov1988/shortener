package config

import (
	"flag"
)

type Config struct {
	Address string
	B_URL   string
}

func InitConfig() *Config {
	config := &Config{}

	flag.StringVar(&config.Address, "a", "localhost:8080", "address to start the HTTP server")
	flag.StringVar(&config.B_URL, "b", "http://localhost:8080", "the URL for the shortURL")
	flag.Parse()

	return config
}
