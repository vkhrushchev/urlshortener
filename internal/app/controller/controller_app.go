package controller

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/vkhrushchev/urlshortener/internal/app/usecase"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/vkhrushchev/urlshortener/internal/util"
)

type AppController struct {
	baseURL               string
	createShortURLUseCase usecase.ICreateShortURLUseCase
	getShortURLUseCase    usecase.IGetShortURLUseCase
}

func NewAppController(
	baseURL string,
	createShortURLUseCase usecase.ICreateShortURLUseCase,
	getShortURLUseCase usecase.IGetShortURLUseCase) *AppController {
	return &AppController{
		baseURL:               baseURL,
		createShortURLUseCase: createShortURLUseCase,
		getShortURLUseCase:    getShortURLUseCase,
	}
}

func (c *AppController) CreateShortURLHandler(w http.ResponseWriter, r *http.Request) {
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
	shortURLDomain, err := c.createShortURLUseCase.CreateShortURL(r.Context(), longURL)
	if err != nil && !errors.Is(err, usecase.ErrConflict) {
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
		"shortURI", shortURLDomain.ShortURI,
		"longURL", shortURLDomain.LongURL,
	)

	w.Header().Add("Content-Type", "plain/text")
	if err != nil && errors.Is(err, usecase.ErrConflict) {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	shortURL := util.GetShortURL(c.baseURL, shortURLDomain.ShortURI)

	_, err = w.Write([]byte(shortURL))
	if err != nil {
		err = fmt.Errorf("app: error writing response: %v", err)
		log.Errorw(err.Error())
	}
}

func (c *AppController) GetURLHandler(w http.ResponseWriter, r *http.Request) {
	shortURI := chi.URLParam(r, "id")

	shortURLEntry, err := c.getShortURLUseCase.GetShortURL(r.Context(), shortURI)
	if err != nil && !errors.Is(err, usecase.ErrNotFound) {
		log.Errorw(
			"app: error when get original url from storage",
			"error", err.Error(),
		)

		w.Header().Add("Content-Type", "plain/text")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err != nil && errors.Is(err, usecase.ErrNotFound) {
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
