package storage

import (
	"database/sql"
	"errors"
)

var ErrURLAlreadyExists = errors.New("URL already exists")

// Storage для всех типов хранилищ
type Storage interface {
	SaveURL(id, shortURL, originalURL string) (string, error)
	GetURL(shortURL string) (string, error)
	BeginTransaction() (*sql.Tx, error)
	SaveURLTx(tx *sql.Tx, id, shortURL, originalURL string) (string, error)
	//Close()
}
