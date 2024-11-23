package storage

import (
	"errors"
)

var ErrURLAlreadyExists = errors.New("URL already exists")

// Storage для всех типов хранилищ
type Storage interface {
	SaveURL(id, shortURL, originalURL string) (string, error)
	GetURL(shortURL string) (string, error)
	BeginTransaction() (TransactionStorage, error)
	//SaveURLTx(tx *sql.Tx, id, shortURL, originalURL string) (string, error)
	//Close()
}

// для операций внутри транзакций
type TransactionStorage interface {
	SaveURLTx(id, shortURL, originalURL string) (string, error)
	Commit() error
	Rollback() error
}
