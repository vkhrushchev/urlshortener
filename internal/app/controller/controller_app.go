package controller

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/vkhrushchev/urlshortener/internal/app/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/vkhrushchev/urlshortener/internal/util"
)

// AppController используется для обработки не API-запросов приложения
type AppController struct {
	createShortURLUseCase usecase.ICreateShortURLUseCase // Сценарий создания короткой ссылки
	getShortURLUseCase    usecase.IGetShortURLUseCase    // Сценарий получения короткой ссылки
	baseURL               string                         // URL до сервера с развернутым приложением
}

// NewAppController создает новый экземпляр структуры AppController
//
//	baseURL - URL до сервера с развернутым приложением
//	createShortURLUseCase - use case создания короткой ссылки
//	getShortURLUseCase - use case получения короткой ссылки
func NewAppController(
	baseURL string,
	createShortURLUseCase usecase.ICreateShortURLUseCase,
	getShortURLUseCase usecase.IGetShortURLUseCase,
) *AppController {
	return &AppController{
		baseURL:               baseURL,
		createShortURLUseCase: createShortURLUseCase,
		getShortURLUseCase:    getShortURLUseCase,
	}
}

// CreateShortURLHandler обрабатывает запрос на создание короткой ссылки
//
//	@Summary	Создание короткой ссылки
//	@Accepts	plain
//	@Produce	plain
//	@Success	201	{string}	string	""
//	@Success	409	{string}	string	"короткая ссылка уже существует"
//	@Failure	500	{string}	string	"внутренняя ошибка сервиса"
//	@Router		/ [post]
//	@Param		body	body	string	true	"ссылка которую требуется сократить"
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

// GetURLHandler возвращает полную ссылку по короткой ссылке
//
//	@Summary	получить короткую ссылку
//	@Accepts	plain
//	@Produce	plain
//	@Success	307	{string}	string
//	@Failure	404	{string}	string	"короткая ссылка не найдена"
//	@Failure	410	{string}	string	"короткая ссылка удалена"
//	@Failure	500	{string}	string	"внутренняя ошибка сервиса"
//	@Router		/{shortURI} [get]
//	@Param		shortURI	path	string	true	"идентификатор короткой ссылки"
func (c *AppController) GetURLHandler(w http.ResponseWriter, r *http.Request) {
	shortURI := chi.URLParam(r, "id")

	shortURLEntry, err := c.getShortURLUseCase.GetShortURLByShortURI(r.Context(), shortURI)
	if err != nil && !errors.Is(err, usecase.ErrNotFound) {
		log.Errorw("app: error when get original url from storage", "err", err)

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
