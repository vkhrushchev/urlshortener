package controller

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/vkhrushchev/urlshortener/internal/app/storage"
	"github.com/vkhrushchev/urlshortener/internal/util"
)

type AppContoller struct {
	baseURL string
	storage storage.Storage
}

func NewAppController(baseURL string, storage storage.Storage) *AppContoller {
	return &AppContoller{
		baseURL: baseURL,
		storage: storage,
	}
}

func (c *AppContoller) CreateShortURLHandler(w http.ResponseWriter, r *http.Request) {
	var bodyBuffer bytes.Buffer
	_, err := bodyBuffer.ReadFrom(r.Body)
	if err != nil && !errors.Is(err, io.EOF) {
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
	shortURLEntry, err := c.storage.SaveURL(r.Context(), longURL)
	if err != nil && !errors.Is(err, storage.ErrConflictOnUniqueConstraint) {
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
		"shortURI", shortURLEntry.ShortURI,
		"longURL", shortURLEntry.LongURL,
	)

	w.Header().Add("Content-Type", "plain/text")
	if err != nil && errors.Is(err, storage.ErrConflictOnUniqueConstraint) {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	shortURL := util.GetShortURL(c.baseURL, shortURLEntry.ShortURI)

	_, err = w.Write([]byte(shortURL))
	if err != nil {
		err = fmt.Errorf("app: error writing response: %v", err)
		log.Errorw(err.Error())
	}
}

func (c *AppContoller) GetURLHandler(w http.ResponseWriter, r *http.Request) {
	shortURI := chi.URLParam(r, "id")

	shortURLEntry, err := c.storage.GetURLByShortURI(r.Context(), shortURI)
	if err != nil {
		log.Errorw(
			"app: error when get original url from storage",
			"error", err.Error(),
		)

		w.Header().Add("Content-Type", "plain/text")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if shortURLEntry == nil {
		w.Header().Add("Content-Type", "plain/text")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if shortURLEntry.Deleted {
		w.Header().Add("Content-Type", "plain/text")
		w.WriteHeader(http.StatusGone)
		return
	}

	w.Header().Add("Content-Type", "plain/text")
	w.Header().Add("Location", strings.TrimSpace(shortURLEntry.LongURL))
	w.WriteHeader(http.StatusTemporaryRedirect)
}
