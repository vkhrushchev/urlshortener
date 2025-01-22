package app

import (
	"context"
	"errors"
	"golang.org/x/crypto/acme/autocert"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/vkhrushchev/urlshortener/internal/app/controller"
	"github.com/vkhrushchev/urlshortener/internal/middleware"

	"github.com/go-chi/chi/v5"

	"go.uber.org/zap"
)

var log = zap.Must(zap.NewDevelopment()).Sugar()

// URLShortenerApp - структура с описанием приложения Shortener
type URLShortenerApp struct {
	appController      *controller.AppController
	apiController      *controller.APIController
	healthController   *controller.HealthController
	internalController *controller.InternalController
	router             chi.Router
	runAddr            string
	enableHTTPS        bool
	trustedSubnet      *net.IPNet
}

// NewURLShortenerApp создает экземпляр структуры URLShortenerApp
func NewURLShortenerApp(
	runAddr string,
	enableHTTPS bool,
	trustedSubnet *net.IPNet,
	appController *controller.AppController,
	apiController *controller.APIController,
	healthController *controller.HealthController,
	internalController *controller.InternalController,
) *URLShortenerApp {
	return &URLShortenerApp{
		appController:      appController,
		apiController:      apiController,
		healthController:   healthController,
		internalController: internalController,
		router:             chi.NewRouter(),
		runAddr:            runAddr,
		enableHTTPS:        enableHTTPS,
		trustedSubnet:      trustedSubnet,
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
	a.router.Get(
		"/api/internal/stats",
		middleware.CheckSubnetMiddleware(
			a.trustedSubnet,
			a.internalController.GetStats))
}

// Run запускает http-сервер с приложением
func (a *URLShortenerApp) Run() {
	log.Infow("app: URLShortenerApp stated", "runAddr", a.runAddr)

	server := &http.Server{
		Addr:    a.runAddr,
		Handler: a.router,
	}

	gracefulShutdownChan := make(chan struct{})
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-signalChan

		if err := server.Shutdown(context.Background()); err != nil {
			log.Errorw("app: failed to shutdown server", "error", err)
		}

		close(gracefulShutdownChan)
	}()

	if a.enableHTTPS {
		manager := &autocert.Manager{
			Cache:      autocert.DirCache("cache-dir"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("localhost"),
		}

		server.TLSConfig = manager.TLSConfig()

		if err := server.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("app: failed to listen and serve https server: %v", err)
		}
	} else {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("app: failed to listen and serve http server: %v", err)
		}
	}

	<-gracefulShutdownChan
	log.Infow("app: URLShortenerApp shutting down")
}
