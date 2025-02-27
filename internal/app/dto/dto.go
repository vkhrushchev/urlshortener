package dto

// APICreateShortURLRequest структура с описанием запроса на создание короткой ссылки
type APICreateShortURLRequest struct {
	URL string `json:"url"`
}

// APICreateShortURLResponse структура с описанием ответа на запрос на создание короткой ссылки
type APICreateShortURLResponse struct {
	Result           string `json:"result,omitempty"`
	ErrorStatus      string `json:"error_status,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// APICreateShortURLBatchRequest слайс запроса на создание коротких ссылок пачкой
type APICreateShortURLBatchRequest []APICreateShortURLBatchRequestEntry

// APICreateShortURLBatchRequestEntry вхождение в слайс APICreateShortURLBatchRequest
type APICreateShortURLBatchRequestEntry struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// APICreateShortURLBatchResponse слайс ответа на запрос на создание коротких ссылок пачкой
type APICreateShortURLBatchResponse []APICreateShortURLBatchResponseEntry

// APICreateShortURLBatchResponseEntry вхождение в слайс APICreateShortURLBatchResponse
type APICreateShortURLBatchResponseEntry struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// APIGetAllURLByUserIDResponse слайс ответа на запрос на получение коротких ссылок пользователя
type APIGetAllURLByUserIDResponse []APIGetAllURLByUserIDResponseEntry

// APIGetAllURLByUserIDResponseEntry вхождение в слайс APIGetAllURLByUserIDResponse
type APIGetAllURLByUserIDResponseEntry struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// APIInternalGetStatsResponse ответ на запрос статистики
type APIInternalGetStatsResponse struct {
	URLCount  int `json:"urls"`
	UserCount int `json:"users"`
}
