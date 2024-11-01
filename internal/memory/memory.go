package memory

import (
	"errors"
	"sync"

	"github.com/ivanmolchanov1988/shortener/internal/filestore"
	"github.com/ivanmolchanov1988/shortener/internal/storage"
)

type MemoryStorage struct {
	data        map[string]string
	fileStorage *filestore.FileStorage
	mu          sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

func (m *MemoryStorage) SaveURL(id, shortURL, originalURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[shortURL] = originalURL
	return nil
}

func (m *MemoryStorage) GetURL(shortURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	originalURL, exists := m.data[shortURL]
	if !exists {
		return "", errors.New("URL not found")
	}
	return originalURL, nil
}

var _ storage.Storage = (*MemoryStorage)(nil)
