package core

// GetUserURLsRequest - запрос на получение всех ссылок пользователя.
type GetUserURLsRequest struct {
	UserID string
}

// UserURL - структура для хранения информации о сокращённой ссылке пользователя.
type UserURL struct {
	ShortURL    string
	OriginalURL string
}

// GetUserURLsResponse - ответ с массивом сокращённых ссылок пользователя.
type GetUserURLsResponse struct {
	URLs []UserURL
}
