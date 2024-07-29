package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/vkhrushchev/urlshortener/internal/app/dto"
	"github.com/vkhrushchev/urlshortener/internal/middleware"
	"github.com/vkhrushchev/urlshortener/internal/util"

	"github.com/go-chi/chi/v5"

	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func init() {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic("cannot initilize zap")
	}

	log = zapLogger.Sugar()
}

type URLShortenerApp struct {
	urls    map[string]string
	router  chi.Router
	runAddr string
	baseURL string
}

func NewURLShortenerApp(runAddr string, baseURL string) *URLShortenerApp {
	return &URLShortenerApp{
		urls:    make(map[string]string),
		router:  chi.NewRouter(),
		runAddr: runAddr,
		baseURL: baseURL,
	}
}

func (a *URLShortenerApp) RegisterHandlers() {
	a.router.Post("/", middleware.LogRequestMiddleware(middleware.GzipMiddleware(a.createShortURLHandler)))
	a.router.Get("/{id}", middleware.LogRequestMiddleware(middleware.GzipMiddleware(a.getURLHandler)))
	a.router.Post("/api/shorten", middleware.LogRequestMiddleware(middleware.GzipMiddleware(a.createShortURLHandlerAPI)))
}

func (a *URLShortenerApp) Run() error {
	fmt.Printf("Listening on %s\n", a.runAddr)
	fmt.Printf("BaseURL: %s\n", a.baseURL)

	err := http.ListenAndServe(a.runAddr, a.router)
	if err != nil {
		return err
	}

	return nil
}

func (a *URLShortenerApp) createShortURLHandler(w http.ResponseWriter, r *http.Request) {
	var bodyBuffer bytes.Buffer
	_, err := bodyBuffer.ReadFrom(r.Body)
	if err != nil && err != io.EOF {
		err = fmt.Errorf("app: error reading requestBody: %v", err)
		if err != nil {
			println(err.Error())
		}

		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte("app: error reading requestBody"))
		if err != nil {
			println(err.Error())
		}

		return
	}

	urlID := a.createShortURLID()
	a.urls[urlID] = strings.TrimSpace(bodyBuffer.String())
	log.Infow(
		"app: short_url_added",
		"urlID", urlID,
		"fullURL", a.urls[urlID],
	)

	w.Header().Add("Content-Type", "plain/text")
	w.WriteHeader(http.StatusCreated)

	shortURL := a.getShortURL(urlID)

	_, err = w.Write([]byte(shortURL))
	if err != nil {
		err = fmt.Errorf("app: error writing response: %v", err)
		println(err.Error())
	}
}

func (a *URLShortenerApp) createShortURLID() string {
	urlID := util.RandStringRunes(10)
	for a.urls[urlID] != "" {
		urlID = util.RandStringRunes(10)
	}

	return urlID
}

func (a *URLShortenerApp) getShortURL(urlID string) string {
	var shortURL string
	if strings.HasSuffix(a.baseURL, "/") {
		shortURL = a.baseURL + urlID
	} else {
		shortURL = a.baseURL + "/" + urlID
	}

	return shortURL
}

func (a *URLShortenerApp) getURLHandler(w http.ResponseWriter, r *http.Request) {
	urlID := chi.URLParam(r, "id")

	url, found := a.urls[urlID]
	if !found {
		w.Header().Add("Content-Type", "plain/text")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "plain/text")
	w.Header().Add("Location", strings.TrimSpace(url))
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (a *URLShortenerApp) createShortURLHandlerAPI(w http.ResponseWriter, r *http.Request) {
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
		log.Infow(
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

	urlID := a.createShortURLID()
	a.urls[urlID] = apiRequest.URL

	apiResponse.Result = a.getShortURL(urlID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(apiResponse)
}
