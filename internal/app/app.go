package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/vkhrushchev/urlshortener/internal/app/controller"
	"github.com/vkhrushchev/urlshortener/internal/app/dto"
	"github.com/vkhrushchev/urlshortener/internal/app/storage"
	"github.com/vkhrushchev/urlshortener/internal/middleware"

	"github.com/go-chi/chi/v5"

	"go.uber.org/zap"
)

var log = zap.Must(zap.NewProduction()).Sugar()

type URLShortenerApp struct {
	storage          storage.Storage
	healthController controller.HealthController
	router           chi.Router
	runAddr          string
	baseURL          string
}

func NewURLShortenerApp(
	runAddr string,
	baseURL string,
	storage storage.Storage,
	healthController controller.HealthController) *URLShortenerApp {
	return &URLShortenerApp{
		storage:          storage,
		healthController: healthController,
		router:           chi.NewRouter(),
		runAddr:          runAddr,
		baseURL:          baseURL,
	}
}

func (a *URLShortenerApp) RegisterHandlers() {
	a.router.Post("/", middleware.LogRequestMiddleware(middleware.GzipMiddleware(a.createShortURLHandler)))
	a.router.Get("/{id}", middleware.LogRequestMiddleware(middleware.GzipMiddleware(a.getURLHandler)))
	a.router.Post("/api/shorten", middleware.LogRequestMiddleware(middleware.GzipMiddleware(a.createShortURLHandlerAPI)))
	a.router.Get("/ping", a.healthController.Ping)
}

func (a *URLShortenerApp) Run() error {
	log.Infow(
		"app: URLShortenerApp stating",
		"runAddr", a.runAddr,
		"baseURL", a.baseURL,
	)

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
			log.Errorw(err.Error())
		}

		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte("app: error reading requestBody"))
		if err != nil {
			log.Errorw(err.Error())
		}

		return
	}

	longURL := strings.TrimSpace(bodyBuffer.String())
	shortURI, err := a.storage.SaveURL(r.Context(), longURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorw(err.Error())
		_, err = w.Write([]byte(err.Error()))
		if err != nil {
			log.Errorw(err.Error())
		}

		return
	}

	log.Infow(
		"app: short_url_added",
		"shortURI", shortURI,
		"longURL", longURL,
	)

	w.Header().Add("Content-Type", "plain/text")
	w.WriteHeader(http.StatusCreated)

	shortURL := a.getShortURL(shortURI)

	_, err = w.Write([]byte(shortURL))
	if err != nil {
		err = fmt.Errorf("app: error writing response: %v", err)
		log.Errorw(err.Error())
	}
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
	shortURI := chi.URLParam(r, "id")

	longURL, found, err := a.storage.GetURLByShortURI(r.Context(), shortURI)
	if err != nil {
		log.Errorw(
			"app: error when get original url from storage",
			"error", err.Error(),
		)

		w.Header().Add("Content-Type", "plain/text")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !found {
		w.Header().Add("Content-Type", "plain/text")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "plain/text")
	w.Header().Add("Location", strings.TrimSpace(longURL))
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
	shortURI, err := a.storage.SaveURL(r.Context(), longURL)
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

	apiResponse.Result = a.getShortURL(shortURI)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(apiResponse)
}
