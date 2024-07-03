package app

import (
	"fmt"
	"github.com/vkhrushchev/urlshortener/internal/util"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

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
	a.router.Post("/", a.createShortURLHandler)
	a.router.Get("/{id}", a.getURLHandler)
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
	rawBody := make([]byte, r.ContentLength)
	_, err := r.Body.Read(rawBody)
	if err != nil && err != io.EOF {
		_ = fmt.Errorf("app: error reading requestBody: %v", err)

		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("app: error reading requestBody"))

		return
	}

	urlID := util.RandStringRunes(10)
	for a.urls[urlID] != "" {
		urlID = util.RandStringRunes(10)
	}

	a.urls[urlID] = string(rawBody)

	w.Header().Add("Content-Type", "plain/text")
	w.WriteHeader(http.StatusCreated)

	var shortUrl string
	if strings.HasSuffix(a.baseURL, "/") {
		shortUrl = a.baseURL + urlID
	} else {
		shortUrl = a.baseURL + "/" + urlID
	}

	_, _ = w.Write([]byte(shortUrl))

	return
}

func (a *URLShortenerApp) getURLHandler(w http.ResponseWriter, r *http.Request) {
	urlID := chi.URLParam(r, "id")
	url := a.urls[urlID]

	if url != "" {
		w.Header().Add("Content-Type", "plain/text")
		w.Header().Add("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)

		return
	}

	w.WriteHeader(http.StatusNotFound)
}
