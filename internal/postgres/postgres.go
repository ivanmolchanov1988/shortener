package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/ivanmolchanov1988/shortener/internal/storage"
	"github.com/lib/pq"
)

type PostgresStorage struct {
	db                   *sql.DB
	insertStmt           *sql.Stmt
	selectStmt           *sql.Stmt
	selectByOrig         *sql.Stmt
	selectUrlsFromUserID *sql.Stmt
}

type PostgresTransaction struct {
	tx *sql.Tx
}

// Для транзакций 1
func (p *PostgresStorage) BeginTransaction() (storage.TransactionStorage, error) {
	if p.db == nil {
		return nil, errors.New("database connection is nil")
	}
	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}
	return &PostgresTransaction{tx: tx}, nil
}

// Для транзакций 2
func (t *PostgresTransaction) SaveURLTx(id, shortURL, originalURL, userID string) (string, error) {
	var existingShortURL string
	query := `
    INSERT INTO urls (id, short_url, original_url, user_id, created_at, updated_at)
    VALUES ($1, $2, $3, $4, DEFAULT, DEFAULT)
    ON CONFLICT (original_url) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
    RETURNING short_url;`
	err := t.tx.QueryRow(query, id, shortURL, originalURL, userID).Scan(&existingShortURL)
	if err != nil {
		return "", fmt.Errorf("failed to save URL: %w", err)
	}

	// Конфликт?
	if existingShortURL != shortURL {
		// URL уже существует, возвращаем существующий shortURL и ошибку
		//return existingShortURL, ErrURLAlreadyExists
		return existingShortURL, storage.ErrURLAlreadyExists
	}

	log.Printf("Saving URL: id=%s, shortURL=%s, originalURL=%s, userID=%s", id, shortURL, originalURL, userID)

	return existingShortURL, nil
}

// Commit завершает транзакцию
func (t *PostgresTransaction) Commit() error {
	return t.tx.Commit()
}

// Rollback откатывает транзакцию
func (t *PostgresTransaction) Rollback() error {
	return t.tx.Rollback()
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

	// INSERT
	insertStmt, err := db.Prepare(`
		INSERT INTO urls (id, short_url, original_url, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, DEFAULT, DEFAULT)
		ON CONFLICT (original_url) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
		RETURNING short_url;`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare insert statement: %w", err)
	}

	// SELECT
	selectStmt, err := db.Prepare(`
		SELECT original_url FROM urls WHERE short_url = $1;`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select statement: %w", err)
	}

	// SELECT URLs FROM USER ID
	selectUrlsFromUserID, err := db.Prepare(`
		SELECT short_url, original_url FROM urls WHERE user_id = $1`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select URLs: %w", err)
	}

	// SELECT by orig
	selectByOrig, err := db.Prepare(`
		SELECT short_url FROM urls WHERE original_url = $1;`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select by original statement: %w", err)
	}

	//return storage, nil
	return &PostgresStorage{
		db:                   db,
		insertStmt:           insertStmt,
		selectStmt:           selectStmt,
		selectByOrig:         selectByOrig,
		selectUrlsFromUserID: selectUrlsFromUserID,
	}, nil
}

func (p *PostgresStorage) createTable() error {
	query := `
    CREATE TABLE IF NOT EXISTS urls (
        id UUID PRIMARY KEY,
        short_url TEXT NOT NULL,
        original_url TEXT UNIQUE NOT NULL,
		user_id UUID,
		delete_flag BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
    );`
	_, err := p.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

//var ErrURLAlreadyExists = errors.New("URL already exists")

func (p *PostgresStorage) SaveURL(id, shortURL, originalURL, userID string) (string, error) {
	log.Printf("Saving URL: id=%s, shortURL=%s, originalURL=%s, userID=%s", id, shortURL, originalURL, userID)
	var existingShortURL string
	err := p.insertStmt.QueryRow(id, shortURL, originalURL, userID).Scan(&existingShortURL)
	if err != nil {
		log.Printf("Failed to save URL: %v", err)
		return "", fmt.Errorf("failed to save URL: %w", err)
	}

	// Конфликт?
	if existingShortURL != shortURL {
		// URL уже существует, возвращаем существующий shortURL и ошибку
		return existingShortURL, storage.ErrURLAlreadyExists
	}

	log.Printf("URL saved successfully: %s", existingShortURL)
	log.Printf("Saving URL: id=%s, shortURL=%s, originalURL=%s, userID=%s", id, shortURL, originalURL, userID)

	return existingShortURL, nil
}

func (p *PostgresStorage) GetURL(shortURL string) (string, error) {
	var originalURL string
	err := p.selectStmt.QueryRow(shortURL).Scan(&originalURL)
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
	err := p.selectByOrig.QueryRow(originalURL).Scan(&shortURL)
	if err != nil {
		return "", fmt.Errorf("failed to get short URL by original URL: %w", err)
	}
	return shortURL, nil
}

func (p *PostgresStorage) Close() error {
	if err := p.insertStmt.Close(); err != nil {
		return err
	}
	if err := p.selectStmt.Close(); err != nil {
		return err
	}
	if err := p.selectByOrig.Close(); err != nil {
		return err
	}
	return p.db.Close()
}

// Для списка URLs пользователя
func (p *PostgresStorage) GetUserURLS(userID string) ([]storage.UserURLS, error) {
	log.Printf("URLs for userID: %s", userID)
	rows, err := p.selectUrlsFromUserID.Query(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var usersURLs []storage.UserURLS
	for rows.Next() {
		var url storage.UserURLS
		if err := rows.Scan(&url.ShortURL, &url.OriginalURL); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		usersURLs = append(usersURLs, url)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return usersURLs, nil
}

// Для удаления URLs
func (p *PostgresStorage) DeleteURLS(userID string, urlsToDelete []string) error {
	log.Printf("The user's %s URLs to delete", userID)

	urlsChan := make(chan string, len(urlsToDelete))
	go func() {
		for _, url := range urlsToDelete {
			urlsChan <- url
		}
		close(urlsChan)
	}()

	// Паттерн fanIn: объединяем данные из канала в батчи
	const batchSize = 2
	batchChan := make(chan []string)

	go func() {
		var batch []string
		for url := range urlsChan {
			batch = append(batch, url)
			if len(batch) == batchSize {
				batchChan <- batch
				batch = nil
			}
		}
		if len(batch) > 0 {
			batchChan <- batch
		}
		close(batchChan)
	}()

	// Обработка пачек
	var wg sync.WaitGroup

	var errResult error
	for batch := range batchChan {
		wg.Add(1)
		go func(batch []string) {
			defer wg.Done()

			query := `
				UPDATE urls
				SET delete_flag = TRUE
				WHERE user_id = $1 AND short_url = ANY($2);
			`
			_, err := p.db.Exec(query, userID, pq.Array(batch))
			if err != nil {
				if errResult != nil {
					errResult = fmt.Errorf("failed to update batch %v: %w", batch, err)
				}
			} else {
				log.Printf("Successfully deleted: %v", batch)
			}
		}(batch)
	}

	wg.Wait()

	// УДАЛЯЕМ
	if err := p.HardDeleteURLs(); err != nil {
		log.Printf("Failed to hard delete marked URLs: %v", err)
	}

	return errResult

}

// Реальное удаление - плохо, но, вроде, требует Яндекс
func (p *PostgresStorage) HardDeleteURLs() error {
	query := `
		DELETE FROM urls
		WHERE delete_flag = TRUE;
	`

	result, err := p.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to hard delete URLs: %w", err)
	}

	rows, _ := result.RowsAffected()
	log.Printf("Hard deleted %d URLs", rows)

	return nil
}
