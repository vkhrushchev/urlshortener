package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vkhrushchev/urlshortener/internal/app/domain"
	"github.com/vkhrushchev/urlshortener/internal/app/usecase"
	"net/http"

	"github.com/vkhrushchev/urlshortener/internal/app/dto"
	"github.com/vkhrushchev/urlshortener/internal/middleware"
	"github.com/vkhrushchev/urlshortener/internal/util"
)

// APIController используется для обработки API-запросов приложения
type APIController struct {
	createShortURLUseCase usecase.ICreateShortURLUseCase // Сценарий создания короткой ссылки
	getShortURLUseCase    usecase.IGetShortURLUseCase    // Сценарий получения короткой ссылки
	deleteShortURLUseCase usecase.IDeleteShortURLUseCase // Сценарий удаления короткой ссылки
	baseURL               string                         // URL до сервера с развернутым приложением
}

// NewAPIController создает новый экземпляр структуры APIController
//
//	baseURL - URL до сервера с развернутым приложением
//	createShortURLUseCase - use case создания короткой ссылки
//	getShortURLUseCase - use case получения короткой ссылки
//	getShortURLUseCase - use case получения короткой ссылки
func NewAPIController(
	baseURL string,
	createShortURLUseCase usecase.ICreateShortURLUseCase,
	getShortURLUseCase usecase.IGetShortURLUseCase,
	deleteShortURLUseCase usecase.IDeleteShortURLUseCase,
) *APIController {
	return &APIController{
		baseURL:               baseURL,
		createShortURLUseCase: createShortURLUseCase,
		getShortURLUseCase:    getShortURLUseCase,
		deleteShortURLUseCase: deleteShortURLUseCase,
	}
}

// CreateShortURLHandler обрабатывает запрос на создание короткой ссылки
//
//	@Summary	Создание короткой ссылки
//	@Accepts	json
//	@Produce	json
//	@Success	201	{object}	dto.APICreateShortURLResponse
//	@Failure	400	{object}	dto.APICreateShortURLResponse	"ошибка в формате запроса"
//	@Success	409	{object}	dto.APICreateShortURLResponse	"короткая ссылка уже существует"
//	@Failure	500	{object}	dto.APICreateShortURLResponse	"внутренняя ошибка сервиса"
//	@Router		/api/shorten [post]
//	@Param		body	body	dto.APICreateShortURLRequest	true "запрос на создание короткой ссылки"
func (c *APIController) CreateShortURLHandler(w http.ResponseWriter, r *http.Request) {
	apiResponse := &dto.APICreateShortURLResponse{}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		apiResponse.ErrorStatus = fmt.Sprintf("%d", http.StatusBadRequest)
		apiResponse.ErrorDescription = fmt.Sprintf("Content-Type = \"%s\" not supported", contentType)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiResponse)

		return
	}

	var apiRequest dto.APICreateShortURLRequest
	if err := json.NewDecoder(r.Body).Decode(&apiRequest); err != nil {
		log.Errorw("app: error when decode request body from json", "err", err)

		apiResponse.ErrorStatus = fmt.Sprintf("%d", http.StatusBadRequest)
		apiResponse.ErrorDescription = fmt.Sprintf("Error when decoding request body: %s", err.Error())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiResponse)

		return
	}

	longURL := apiRequest.URL
	shortURLDomain, err := c.createShortURLUseCase.CreateShortURL(r.Context(), longURL)
	if err != nil && !errors.Is(err, usecase.ErrConflict) {
		apiResponse.ErrorStatus = fmt.Sprintf("%d", http.StatusInternalServerError)
		apiResponse.ErrorDescription = fmt.Sprintf("Error when saving short URL: %s", err.Error())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(apiResponse)

		return
	}

	apiResponse.Result = util.GetShortURL(c.baseURL, shortURLDomain.ShortURI)

	w.Header().Set("Content-Type", "application/json")
	if err != nil && errors.Is(err, usecase.ErrConflict) {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	json.NewEncoder(w).Encode(apiResponse)
}

