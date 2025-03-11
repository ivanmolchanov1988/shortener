package core

// DeleteURLsRequest - запрос на удаление ссылок пользователя.
type DeleteURLsRequest struct {
	UserID    string
	ShortURLs []string
}

// DeleteURLsResponse - ответ (статус удаления).
type DeleteURLsResponse struct {
	Status string
}
