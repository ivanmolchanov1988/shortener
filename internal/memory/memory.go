package memory

import (
	"database/sql"
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

func (m *MemoryStorage) SaveURL(id, shortURL, originalURL string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Cуществует уже такой originalURL?
	for existingShortURL, existingOriginalURL := range m.data {
		if existingOriginalURL == originalURL {
			// Возвращаем существующий shortURL и ошибку
			return existingShortURL, storage.ErrURLAlreadyExists
		}
	}

	m.data[shortURL] = originalURL
	return shortURL, nil
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

// PASS для БД BeginTransaction
func (m *MemoryStorage) BeginTransaction() (*sql.Tx, error) {
	return nil, errors.New("transaction is not in MemoryStorage")
}

// PASS для БД SaveURLTx
func (m *MemoryStorage) SaveURLTx(tx *sql.Tx, id, shortURL, originalURL string) (string, error) {
	return m.SaveURL(id, shortURL, originalURL)
}

var _ storage.Storage = (*MemoryStorage)(nil)
