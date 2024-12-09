package dto

// APICreateShortURLRequest запрос на создание короткой ссылки
type APICreateShortURLRequest struct {
	URL string `json:"url"` // URL - url для которой требуется получить короткую ссылку
}

// APICreateShortURLResponse ответ на запрос на создание короткой ссылки
type APICreateShortURLResponse struct {
	Result           string `json:"result,omitempty"`            // Короткая ссылка
	ErrorStatus      string `json:"error_status,omitempty"`      // Статус в случае ошибки
	ErrorDescription string `json:"error_description,omitempty"` // Описание ошибки
}

// APICreateShortURLBatchRequest запрос на создание коротких ссылок пачкой
type APICreateShortURLBatchRequest []APICreateShortURLBatchRequestEntry

type APICreateShortURLBatchRequestEntry struct {
	CorrelationID string `json:"correlation_id"` // Идентификатор для корреляции
	OriginalURL   string `json:"original_url"`   // Оригинальная ссылка
}

// APICreateShortURLBatchResponse ответ на запрос коротких ссылок пачкой
type APICreateShortURLBatchResponse []APICreateShortURLBatchResponseEntry

type APICreateShortURLBatchResponseEntry struct {
	CorrelationID string `json:"correlation_id"` // Идентификатор для корреляции
	ShortURL      string `json:"short_url"`      // Коротка ссылка
}

// APIGetAllURLByUserIDResponse ответ на получение коротких ссылок созданных пользователем
type APIGetAllURLByUserIDResponse []APIGetAllURLByUserIDResponseEntry

type APIGetAllURLByUserIDResponseEntry struct {
	ShortURL    string `json:"short_url"`    // Короткая ссылка
	OriginalURL string `json:"original_url"` // Оригинальная ссылка
}
