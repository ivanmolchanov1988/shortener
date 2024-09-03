package memory

import (
	"errors"
	"sync"

	"github.com/ivanmolchanov1988/shortener/internal/file"
)

type MemoryStorage struct {
	data        map[string]string
	fileStorage *file.FileStorage
	mu          sync.RWMutex
}

//	func NewMemoryStorage() *MemoryStorage {
//		return &MemoryStorage{
//			data: make(map[string]string),
//		}
//	}
func NewMemoryStorage(fileStorage *file.FileStorage) (*MemoryStorage, error) {
	memStorage := &MemoryStorage{
		data:        make(map[string]string),
		fileStorage: fileStorage,
	}

	// Загрузка даты из файла
	if err := memStorage.loadDataFromFile(); err != nil {
		return nil, err
	}

	return memStorage, nil
}

func (m *MemoryStorage) SaveURL(shortURL, originalURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[shortURL] = originalURL
	// в файл
	m.fileStorage.SaveURL(shortURL, originalURL)

	return nil
}

func (m *MemoryStorage) GetURL(shortURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	originalURL, exists := m.data[shortURL]
	if !exists {
		return "", errors.New("the URL not found")
	}
	return originalURL, nil
}

func (m *MemoryStorage) loadDataFromFile() error {
	data, err := m.fileStorage.LoadDataFromFile()
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, item := range data {
		m.data[item.ShortURL] = item.OriginalURL
	}

	return nil
}
