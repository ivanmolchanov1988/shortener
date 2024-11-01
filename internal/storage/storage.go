package storage

// Storage для всех типов хранилищ
type Storage interface {
	SaveURL(id, shortURL, originalURL string) error
	GetURL(shortURL string) (string, error)
	//Close()
}
