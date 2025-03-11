// Package storage определяет интерфейсы для взаимодействия с различными хранилищами данных.
package storage

// Storage для всех типов хранилищ.
type Storage interface {
	SaveURL(id, shortURL, originalURL, userID string) (string, error)
	GetURL(shortURL string) (string, error)
	BeginTransaction() (TransactionStorage, error)
	GetUserURLS(userID string) ([]UserURLS, error)
	DeleteURLS(userID string, urlsTodelete []string) error
	GetStats() (Stats, error)
}

// TransactionStorage для операций внутри транзакций.
type TransactionStorage interface {
	SaveURLTx(id, shortURL, originalURL, userID string) (string, error)
	Commit() error
	Rollback() error
}

// UserURLS структура для URLs
type UserURLS struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// BatchURLResult представляет ответ на batch-запрос
type BatchURLResult struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// Closer описывает интерфейс для закрытия хранилища
type Closer interface {
	Close() error
}

// Stats структура для возврата статистики по кол-ву urls и users
type Stats struct {
	URLs  int
	Users int
}
