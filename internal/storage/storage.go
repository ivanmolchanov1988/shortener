package storage

import "database/sql"

// Storage для всех типов хранилищ
type Storage interface {
	SaveURL(id, shortURL, originalURL string) error
	GetURL(shortURL string) (string, error)
	BeginTransaction() (*sql.Tx, error)
	SaveURLTx(tx *sql.Tx, id, shortURL, originalURL string) error
	//Close()
}
