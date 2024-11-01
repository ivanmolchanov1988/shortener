package filestore

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/ivanmolchanov1988/shortener/internal/storage"
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

var _ storage.Storage = (*FileStorage)(nil)

func NewFileStorage(filePath string) (*FileStorage, error) {
	fs := &FileStorage{
		filePath:      filePath,
		shortLinkData: []ShortLinkData{},
	}

	if _, err := fs.LoadDataFromFile(); err != nil {
		return nil, err
	}

	return fs, nil
}

func (f *FileStorage) SaveURL(id, shortURL, originalURL string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	newShortLinkData := ShortLinkData{
		UUID:        id,
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
	f.shortLinkData = append(f.shortLinkData, newShortLinkData)

	return nil
}

func (f *FileStorage) GetURL(shortURL string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, data := range f.shortLinkData {
		if data.ShortURL == shortURL {
			return data.OriginalURL, nil
		}
	}

	return "", errors.New("URL not found")
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
	f.mu.Lock()
	defer f.mu.Unlock()

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

	f.shortLinkData = data
	return data, nil
}
