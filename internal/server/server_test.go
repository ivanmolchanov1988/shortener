package server

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Тест парсинга флагов
func TestGetFlags(t *testing.T) {
	os.Setenv("SERVER_ADDRESS", "127.0.0.1:8081")
	os.Setenv("BASE_URL", "http://127.0.0.1:8081")
	os.Setenv("LOG_LVL", "debug")
	os.Setenv("FILE_STORAGE_PATH", "/tmp/shortener_test.json")
	os.Setenv("DATABASE_DSN", "postgres://user:password@localhost:5432/dbname")

	flags := getFlagsForTesting()

	assert.Equal(t, "127.0.0.1:8081", flags.Address)
	assert.Equal(t, "http://127.0.0.1:8081", flags.BaseURL)
	assert.Equal(t, "debug", flags.Logging)
	assert.Equal(t, "/tmp/shortener_test.json", flags.FilePath)
	assert.Equal(t, "postgres://user:password@localhost:5432/dbname", flags.DatabaseDsn)

	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("LOG_LVL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("DATABASE_DSN")
}

func TestInitConfig(t *testing.T) {
	os.Setenv("SERVER_ADDRESS", "localhost:8082")
	os.Setenv("BASE_URL", "http://localhost:8082")
	os.Setenv("LOG_LVL", "info")
	os.Setenv("FILE_STORAGE_PATH", "/tmp/shortener_config.json")
	os.Setenv("DATABASE_DSN", "")

	cfg, err := InitConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, "localhost:8082", cfg.Address)
	assert.Equal(t, "http://localhost:8082", cfg.BaseURL)
	assert.Equal(t, "info", cfg.Logging)
	assert.Equal(t, "/tmp/shortener_config.json", cfg.FileStoragePath)
	assert.Empty(t, cfg.DatabaseDsn)

	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("LOG_LVL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("DATABASE_DSN")
}

func getFlagsForTesting() FlagsConfig {
	fs := flag.NewFlagSet("testFlags", flag.ExitOnError)

	tempAddress := fs.String("a", "localhost:8080", "address to start the HTTP server")
	tempBaseURL := fs.String("b", "http://localhost:8080", "the URL for the shortURL")
	tempLogging := fs.String("log-level", "info", "logging for INFO lvl")
	tempFilePath := fs.String("f", getDefaultFilePath(), "file for urls data")
	tempDB := fs.String("d", "", "Postgre DSN (Data Source Name)")

	fs.Parse([]string{}) // Очищаем стандартные флаги перед парсингом

	address := os.Getenv("SERVER_ADDRESS")
	baseURL := os.Getenv("BASE_URL")
	logging := os.Getenv("LOG_LVL")
	filePath := os.Getenv("FILE_STORAGE_PATH")
	dbDSN := os.Getenv("DATABASE_DSN")

	if address == "" {
		address = *tempAddress
	}
	if baseURL == "" {
		baseURL = *tempBaseURL
	}
	if filePath == "" {
		filePath = *tempFilePath
	}
	if dbDSN == "" {
		dbDSN = *tempDB
	}
	if logging == "" {
		logging = *tempLogging
	}

	return FlagsConfig{
		Address:     address,
		BaseURL:     baseURL,
		FilePath:    filePath,
		Logging:     logging,
		DatabaseDsn: dbDSN,
	}
}
