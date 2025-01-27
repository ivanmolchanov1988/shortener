package storage

import (
	"errors"
)

var (
	ErrURLAlreadyExists = errors.New("URL already exists")
	ErrURLIsGone        = errors.New("url is gone")
	ErrURLNotFound      = errors.New("URL not found")
)

// Storage для всех типов хранилищ
type Storage interface {
	SaveURL(id, shortURL, originalURL, userID string) (string, error)
	GetURL(shortURL string) (string, error)
	BeginTransaction() (TransactionStorage, error)
	GetUserURLS(userID string) ([]UserURLS, error)
	DeleteURLS(userID string, urlsTodelete []string) error
}

// для операций внутри транзакций
type TransactionStorage interface {
	SaveURLTx(id, shortURL, originalURL, userID string) (string, error)
	Commit() error
	Rollback() error
}

type UserURLS struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
