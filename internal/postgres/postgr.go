package postgr

import (
	"database/sql"
	"errors"
	"fmt"
)

type PostgresStorage struct {
	db *sql.DB
}

// Для транзакций 1
func (p *PostgresStorage) BeginTransaction() (*sql.Tx, error) {
	if p.db == nil {
		return nil, errors.New("database connection is nil")
	}
	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// Для транзакций 2
func (p *PostgresStorage) SaveURLTx(tx *sql.Tx, id, shortURL, originalURL string) error {
	if tx == nil {
		return errors.New("transaction is nil")
	}
	query := `INSERT INTO urls (id, short_url, original_url) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING`
	_, err := tx.Exec(query, id, shortURL, originalURL)
	if err != nil {
		return err
	}
	return nil
}

func NewPostgresStorage(db *sql.DB) (*PostgresStorage, error) {
	storage := &PostgresStorage{db: db}
	// Проверяем наличие таблицы
	if err := storage.createTable(); err != nil {
		return nil, err
	}
	return storage, nil
}

func (p *PostgresStorage) createTable() error {
	query := `
    CREATE TABLE IF NOT EXISTS urls (
        id UUID PRIMARY KEY,
        short_url TEXT UNIQUE NOT NULL,
        original_url TEXT NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
    );`
	_, err := p.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

func (p *PostgresStorage) SaveURL(id, shortURL, originalURL string) error {
	query := `
    INSERT INTO urls (id, short_url, original_url, created_at, updated_at)
    VALUES ($1, $2, $3, DEFAULT, DEFAULT)
    ON CONFLICT (short_url) DO UPDATE
    SET original_url = EXCLUDED.original_url,
        updated_at = CURRENT_TIMESTAMP;`
	_, err := p.db.Exec(query, id, shortURL, originalURL)
	if err != nil {
		return fmt.Errorf("failed to save URL: %w", err)
	}
	return nil
}

func (p *PostgresStorage) GetURL(shortURL string) (string, error) {
	var originalURL string
	query := `SELECT original_url FROM urls WHERE short_url = $1;`
	err := p.db.QueryRow(query, shortURL).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("URL not found")
		}
		return "", fmt.Errorf("failed to get URL: %w", err)
	}
	return originalURL, nil
}
