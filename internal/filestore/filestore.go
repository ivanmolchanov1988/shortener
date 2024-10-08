package filestore

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"

	"github.com/ivanmolchanov1988/shortener/pkg/utils"
)

type ShortLinkData struct {
	UUID        string `json:"id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorage struct {
	filePath      string
	shortLinkData []ShortLinkData
	mu            sync.RWMutex
}

func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{
		filePath:      filePath,
		shortLinkData: []ShortLinkData{},
	}
}

func (f *FileStorage) SaveURL(shortURL, originalURL string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	uuid := utils.GenUUID()
	newShortLinkData := ShortLinkData{
		UUID:        uuid,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}

	// Файл уже есть. Проверка в main.
	file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(newShortLinkData); err != nil {
		return err
	}

	return nil
}

func (f *FileStorage) saveData(data []ShortLinkData) error {
	file, err := os.Create(f.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

func (f *FileStorage) LoadDataFromFile() ([]ShortLinkData, error) {
	file, err := os.Open(f.filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data []ShortLinkData
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var ok ShortLinkData
		line := scanner.Text()
		if err := json.Unmarshal([]byte(line), &ok); err != nil {
			return nil, err
		}
		data = append(data, ok)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return data, nil
}
