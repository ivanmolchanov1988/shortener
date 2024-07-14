package memory

import "errors"

type MemoryStorage struct {
	data map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

func (m *MemoryStorage) SaveURL(shortURL, originalURL string) error {
	if _, exists := m.data[shortURL]; exists {
		return errors.New("the URL already exists")
	}
	m.data[shortURL] = originalURL
	return nil
}

func (m *MemoryStorage) GetURL(shortURL string) (string, error) {
	originalURL, exists := m.data[shortURL]
	if !exists {
		return "", errors.New("the URL not found")
	}
	return originalURL, nil
}
