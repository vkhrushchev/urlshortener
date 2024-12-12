package controller

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/vkhrushchev/urlshortener/internal/app/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/vkhrushchev/urlshortener/internal/util"
)

// AppController используется для обработки не API-запросов приложения
type AppController struct {
	baseURL               string                         // URL до сервера с развернутым приложением
	createShortURLUseCase usecase.ICreateShortURLUseCase // Сценарий создания короткой ссылки
	getShortURLUseCase    usecase.IGetShortURLUseCase    // Сценарий получения короткой ссылки
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
	requestBodyBytes, err := io.ReadAll(r.Body)
	if err != nil && !errors.Is(err, io.EOF) {
		log.Errorw("app: error reading request body", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	longURL := strings.TrimSpace(string(requestBodyBytes))
	shortURLDomain, err := c.createShortURLUseCase.CreateShortURL(r.Context(), longURL)
	if err != nil && !errors.Is(err, usecase.ErrConflict) {
		log.Errorw("app: error when create shortURL", "err", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "plain/text")
	if err != nil && errors.Is(err, usecase.ErrConflict) {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	shortURL := util.GetShortURL(c.baseURL, shortURLDomain.ShortURI)

	if _, err = w.Write([]byte(shortURL)); err != nil {
		log.Errorw("app: error writing response", "err", err.Error())
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
		log.Errorw("app: error when get original url", "err", err)
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
