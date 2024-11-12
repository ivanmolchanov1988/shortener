package postgr

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"

	"github.com/ivanmolchanov1988/shortener/internal/storage"
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
func (p *PostgresStorage) SaveURLTx(tx *sql.Tx, id, shortURL, originalURL string) (string, error) {
	var existingShortURL string
	query := `
    INSERT INTO urls (id, short_url, original_url, created_at, updated_at)
    VALUES ($1, $2, $3, DEFAULT, DEFAULT)
    ON CONFLICT (original_url) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
    RETURNING short_url;`
	err := tx.QueryRow(query, id, shortURL, originalURL).Scan(&existingShortURL)
	if err != nil {
		// Ошибка - конфликт уникальности?
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			// Получаем short_url для original_url
			existingShortURL, getErr := p.GetShortURLByOriginalURLTx(tx, originalURL)
			if getErr != nil {
				return "", fmt.Errorf("failed to get existing short URL: %w", getErr)
			}
			return existingShortURL, storage.ErrURLAlreadyExists
		}
		return "", fmt.Errorf("failed to save URL: %w", err)
	}
	return existingShortURL, nil
}

func (p *PostgresStorage) GetShortURLByOriginalURLTx(tx *sql.Tx, originalURL string) (string, error) {
	var shortURL string
	query := `SELECT short_url FROM urls WHERE original_url = $1;`
	err := tx.QueryRow(query, originalURL).Scan(&shortURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("URL not found")
		}
		return "", fmt.Errorf("failed to get short URL by original URL: %w", err)
	}
	return shortURL, nil
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
        original_url TEXT UNIQUE NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
    );`
	_, err := p.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

var ErrURLAlreadyExists = errors.New("URL already exists")

func (p *PostgresStorage) SaveURL(id, shortURL, originalURL string) (string, error) {
	var existingShortURL string
	query := `
    INSERT INTO urls (id, short_url, original_url, created_at, updated_at)
    VALUES ($1, $2, $3, DEFAULT, DEFAULT)
    ON CONFLICT (original_url) DO UPDATE 
    RETURNING short_url;`
	err := p.db.QueryRow(query, id, shortURL, originalURL).Scan(&existingShortURL)
	if err != nil {
		// Была ошибка ошибкой уникальности?
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			existingShortURL, getErr := p.GetShortURLByOriginalURL(originalURL)
			if getErr != nil {
				return "", fmt.Errorf("failed to get existing shortURL: %w", getErr)
			}
			return existingShortURL, ErrURLAlreadyExists
		}
		return "", fmt.Errorf("failed to save URL: %w", err)
	}
	return existingShortURL, nil
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

func (p *PostgresStorage) GetShortURLByOriginalURL(originalURL string) (string, error) {
	var shortURL string
	query := `
	SELECT short_url FROM urls WHERE original_url = $1;`
	err := p.db.QueryRow(query, originalURL).Scan(&shortURL)
	if err != nil {
		return "", fmt.Errorf("failed to get short URL by original URL: %w", err)
	}
	return shortURL, nil
}
