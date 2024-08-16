package app

import (
	"net/http"

	"github.com/vkhrushchev/urlshortener/internal/app/controller"
	"github.com/vkhrushchev/urlshortener/internal/middleware"

	"github.com/go-chi/chi/v5"

	"go.uber.org/zap"
)

var log = zap.Must(zap.NewProduction()).Sugar()

type URLShortenerApp struct {
	appController    *controller.AppContoller
	apiController    *controller.APIController
	healthController *controller.HealthController
	router           chi.Router
	runAddr          string
}

func NewURLShortenerApp(
	runAddr string,
	appController *controller.AppContoller,
	apiController *controller.APIController,
	healthController *controller.HealthController) *URLShortenerApp {
	return &URLShortenerApp{
		appController:    appController,
		apiController:    apiController,
		healthController: healthController,
		router:           chi.NewRouter(),
		runAddr:          runAddr,
	}
}

func (a *URLShortenerApp) RegisterHandlers() {
	a.router.Post(
		"/",
		middleware.LogRequestMiddleware(middleware.GzipMiddleware(a.appController.CreateShortURLHandler)))
	a.router.Get(
		"/{id}",
		middleware.LogRequestMiddleware(middleware.GzipMiddleware(a.appController.GetURLHandler)))
	a.router.Post(
		"/api/shorten",
		middleware.LogRequestMiddleware(middleware.GzipMiddleware(a.apiController.CreateShortURLHandler)))
	a.router.Post(
		"/api/shorten/batch",
		middleware.LogRequestMiddleware(middleware.GzipMiddleware(a.apiController.CreateShortURLBatchHandler)))
	a.router.Get(
		"/ping",
		a.healthController.Ping)
}

func (a *URLShortenerApp) Run() error {
	log.Infow(
		"app: URLShortenerApp stating",
		"runAddr", a.runAddr,
	)

	err := http.ListenAndServe(a.runAddr, a.router)
	if err != nil {
		return err
	}

	return nil
}
