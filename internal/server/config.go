// Package server загрузка конфигурации из JSON файла.
package server

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadConfig - загружает конфигурацию из JSON-файла, если указан `-c`/`-config`.
// Возвращает загруженный конфиг и `true`, если удалось загрузить файл, иначе `false`.
func loadConfig(filePath string) (*Config, bool, error) {
	if filePath == "" {
		return nil, false, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, false, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, false, fmt.Errorf("failed to decode JSON config: %w", err)
	}

	return &cfg, true, nil
}
