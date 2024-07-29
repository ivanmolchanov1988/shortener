package memory

import (
	"errors"
	"sync"
)

type MemoryStorage struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

func (m *MemoryStorage) SaveURL(shortURL, originalURL string) error {
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
		return "", errors.New("the URL not found")
	}
	return originalURL, nil
}