// CreateShortURLBatchHandler обрабатывает запрос на создание коротких ссылок пачкой
//
//	@Summary	Создание коротких ссылок пачкой
//	@Accepts	json
//	@Produce	json
//	@Success	200	{object}	dto.APICreateShortURLBatchResponse
//	@Failure	400	{string}	string	"ошибка в формате запроса"
//	@Failure	500	{string}	string	"внутренняя ошибка сервиса"
//	@Router		/api/shorten/batch [post]
//	@Param		body	body	dto.APICreateShortURLBatchRequest	true "запрос на создание коротких ссылок пачкой"
func (c *APIController) CreateShortURLBatchHandler(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var apiRequest dto.APICreateShortURLBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&apiRequest); err != nil {
		log.Errorw("app: error when decode request body from json", "err", err)

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	createShortURLBatchDomains := make([]domain.CreateShortURLBatchDomain, 0, len(apiRequest))
	for _, apiShortURLEntry := range apiRequest {
		shortURLBatchDomain := domain.CreateShortURLBatchDomain{
			CorrelationUUID: apiShortURLEntry.CorrelationID,
			LongURL:         apiShortURLEntry.OriginalURL,
		}
		createShortURLBatchDomains = append(createShortURLBatchDomains, shortURLBatchDomain)
	}

	createShortURLBatchResultDomains, err := c.createShortURLUseCase.CreateShortURLBatch(r.Context(), createShortURLBatchDomains)
	if err != nil {
		log.Errorw("app: error when store batch of URLs", "err", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	apiResponse := make(dto.APICreateShortURLBatchResponse, 0, len(apiRequest))
	for _, createShortURLBatchResultDomain := range createShortURLBatchResultDomains {
		apiResponseEntry := dto.APICreateShortURLBatchResponseEntry{
			CorrelationID: createShortURLBatchResultDomain.CorrelationUUID,
			ShortURL:      util.GetShortURL(c.baseURL, createShortURLBatchResultDomain.ShortURI),
		}
		apiResponse = append(apiResponse, apiResponseEntry)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(apiResponse)
}

// GetShortURLByUserID обрабатывает запрос на получение коротких ссылок созданных пользователем
//
//	@Summary	Получение коротких ссылок созданных пользователем
//	@Accepts	json
//	@Produce	json
//	@Success	200	{object}	dto.APIGetAllURLByUserIDResponse
//	@Success	204	{string}	string	"у пользователя нет коротких ссылок"
//	@Failure	400	{string}	string	"ошибка в формате запроса"
//	@Failure	500	{string}	string	"внутренняя ошибка сервиса"
//	@Router		/api/user/urls [get]
func (c *APIController) GetShortURLByUserID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDContextKey).(string)
	shortURLDomains, err := c.getShortURLUseCase.GetShortURLsByUserID(r.Context(), userID)
	if err != nil {
		log.Errorw("app: error when get shortURL by userID", "userID", userID, "err", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(shortURLDomains) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	apiResponse := make(dto.APIGetAllURLByUserIDResponse, 0, len(shortURLDomains))
	for _, storageEntry := range shortURLDomains {
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

// DeleteShortURLs обрабатывает запрос на удаление коротких ссылок
//
//	@Summary	Удаление коротких ссылок
//	@Accepts	json
//	@Produce	json
//	@Success	200	{string}	string
//	@Failure	400	{string}	string	"ошибка в формате запроса"
//	@Failure	500	{string}	string	"внутренняя ошибка сервиса"
//	@Router		/api/user/urls [delete]
//	@Param		body	body	[]string	true "список идентификаторов коротких ссылок"
func (c *APIController) DeleteShortURLs(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var apiRequest []string
	if err := json.NewDecoder(r.Body).Decode(&apiRequest); err != nil {
		log.Errorw("app: error when decode request body from json", "err", err)

		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := c.deleteShortURLUseCase.DeleteShortURLsByShortURIs(context.WithoutCancel(r.Context()), apiRequest)
	if err != nil {
		log.Errorw("app: error when delete by shortURIs", "err", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
