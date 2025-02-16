package main

import (
	"context"
	"github.com/vkhrushchev/urlshortener/config"
	"github.com/vkhrushchev/urlshortener/internal/app/grpc"
	"github.com/vkhrushchev/urlshortener/internal/app/repository"
	"github.com/vkhrushchev/urlshortener/internal/app/usecase"
	"net"

	"github.com/vkhrushchev/urlshortener/internal/app"
	"github.com/vkhrushchev/urlshortener/internal/app/controller"
	"github.com/vkhrushchev/urlshortener/internal/app/db"
	"go.uber.org/zap"
)

var log = zap.Must(zap.NewDevelopment()).Sugar()

// buildVersion = определяет версию приложения
// buildDate = определяет дату сборки
// buildCommit = определяет коммит сборки
var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	log.Infof("Build version: %s\n", buildVersion)
	log.Infof("Build date: %s\n", buildDate)
	log.Infof("Build commit: %s\n", buildCommit)

	shortenerConfig := config.ReadConfig()
	_, trustedSubnet, err := net.ParseCIDR(shortenerConfig.TrustedSubnet)
	if err != nil {
		log.Warnf("main: failed to parse trusted subnet: %v", err)
	}

	dbLookup, err := db.NewDBLookup(shortenerConfig.DatabaseDSN)
	if err != nil {
		log.Fatalf("main: error when init DBLookup: %v", err)
	}

	shortURLRepo := initShortURLRepository(dbLookup, shortenerConfig)

	createShortURLUseCase := usecase.NewCreateShortURLUseCase(shortURLRepo)
	getShortURLUseCase := usecase.NewGetShortURLUseCase(shortURLRepo)
	deleteShortURLUseCase := usecase.NewDeleteShortURLUseCase(shortURLRepo)
	statsUseCase := usecase.NewStatsUseCase(shortURLRepo)

	appController := controller.NewAppController(shortenerConfig.BaseURL, createShortURLUseCase, getShortURLUseCase)
	apiController := controller.NewAPIController(
		shortenerConfig.BaseURL, createShortURLUseCase, getShortURLUseCase, deleteShortURLUseCase)
	healthController := controller.NewHealthController(dbLookup)
	internalController := controller.NewInternalController(statsUseCase)

	grpcShortenerServiceServer := grpc.NewShortenerServiceServer(
		createShortURLUseCase,
		getShortURLUseCase,
		deleteShortURLUseCase,
		statsUseCase,
		dbLookup,
		shortenerConfig.BaseURL,
	)

	shortenerApp := app.NewURLShortenerApp(
		shortenerConfig.RunAddr,
		shortenerConfig.EnableHTTPS,
		trustedSubnet,
		appController,
		apiController,
		healthController,
		internalController,
		grpcShortenerServiceServer,
	)

	shortenerApp.RegisterHTTPHandlers()

	gracefulHTTPShutdownChan := make(chan struct{})
	gracefulGRPCShutdownChan := make(chan struct{})
	go shortenerApp.RunHTTPServer(gracefulHTTPShutdownChan)
	go shortenerApp.RunGRPCServer(gracefulGRPCShutdownChan)

	<-gracefulHTTPShutdownChan
	log.Infow("main: URLShortenerApp HTTP shutting down")
	<-gracefulGRPCShutdownChan
	log.Infow("main: URLShortenerApp GRPC shutting down")
}

func initShortURLRepository(dbLookup *db.DBLookup, config config.Config) repository.IShortURLRepository {
	var repo repository.IShortURLRepository
	var err error

	if config.DatabaseDSN != "" {
		repo = repository.NewDBShortURLRepository(dbLookup)
		err = dbLookup.InitDB(context.Background())
		if err != nil {
			log.Fatalf("main: failure to init database scheme: %v", err)
		}

		log.Infow("main: success init of DBShortURLRepository")
	}

	if repo == nil && config.FileStoragePath != "" {
		repo, err = repository.NewJSONFileShortURLRepository(config.FileStoragePath)
		if err != nil {
			log.Fatalf("main: failure to init JSONFileShortURLRepository: %v", err)
		}

		log.Infow("main: success init of JSONFileShortURLRepository")
	}

	if repo == nil {
		repo = repository.NewInMemoryShortURLRepository()

		log.Infow("main: success init of InMemoryShortURLRepository")
	}

	return repo
}
