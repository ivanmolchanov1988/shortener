package core

// ShortenRequest - запрос на сокращение ссылки.
type ShortenRequest struct {
	OriginalURL string
}

// ShortenResponse - ответ с сокращённой ссылкой.
type ShortenResponse struct {
	ShortURL string
}
