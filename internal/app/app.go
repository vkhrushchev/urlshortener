package app

import (
	"fmt"
	"github.com/vkhrushchev/urlshortener/internal/util"
	"io"
	"net/http"
)

type UrlShortenerApp struct {
	urls map[string]string
}

func NewUrlShortenerApp() *UrlShortenerApp {
	return &UrlShortenerApp{
		urls: make(map[string]string),
	}
}

func (a *UrlShortenerApp) Run() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			rawBody := make([]byte, r.ContentLength)
			_, err := r.Body.Read(rawBody)
			if err != nil && err != io.EOF {
				_ = fmt.Errorf("app: error reading body: %v", err)

				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("app: error reading body"))

				return
			}

			urlId := util.RandStringRunes(10)
			for a.urls[urlId] != "" {
				urlId = util.RandStringRunes(10)
			}

			a.urls[urlId] = string(rawBody)

			w.Header().Add("Content-Type", "plain/text")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("http://localhost:8080/" + urlId))

			return
		}

		if r.Method == http.MethodGet {
			urlId := r.RequestURI[1:len(r.RequestURI)]

			w.Header().Add("Content-Type", "plain/text")
			w.Header().Add("Location", a.urls[urlId])
			w.WriteHeader(http.StatusTemporaryRedirect)

			return
		}

		w.WriteHeader(http.StatusBadRequest)
	})

	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		return err
	}

	return nil
}
