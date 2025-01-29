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
	UserID      string `json:"user_id"`
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

func (f *FileStorage) SaveURL(id, shortURL, originalURL, userID string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Cуществует такой originalURL?
	for _, data := range f.shortLinkData {
		if data.OriginalURL == originalURL {
			return data.ShortURL, storage.ErrURLAlreadyExists
		}
	}

	newShortLinkData := ShortLinkData{
		UUID:        id,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
	}

	// Файл уже есть. Проверка в main.
	file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(newShortLinkData); err != nil {
		return "", err
	}
	f.shortLinkData = append(f.shortLinkData, newShortLinkData)

	return shortURL, nil
}

func (f *FileStorage) GetURL(shortURL string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, data := range f.shortLinkData {
		if data.ShortURL == shortURL {
			return data.OriginalURL, nil
		}
	}

	//return "", errors.New("URL not found")
	return "", storage.ErrURLNotFound
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

// PASS для БД BeginTransaction
func (f *FileStorage) BeginTransaction() (storage.TransactionStorage, error) {
	return nil, errors.New("transaction is not in FileStorage")
}

// PASS для БД SaveURLTx
func (f *FileStorage) SaveURLTx(id, shortURL, originalURL, userID string) (string, error) {
	return f.SaveURL(id, shortURL, originalURL, userID)
}

// PASS для URLs для user
func (f *FileStorage) GetUserURLS(userID string) ([]storage.UserURLS, error) {
	return nil, errors.New("get URLs for user is not in FileStorage")
}

// PASS для URLsForDelete
func (f *FileStorage) DeleteURLS(userID string, urlsTodelete []string) error {
	return errors.New("delete URLs is not in FileStorage")
}
