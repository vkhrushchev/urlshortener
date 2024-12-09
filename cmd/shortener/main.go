package main

import (
	"context"
	"github.com/vkhrushchev/urlshortener/internal/app/repository"
	"github.com/vkhrushchev/urlshortener/internal/app/usecase"

	"github.com/vkhrushchev/urlshortener/internal/app"
	"github.com/vkhrushchev/urlshortener/internal/app/controller"
	"github.com/vkhrushchev/urlshortener/internal/app/db"
	"go.uber.org/zap"
)

var log = zap.Must(zap.NewProduction()).Sugar()

func main() {
	parseFlags()

	dbLookup, err := db.NewDBLookup(flags.databaseDSN)
	if err != nil {
		log.Fatalf("main: error when init DBLookup: %v", err)
	}

	shortUrlRepo := initShortURLRepository(dbLookup, flags.fileStoragePath)

	createShortURLUseCase := usecase.NewCreateShortURLUseCase(shortUrlRepo)
	getShortURLUseCase := usecase.NewGetShortURLUseCase(shortUrlRepo)
	deleteShortURLUseCase := usecase.NewDeleteShortURLUseCase(shortUrlRepo)

	appController := controller.NewAppController(flags.baseURL, createShortURLUseCase, getShortURLUseCase)
	apiController := controller.NewAPIController(flags.baseURL, createShortURLUseCase, getShortURLUseCase, deleteShortURLUseCase)
	healthController := controller.NewHealthController(dbLookup)

	shortenerApp := app.NewURLShortenerApp(flags.runAddr, appController, apiController, healthController)

	shortenerApp.RegisterHandlers()
	err = shortenerApp.Run()
	if err != nil {
		log.Fatalf("main: error when run application: %v", err)
	}
}

func initShortURLRepository(dbLookup *db.DBLookup, fileStoragePath string) repository.IShortURLRepository {
	var repo repository.IShortURLRepository
	var err error

	if flags.databaseDSN != "" {
		repo = repository.NewDBShortURLRepository(dbLookup)
		err = dbLookup.InitDB(context.Background())
		if err != nil {
			log.Fatalf("main: failure to init database scheme: %v", err)
		}

		log.Infow("main: success init of DBShortURLRepository")
	}

	if repo == nil && flags.fileStoragePath != "" {
		repo, err = repository.NewJSONFileShortURLRepository(fileStoragePath)
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
