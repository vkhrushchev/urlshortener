package app

import (
	"golang.org/x/crypto/acme/autocert"
	"net/http"

	"github.com/vkhrushchev/urlshortener/internal/app/controller"
	"github.com/vkhrushchev/urlshortener/internal/middleware"

	"github.com/go-chi/chi/v5"

	"go.uber.org/zap"
)

var log = zap.Must(zap.NewDevelopment()).Sugar()

// URLShortenerApp - структура с описанием приложения Shortener
type URLShortenerApp struct {
	appController    *controller.AppController
	apiController    *controller.APIController
	healthController *controller.HealthController
	router           chi.Router
	runAddr          string
	enableHTTPS      bool
}

// NewURLShortenerApp создает экземпляр структуры URLShortenerApp
func NewURLShortenerApp(
	runAddr string,
	enableHTTPS bool,
	appController *controller.AppController,
	apiController *controller.APIController,
	healthController *controller.HealthController,
) *URLShortenerApp {
	return &URLShortenerApp{
		appController:    appController,
		apiController:    apiController,
		healthController: healthController,
		router:           chi.NewRouter(),
		runAddr:          runAddr,
		enableHTTPS:      enableHTTPS,
	}
}

// RegisterHandlers регистрирует обработчики http-запросов
func (a *URLShortenerApp) RegisterHandlers() {
	a.router.Post(
		"/",
		middleware.LogRequestMiddleware(
			middleware.UserIDCookieMiddleware(
				middleware.GzipMiddleware(a.appController.CreateShortURLHandler))))
	a.router.Get(
		"/{id}",
		middleware.LogRequestMiddleware(middleware.GzipMiddleware(a.appController.GetURLHandler)))
	a.router.Post(
		"/api/shorten",
		middleware.LogRequestMiddleware(
			middleware.UserIDCookieMiddleware(
				middleware.GzipMiddleware(a.apiController.CreateShortURLHandler))))
	a.router.Post(
		"/api/shorten/batch",
		middleware.LogRequestMiddleware(
			middleware.UserIDCookieMiddleware(
				middleware.GzipMiddleware(a.apiController.CreateShortURLBatchHandler))))
	a.router.Get(
		"/api/user/urls",
		middleware.LogRequestMiddleware(
			middleware.AuthByUserIDCookieMiddleware(
				middleware.GzipMiddleware(a.apiController.GetShortURLByUserID))))
	a.router.Delete(
		"/api/user/urls",
		middleware.LogRequestMiddleware(
			middleware.AuthByUserIDCookieMiddleware(
				middleware.GzipMiddleware(a.apiController.DeleteShortURLs))))
	a.router.Get(
		"/ping",
		a.healthController.Ping)
}

// Run запускает http-сервер с приложением
func (a *URLShortenerApp) Run() error {
	log.Infow(
		"app: URLShortenerApp stated",
		"runAddr", a.runAddr,
	)

	if a.enableHTTPS {
		manager := &autocert.Manager{
			Cache:      autocert.DirCache("cache-dir"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("localhost"),
		}

		server := &http.Server{
			Addr:      a.runAddr,
			Handler:   a.router,
			TLSConfig: manager.TLSConfig(),
		}

		if err := server.ListenAndServeTLS("", ""); err != nil {
			return err
		}
	} else {
		if err := http.ListenAndServe(a.runAddr, a.router); err != nil {
			return err
		}
	}

	return nil
}
