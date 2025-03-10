// Package memory реализует хранение данных в памяти (in-memory storage) для работы сервиса без базы данных.
package memory

import (
	"errors"
	"sync"

	//"github.com/ivanmolchanov1988/shortener/internal/filestore"
	"github.com/ivanmolchanov1988/shortener/internal/storage"
)

// MemoryStorage - структура для хранения данных в памяти.
type MemoryStorage struct {
	data map[string]string
	//fileStorage *filestore.FileStorage
	mu sync.RWMutex
}

// NewMemoryStorage возвращает мапу MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]string),
	}
}

// MemoryTransaction - PASS transaction
type MemoryTransaction struct{}

// SaveURL сохраняет URLs в MemoryStorage.
func (m *MemoryStorage) SaveURL(id, shortURL, originalURL, userID string) (string, error) {
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

// GetURL забирает URLs из MemoryStorage.
func (m *MemoryStorage) GetURL(shortURL string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	originalURL, exists := m.data[shortURL]
	if !exists {
		return "", storage.ErrURLNotFound
	}
	return originalURL, nil
}

// BeginTransaction - PASS для БД.
func (m *MemoryStorage) BeginTransaction() (storage.TransactionStorage, error) {
	return nil, errors.New("transaction is not in MemoryStorage")
}

// SaveURLTx - PASS для БД.
func (m *MemoryStorage) SaveURLTx(id, shortURL, originalURL string) (string, error) {
	//return m.SaveURL(id, shortURL, originalURL)
	return "", errors.New("transactions are not supported in MemoryStorage")
}

// Commit - PASS.
func (m *MemoryTransaction) Commit() error {
	return errors.New("transactions are not supported in MemoryStorage")
}

// Rollback - PASS.
func (m *MemoryTransaction) Rollback() error {
	return errors.New("transactions are not supported in MemoryStorage")
}

// DeleteURLS - PASS URLsForDelete.
func (m *MemoryStorage) DeleteURLS(userID string, urlsTodelete []string) error {
	return errors.New("delete URLs for user are not supported in MemoryStorage")
}

// GetUserURLS - PASS.
func (m *MemoryStorage) GetUserURLS(userID string) ([]storage.UserURLS, error) {
	return nil, errors.New("get URLs for user are not supported in MemoryStorage")
}

// GetStats - PASS.
func (m *MemoryStorage) GetStats() (storage.Stats, error) {
	return storage.Stats{}, errors.New("get stats is not supported in MemoryStorage")
}

var _ storage.Storage = (*MemoryStorage)(nil)
