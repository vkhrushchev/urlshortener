package main

import (
	"github.com/vkhrushchev/urlshortener/internal/app"
	"github.com/vkhrushchev/urlshortener/internal/app/controller"
	"github.com/vkhrushchev/urlshortener/internal/app/db"
	"github.com/vkhrushchev/urlshortener/internal/app/storage"
	"go.uber.org/zap"
)

var log = zap.Must(zap.NewProduction()).Sugar()

func main() {
	parseFlags()

	storage, err := storage.NewFileJSONStorage(flags.fileStoragePathEnv)
	if err != nil {
		log.Fatalf("main: ошибка при инициализации FileJsonStorage: %v", err)
	}

	dbLookup, err := db.NewDBLookup(flags.databaseDSN)
	if err != nil {
		log.Fatalf("main: ошибка при инициализации DBLookUp: %v", err)
	}

	healthController := controller.NewHealthController(dbLookup)

	shortenerApp := app.NewURLShortenerApp(flags.runAddr, flags.baseURL, storage, *healthController)

	shortenerApp.RegisterHandlers()
	err = shortenerApp.Run()
	if err != nil {
		log.Fatalf("main: ошибка при инициализации FileJsonStorage: %v", err)
	}
}
