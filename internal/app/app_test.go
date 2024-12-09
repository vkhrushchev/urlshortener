package app

import (
	"context"
	"encoding/json"
	"github.com/vkhrushchev/urlshortener/internal/app/repository"
	"github.com/vkhrushchev/urlshortener/internal/app/use_case"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vkhrushchev/urlshortener/internal/app/controller"
	"github.com/vkhrushchev/urlshortener/internal/app/dto"
	"github.com/vkhrushchev/urlshortener/internal/middleware"
)

func TestURLShortenerApp_createShortURLHandler(t *testing.T) {
	shortUrlRepo := repository.NewInMemoryShortURLRepository()

	createShortURLUseCase := use_case.NewCreateShortURLUseCase(shortUrlRepo)
	getShortURLUseCase := use_case.NewGetShortURLUseCase(shortUrlRepo)
	deleteShortURLUseCase := use_case.NewDeleteShortURLUseCase(shortUrlRepo)

	appController := controller.NewAppController("", createShortURLUseCase, getShortURLUseCase)
	apiController := controller.NewAPIController("", createShortURLUseCase, getShortURLUseCase, deleteShortURLUseCase)
	// TODO mock healthController
	healthController := controller.NewHealthController(nil)

	app := NewURLShortenerApp("", appController, apiController, healthController)
	app.RegisterHandlers()

	ts := httptest.NewServer(app.router)
	defer ts.Close()

	tests := []struct {
		name        string
		requestBody string
		status      int
	}{
		{
			name:        "create success",
			requestBody: "https://google.com",
			status:      http.StatusCreated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, _, responseBody := executeRequest(
				t,
				ts,
				http.MethodPost,
				"/",
				tt.requestBody,
				"plain/text",
			)

			assert.Equal(t, tt.status, statusCode)
			if statusCode == http.StatusCreated {
				assert.NotEmpty(t, responseBody)
			}
		})
	}
}

func TestURLShortenerApp_getURLHandler(t *testing.T) {
	shortUrlRepo := repository.NewInMemoryShortURLRepository()

	createShortURLUseCase := use_case.NewCreateShortURLUseCase(shortUrlRepo)
	getShortURLUseCase := use_case.NewGetShortURLUseCase(shortUrlRepo)
	deleteShortURLUseCase := use_case.NewDeleteShortURLUseCase(shortUrlRepo)

	appController := controller.NewAppController("", createShortURLUseCase, getShortURLUseCase)
	apiController := controller.NewAPIController("", createShortURLUseCase, getShortURLUseCase, deleteShortURLUseCase)
	// TODO mock healthController
	healthController := controller.NewHealthController(nil)

	app := NewURLShortenerApp("", appController, apiController, healthController)
	app.RegisterHandlers()

	// добавляем подготовленные данные для тестов
	shortURLEntry, err := createShortURLUseCase.CreateShortURL(
		context.WithValue(context.Background(), middleware.UserIDContextKey, uuid.NewString()),
		"https://google.com",
	)
	require.NoError(t, err, "unexpected error when save URL")

	ts := httptest.NewServer(app.router)
	defer ts.Close()

	tests := []struct {
		name     string
		path     string
		status   int
		location string
	}{
		{
			name:     "get success",
			path:     "/" + shortURLEntry.ShortURI,
			status:   http.StatusTemporaryRedirect,
			location: "https://google.com",
		},
		{
			name:   "not found",
			path:   "/cba",
			status: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, headers, responseBody := executeRequest(
				t,
				ts,
				http.MethodGet,
				tt.path,
				"",
				"plain/text",
			)

			assert.Equal(t, tt.status, statusCode)
			assert.Empty(t, responseBody)
			if statusCode == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.location, headers.Get("Location"))
			}
		})
	}
}

