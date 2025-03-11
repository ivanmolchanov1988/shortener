package core

// GetStatsRequest - запрос на получение статистики (с IP клиента).
type GetStatsRequest struct {
	ClientIP string
	Subnet   string
}

// GetStatsResponse - ответ с количеством пользователей и сокращённых ссылок.
type GetStatsResponse struct {
	URLs  int
	Users int
}
