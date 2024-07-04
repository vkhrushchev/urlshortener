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

	urlID := util.RandStringRunes(10)
	for a.urls[urlID] != "" {
		urlID = util.RandStringRunes(10)
	}

	a.urls[urlID] = string(rawBody)

	w.Header().Add("Content-Type", "plain/text")
	w.WriteHeader(http.StatusCreated)

	var shortURL string
	if strings.HasSuffix(a.baseURL, "/") {
		shortURL = a.baseURL + urlID
	} else {
		shortURL = a.baseURL + "/" + urlID
	}

	_, err = w.Write([]byte(shortURL))
	err = fmt.Errorf("app: error writing response: %v", err)
	if err != nil {
		println(err.Error())
	}
}

func (a *URLShortenerApp) getURLHandler(w http.ResponseWriter, r *http.Request) {
	urlID := chi.URLParam(r, "id")

	url, found := a.urls[urlID]
	if !found {
		w.WriteHeader(http.StatusNotFound)
	}

	w.Header().Add("Content-Type", "plain/text")
	w.Header().Add("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
