package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/vkhrushchev/urlshortener/internal/app/dto"
	"github.com/vkhrushchev/urlshortener/internal/app/storage"
	"github.com/vkhrushchev/urlshortener/internal/middleware"
	"github.com/vkhrushchev/urlshortener/internal/util"
)

type APIController struct {
	baseURL string
	storage storage.Storage
}

func NewAPIController(baseURL string, storage storage.Storage) *APIController {
	return &APIController{
		baseURL: baseURL,
		storage: storage,
	}
}

func (c *APIController) CreateShortURLHandler(w http.ResponseWriter, r *http.Request) {
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
	shortURLEntry, err := c.storage.SaveURL(r.Context(), longURL)
	if err != nil && !errors.Is(err, storage.ErrConflictOnUniqueConstraint) {
		apiResponse.ErrorStatus = fmt.Sprintf("%d", http.StatusInternalServerError)
		apiResponse.ErrorDescription = fmt.Sprintf("Error when saving short URL: %s", err.Error())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(apiResponse)

		return
	}

	log.Infow(
		"app: short_url_added",
		"shortURI", shortURLEntry.ShortURI,
		"longURL", shortURLEntry.LongURL,
	)

	apiResponse.Result = util.GetShortURL(c.baseURL, shortURLEntry.ShortURI)

	w.Header().Set("Content-Type", "application/json")
	if err != nil && errors.Is(err, storage.ErrConflictOnUniqueConstraint) {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	json.NewEncoder(w).Encode(apiResponse)
}

func (c *APIController) CreateShortURLBatchHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Infow(
			"app: not supported \"Content-Type\" header",
			"Content-Type", contentType,
		)

		w.WriteHeader(http.StatusBadRequest)

		return
	}

	var apiRequest dto.APICreateShortURLBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&apiRequest); err != nil {
		log.Errorw(
			"app: error when decode request body from json",
			"erorr", err.Error(),
		)

		w.WriteHeader(http.StatusBadRequest)

		return
	}

	storageShortURLEntries := make([]*dto.StorageShortURLEntry, 0, len(apiRequest))
	for _, apiShortURLEntry := range apiRequest {
		storageShortURLEntry := &dto.StorageShortURLEntry{
			UUID:    apiShortURLEntry.CorrelationID,
			LongURL: apiShortURLEntry.OriginalURL,
		}
		storageShortURLEntries = append(storageShortURLEntries, storageShortURLEntry)
	}

	storageShortURLEntries, err := c.storage.SaveURLBatch(r.Context(), storageShortURLEntries)
	if err != nil {
		log.Errorw(
			"app: error when store batch of URLs",
			"error", err.Error(),
		)

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	apiResponse := make(dto.APICreateShortURLBatchResponse, 0, len(apiRequest))
	for _, storageShortURLEntry := range storageShortURLEntries {
		apiResponseEntry := dto.APICreateShortURLBatchResponseEntry{
			CorrelationID: storageShortURLEntry.UUID,
			ShortURL:      util.GetShortURL(c.baseURL, storageShortURLEntry.ShortURI),
		}
		apiResponse = append(apiResponse, apiResponseEntry)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(apiResponse)
}

func (c *APIController) GetShortURLByUserID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(string)
	storageEntries, err := c.storage.GetURLByUserID(r.Context(), userID)
	if err != nil {
		log.Errorw(
			"app: error when get shortURL by userID",
			"erorr", err.Error(),
			"userID", userID,
		)

		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if len(storageEntries) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	apiResponse := make(dto.APIGetAllURLByUserIDResponse, 0, len(storageEntries))
	for _, storageEntry := range storageEntries {
		apiResponseEntry := dto.APIGetAllURLByUserIDResponseEntry{
			ShortURL:    util.GetShortURL(c.baseURL, storageEntry.ShortURI),
			OriginalURL: storageEntry.LongURL,
		}
		apiResponse = append(apiResponse, apiResponseEntry)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiResponse)
}

func (c *APIController) DeleteShortURLs(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Infow(
			"app: not supported \"Content-Type\" header",
			"Content-Type", contentType,
		)

		w.WriteHeader(http.StatusBadRequest)

		return
	}

	var apiRequest []string
	if err := json.NewDecoder(r.Body).Decode(&apiRequest); err != nil {
		log.Errorw(
			"app: error when decode request body from json",
			"erorr", err.Error(),
		)

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := c.storage.DeleteByShortURIs(context.WithoutCancel(r.Context()), apiRequest)
	if err != nil {
		log.Errorw(
			"app: error when delete by shortURIs",
			"erorr", err.Error(),
		)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
