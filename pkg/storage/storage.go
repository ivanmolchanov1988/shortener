package storage

type Storage interface {
	SaveURL(shortURL, originalURL string) error
	GetURL(shortURL string) (string, error)
}
