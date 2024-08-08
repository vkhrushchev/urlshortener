package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vkhrushchev/urlshortener/internal/app/dto"
	"github.com/vkhrushchev/urlshortener/internal/app/storage"
	"github.com/vkhrushchev/urlshortener/internal/util"
)

type ApiController struct {
	baseURL string
	storage storage.Storage
}

func NewApiController(baseURL string, storage storage.Storage) *ApiController {
	return &ApiController{
		baseURL: baseURL,
		storage: storage,
	}
}

func (c *ApiController) CreateShortURLHandlerAPI(w http.ResponseWriter, r *http.Request) {
	apiResponse := &dto.APICreateShortURLResponse{}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Infow(
			"app: Not supported \"Content-Type\" header",
			"Content-Type", contentType,
		)

		apiResponse.ErrorStatus = fmt.Sprintf("%d", http.StatusBadRequest)
		apiResponse.ErrorDescription = fmt.Sprintf("Content-Type = \"%s\" not supported", contentType)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiResponse)

		return
	}

	var apiRequest dto.APICreateShortURLRequest
	if err := json.NewDecoder(r.Body).Decode(&apiRequest); err != nil {
		log.Errorw(
			"app: Error when decode request body from json",
			"erorr", err.Error(),
		)

		apiResponse.ErrorStatus = fmt.Sprintf("%d", http.StatusBadRequest)
		apiResponse.ErrorDescription = fmt.Sprintf("Error when decoding request body: %s", err.Error())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiResponse)

		return
	}

	longURL := apiRequest.URL
	shortURI, err := c.storage.SaveURL(r.Context(), longURL)
	if err != nil {
		apiResponse.ErrorStatus = fmt.Sprintf("%d", http.StatusInternalServerError)
		apiResponse.ErrorDescription = fmt.Sprintf("Error when saving short URL: %s", err.Error())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiResponse)

		return
	}

	log.Infow(
		"app: short_url_added",
		"shortURI", shortURI,
		"longURL", longURL,
	)

	apiResponse.Result = util.GetShortURL(c.baseURL, shortURI)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(apiResponse)
}
