package dto

type ApiCreateShortURLRequest struct {
	Url string `json:"url"`
}

type ApiCreateShortURLResponse struct {
	Result           string `json:"result,omitempty"`
	ErrorStatus      string `json:"error_status,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}
