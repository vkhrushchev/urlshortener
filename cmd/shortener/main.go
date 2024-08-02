package main

import (
	"github.com/vkhrushchev/urlshortener/internal/app"
	"github.com/vkhrushchev/urlshortener/internal/app/storage"
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func init() {
	log = zap.Must(zap.NewProduction()).Sugar()
}

func main() {
	parseFlags()

	storage, err := storage.NewFileJSONStorage(flags.fileStoragePathEnv)
	if err != nil {
		log.Fatalf("main: ошибка при инициализации FileJsonStorage: %v", err)
	}

	shortenerApp := app.NewURLShortenerApp(flags.runAddr, flags.baseURL, storage)

	shortenerApp.RegisterHandlers()
	err = shortenerApp.Run()
	if err != nil {
		log.Fatalf("main: ошибка при инициализации FileJsonStorage: %v", err)
	}
}