func TestURLShortenerApp_createShortURLHandlerAPI(t *testing.T) {
	shortUrlRepo := repository.NewInMemoryShortURLRepository()

	createShortURLUseCase := use_case.NewCreateShortURLUseCase(shortUrlRepo)
	getShortURLUseCase := use_case.NewGetShortURLUseCase(shortUrlRepo)
	deleteShortURLUseCase := use_case.NewDeleteShortURLUseCase(shortUrlRepo)

	appController := controller.NewAppController("", createShortURLUseCase, getShortURLUseCase)
	apiController := controller.NewAPIController("", createShortURLUseCase, getShortURLUseCase, deleteShortURLUseCase)
	// TODO mock healthController
	healthController := controller.NewHealthController(nil)

	app := NewURLShortenerApp("", appController, apiController, healthController)
	app.RegisterHandlers()

	ts := httptest.NewServer(app.router)
	defer ts.Close()

	testCases := []struct {
		name               string
		contentType        string
		apiRequestRaw      string
		apiRequest         *dto.APICreateShortURLRequest
		expectedStatusCode int
	}{
		{
			name:        "success",
			contentType: "application/json",
			apiRequest: &dto.APICreateShortURLRequest{
				URL: "https://google.com",
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:        "wrong content type",
			contentType: "plain/text",
			apiRequest: &dto.APICreateShortURLRequest{
				URL: "https://google.com",
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "bad json",
			contentType:        "application/json",
			apiRequestRaw:      "{",
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var requestBodyBytes []byte
			var err error
			if tc.apiRequest != nil {
				requestBodyBytes, err = json.Marshal(tc.apiRequest)
				require.NoError(t, err, "app_test: error when marshall dto.ApiCreateShortURLRequest: %v", err)
			} else if tc.apiRequestRaw != "" {
				requestBodyBytes = []byte(tc.apiRequestRaw)
			}

			statusCode, headers, responseBody := executeRequest(
				t,
				ts,
				http.MethodPost,
				"/api/shorten",
				string(requestBodyBytes),
				tc.contentType,
			)

			assert.Equal(t, tc.expectedStatusCode, statusCode)
			assert.Equal(t, "application/json", headers.Get("Content-Type"))

			var apiResponse dto.APICreateShortURLResponse
			err = json.Unmarshal([]byte(responseBody), &apiResponse)
			if err != nil {
				require.NoError(t, err, "app_test: error when unmarshall dto.APICreateShortURLResponse: %v", err)
			}

			if statusCode == http.StatusCreated {
				assert.Empty(t, apiResponse.ErrorStatus)
				assert.Empty(t, apiResponse.ErrorDescription)
				assert.NotEmpty(t, apiResponse.Result)
			}

			if statusCode == http.StatusBadRequest {
				assert.NotEmpty(t, apiResponse.ErrorStatus)
				assert.NotEmpty(t, apiResponse.ErrorDescription)
				assert.Empty(t, apiResponse.Result)
			}
		})
	}
}

func TestURLShortenerApp_createShortURLBatchHandlerAPI(t *testing.T) {
	shortUrlRepo := repository.NewInMemoryShortURLRepository()

	createShortURLUseCase := use_case.NewCreateShortURLUseCase(shortUrlRepo)
	getShortURLUseCase := use_case.NewGetShortURLUseCase(shortUrlRepo)
	deleteShortURLUseCase := use_case.NewDeleteShortURLUseCase(shortUrlRepo)

	appController := controller.NewAppController("", createShortURLUseCase, getShortURLUseCase)
	apiController := controller.NewAPIController("", createShortURLUseCase, getShortURLUseCase, deleteShortURLUseCase)
	// TODO mock healthController
	healthController := controller.NewHealthController(nil)

	app := NewURLShortenerApp("", appController, apiController, healthController)
	app.RegisterHandlers()

	ts := httptest.NewServer(app.router)
	defer ts.Close()

	testCases := []struct {
		name               string
		contentType        string
		apiRequestRaw      string
		apiRequest         dto.APICreateShortURLBatchRequest
		expectedStatusCode int
	}{
		{
			name:        "success",
			contentType: "application/json",
			apiRequest: dto.APICreateShortURLBatchRequest{
				dto.APICreateShortURLBatchRequestEntry{
					CorrelationID: "96f178e8-e1ae-4744-9501-69da1fba5def",
					OriginalURL:   "http://www.google.com",
				},
				dto.APICreateShortURLBatchRequestEntry{
					CorrelationID: "3f6e4e67-a5ba-4d6c-b76a-cd56d30499d9",
					OriginalURL:   "http://ya.ru",
				},
			},
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:        "wrong content type",
			contentType: "plain/text",
			apiRequest: dto.APICreateShortURLBatchRequest{
				dto.APICreateShortURLBatchRequestEntry{
					CorrelationID: "96f178e8-e1ae-4744-9501-69da1fba5def",
					OriginalURL:   "http://www.google.com",
				},
				dto.APICreateShortURLBatchRequestEntry{
					CorrelationID: "3f6e4e67-a5ba-4d6c-b76a-cd56d30499d9",
					OriginalURL:   "http://ya.ru",
				},
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "bad json",
			contentType:        "application/json",
			apiRequestRaw:      "{",
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var requestBodyBytes []byte
			var err error
			if tc.apiRequest != nil {
				requestBodyBytes, err = json.Marshal(tc.apiRequest)
				require.NoError(t, err, "app_test: error when marshall dto.APICreateShortURLBatchRequest: %v", err)
			} else if tc.apiRequestRaw != "" {
				requestBodyBytes = []byte(tc.apiRequestRaw)
			}

			statusCode, headers, responseBody := executeRequest(
				t,
				ts,
				http.MethodPost,
				"/api/shorten/batch",
				string(requestBodyBytes),
				tc.contentType,
			)

			assert.Equal(t, tc.expectedStatusCode, statusCode)

			if statusCode == http.StatusCreated {
				var apiResponse dto.APICreateShortURLBatchResponse
				err = json.Unmarshal([]byte(responseBody), &apiResponse)
				if err != nil {
					require.NoError(t, err, "app_test: error when unmarshall dto.APICreateShortURLBatchResponse: %v", err)
				}

				assert.Equal(t, "application/json", headers.Get("Content-Type"))
				assert.Equal(t, 2, len(apiResponse))
			}
		})
	}
}

func executeRequest(
	t *testing.T,
	ts *httptest.Server,
	method string,
	path string,
	requestBody string,
	contentType string,
) (int, http.Header, string) {
	request, err := http.NewRequest(method, ts.URL+path, strings.NewReader(requestBody))
	require.NoError(t, err)

	request.Header.Add("Content-Type", contentType)

	// отключаем редирект
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	response, err := ts.Client().Do(request)
	require.NoError(t, err)

	responseBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	defer response.Body.Close()

	return response.StatusCode, response.Header, string(responseBody)
}
