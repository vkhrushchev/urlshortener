package dto

type APICreateShortURLRequest struct {
	URL string `json:"url"`
}

type APICreateShortURLResponse struct {
	Result           string `json:"result,omitempty"`
	ErrorStatus      string `json:"error_status,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}
