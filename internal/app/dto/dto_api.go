package dto

type APICreateShortURLRequest struct {
	URL string `json:"url"`
}

type APICreateShortURLResponse struct {
	Result           string `json:"result,omitempty"`
	ErrorStatus      string `json:"error_status,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}

type APICreateShortURLBatchRequest []APICreateShortURLBatchRequestEntry

type APICreateShortURLBatchRequestEntry struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type APICreateShortURLBatchResponse []APICreateShortURLBatchResponseEntry

type APICreateShortURLBatchResponseEntry struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
