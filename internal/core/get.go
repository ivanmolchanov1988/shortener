package core

// GetURLRequest - запрос на получение оригинального URL.
type GetURLRequest struct {
	ShortURL string
}

// GetURLResponse - ответ с оригинальным URL.
type GetURLResponse struct {
	OriginalURL string
}
