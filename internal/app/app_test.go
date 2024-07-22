package app

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestURLShortenerApp_createShortURLHandler(t *testing.T) {
	app := NewURLShortenerApp("", "")
	app.RegisterHandlers()

	ts := httptest.NewServer(app.router)
	defer ts.Close()

	tests := []struct {
		name        string
		requestBody string
		status      int
	}{
		{
			name:        "get success",
			requestBody: "https://google.com",
			status:      http.StatusCreated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, _, responseBody := executeRequest(t, ts, http.MethodPost, "/", tt.requestBody)

			assert.Equal(t, tt.status, statusCode)
			if statusCode == http.StatusCreated {
				assert.NotEmpty(t, responseBody)
			}
		})
	}
}

func TestURLShortenerApp_getURLHandler(t *testing.T) {
	app := NewURLShortenerApp("", "")
	app.RegisterHandlers()

	// добавляем подготовленные данные для тестов
	app.urls["abc"] = "https://google.com"

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
			path:     "/abc",
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
			statusCode, headers, responseBody := executeRequest(t, ts, http.MethodGet, tt.path, "")

			assert.Equal(t, tt.status, statusCode)
			assert.Empty(t, responseBody)
			if statusCode == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.location, headers.Get("Location"))
			}
		})
	}
}

func executeRequest(t *testing.T, ts *httptest.Server, method string, path string, requestBody string) (int, http.Header, string) {
	request, err := http.NewRequest(method, ts.URL+path, strings.NewReader(requestBody))
	require.NoError(t, err)

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
