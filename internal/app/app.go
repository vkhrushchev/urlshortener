package app

import (
	"fmt"
	"github.com/vkhrushchev/urlshortener/internal/util"
	"io"
	"net/http"
)

type URLShortenerApp struct {
	urls map[string]string
}

func NewURLShortenerApp() *URLShortenerApp {
	return &URLShortenerApp{
		urls: make(map[string]string),
	}
}

func (a *URLShortenerApp) Run() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", a.handleRequest)

	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		return err
	}

	return nil
}

func (a *URLShortenerApp) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		rawBody := make([]byte, r.ContentLength)
		_, err := r.Body.Read(rawBody)
		if err != nil && err != io.EOF {
			_ = fmt.Errorf("app: error reading body: %v", err)

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("app: error reading body"))

			return
		}

		urlID := util.RandStringRunes(10)
		for a.urls[urlID] != "" {
			urlID = util.RandStringRunes(10)
		}

		a.urls[urlID] = string(rawBody)

		w.Header().Add("Content-Type", "plain/text")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("http://localhost:8080/" + urlID))

		return
	}

	if r.Method == http.MethodGet {
		urlID := r.RequestURI[1:len(r.RequestURI)]

		w.Header().Add("Content-Type", "plain/text")
		w.Header().Add("Location", a.urls[urlID])
		w.WriteHeader(http.StatusTemporaryRedirect)

		return
	}

	w.WriteHeader(http.StatusBadRequest)
}
