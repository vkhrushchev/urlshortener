package app

import (
	"context"
	"errors"
	shortenergrpc "github.com/vkhrushchev/urlshortener/internal/app/grpc"
	"github.com/vkhrushchev/urlshortener/internal/interceptor"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/vkhrushchev/urlshortener/internal/app/controller"
	"github.com/vkhrushchev/urlshortener/internal/middleware"

	"github.com/go-chi/chi/v5"

	"go.uber.org/zap"

	pb "github.com/vkhrushchev/urlshortener/grpc"
)

var log = zap.Must(zap.NewDevelopment()).Sugar()

// URLShortenerApp - структура с описанием приложения Shortener
type URLShortenerApp struct {
	appController                  *controller.AppController
	apiController                  *controller.APIController
	healthController               *controller.HealthController
	internalController             *controller.InternalController
	grpcShortenerServiceServerImpl *shortenergrpc.ShortenerServiceServerImpl
	router                         chi.Router
	runAddr                        string
	enableHTTPS                    bool
	trustedSubnet                  *net.IPNet
	grpcAddr                       string
	salt                           string
}

// NewURLShortenerApp создает экземпляр структуры URLShortenerApp
func NewURLShortenerApp(
	runAddr string,
	enableHTTPS bool,
	trustedSubnet *net.IPNet,
	grpcAddr string,
	salt string,
	appController *controller.AppController,
	apiController *controller.APIController,
	healthController *controller.HealthController,
	internalController *controller.InternalController,
	grpcServer *shortenergrpc.ShortenerServiceServerImpl) *URLShortenerApp {
	return &URLShortenerApp{
		appController:                  appController,
		apiController:                  apiController,
		healthController:               healthController,
		internalController:             internalController,
		grpcShortenerServiceServerImpl: grpcServer,
		router:                         chi.NewRouter(),
		runAddr:                        runAddr,
		enableHTTPS:                    enableHTTPS,
		trustedSubnet:                  trustedSubnet,
		grpcAddr:                       grpcAddr,
		salt:                           salt,
	}
}

// RegisterHTTPHandlers регистрирует обработчики http-запросов
func (a *URLShortenerApp) RegisterHTTPHandlers() {
	a.router.Post(
		"/",
		middleware.LogRequestMiddleware(
			middleware.UserIDCookieMiddleware(
				a.salt,
				middleware.GzipMiddleware(a.appController.CreateShortURLHandler))))
	a.router.Get(
		"/{id}",
		middleware.LogRequestMiddleware(middleware.GzipMiddleware(a.appController.GetURLHandler)))
	a.router.Post(
		"/api/shorten",
		middleware.LogRequestMiddleware(
			middleware.UserIDCookieMiddleware(
				a.salt,
				middleware.GzipMiddleware(a.apiController.CreateShortURLHandler))))
	a.router.Post(
		"/api/shorten/batch",
		middleware.LogRequestMiddleware(
			middleware.UserIDCookieMiddleware(
				a.salt,
				middleware.GzipMiddleware(a.apiController.CreateShortURLBatchHandler))))
	a.router.Get(
		"/api/user/urls",
		middleware.LogRequestMiddleware(
			middleware.AuthByUserIDCookieMiddleware(
				a.salt,
				middleware.GzipMiddleware(a.apiController.GetShortURLByUserID))))
	a.router.Delete(
		"/api/user/urls",
		middleware.LogRequestMiddleware(
			middleware.AuthByUserIDCookieMiddleware(
				a.salt,
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

// RunHTTPServer запускает http-сервер с приложением
func (a *URLShortenerApp) RunHTTPServer(gracefulShutdownCh chan struct{}) {
	log.Infow("app: URLShortenerApp stated", "runAddr", a.runAddr)

	server := &http.Server{
		Addr:    a.runAddr,
		Handler: a.router,
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-signalChan

		if err := server.Shutdown(context.Background()); err != nil {
			log.Errorw("app: failed to shutdown server", "error", err)
		}

		close(gracefulShutdownCh)
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
}

// RunGRPCServer запускает grpc-сервер с приложением
func (a *URLShortenerApp) RunGRPCServer(gracefulShutdownCh chan struct{}) {
	listenTcpPort, err := net.Listen("tcp", "localhost:18080")
	if err != nil {
		log.Fatalw("app: failed to acquire TCP port for gRPC service", "error", err)
	}

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		interceptor.CheckSubnetInterceptor(
			a.trustedSubnet,
			[]string{
				"GetStats",
			},
		),
		interceptor.UserIDInterceptor(
			a.salt,
			[]string{
				"CreateShortURL",
				"GetShortURL",
				"CreateShortURLBatch",
				"GetShortURLByUserID",
				"DeleteShortURLsByShortURIs",
			},
		),
		interceptor.AuthByUserIDInterceptor(
			a.salt,
			[]string{
				"GetShortURLByUserID",
				"DeleteShortURLsByShortURIs",
			},
		),
	))
	pb.RegisterShortenerServiceServer(grpcServer, a.grpcShortenerServiceServerImpl)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-signalChan

		grpcServer.GracefulStop()

		close(gracefulShutdownCh)
	}()

	if err := grpcServer.Serve(listenTcpPort); err != nil {
		log.Fatalw("app: failed to serve grpc server", "error", err)
	}
}
