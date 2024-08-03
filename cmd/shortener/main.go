package main

import (
	"context"

	"github.com/vkhrushchev/urlshortener/internal/app"
	"github.com/vkhrushchev/urlshortener/internal/app/controller"
	"github.com/vkhrushchev/urlshortener/internal/app/db"
	"github.com/vkhrushchev/urlshortener/internal/app/storage"
	"go.uber.org/zap"
)

var log = zap.Must(zap.NewProduction()).Sugar()

func main() {
	parseFlags()

	dbLookup, err := db.NewDBLookup(flags.databaseDSN)
	if err != nil {
		log.Fatalf("main: ошибка при инициализации DBLookUp: %v", err)
	}

	storage := initStorage(dbLookup, flags.fileStoragePath)
	healthController := controller.NewHealthController(dbLookup)

	shortenerApp := app.NewURLShortenerApp(flags.runAddr, flags.baseURL, storage, *healthController)

	shortenerApp.RegisterHandlers()
	err = shortenerApp.Run()
	if err != nil {
		log.Fatalf("main: ошибка при инициализации FileJsonStorage: %v", err)
	}
}

func initStorage(dbLookup *db.DBLookup, fileStoragePath string) storage.Storage {
	var store storage.Storage
	var err error
	if flags.databaseDSN != "" {
		store = storage.NewDBStorage(dbLookup)
		err = dbLookup.InitDB(context.Background())
		if err != nil {
			log.Fatalf("main: ошибка при инициализации структуры БД: %v", err)
		}
	}

	if store == nil && flags.fileStoragePath != "" {
		store, err = storage.NewFileJSONStorage(fileStoragePath)
		if err != nil {
			log.Fatalf("main: ошибка при инициализации FileJsonStorage: %v", err)
		}
	}

	if store == nil {
		store = storage.NewInMemoryStorage()
	}

	return store
}
