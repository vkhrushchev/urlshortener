package main

import (
	"fmt"
	"os"

	"github.com/vkhrushchev/urlshortener/internal/app"
	"github.com/vkhrushchev/urlshortener/internal/app/storage"
)

func main() {
	parseFlags()

	storage, err := storage.NewFileJsonStorage(flags.fileStoragePathEnv)
	if err != nil {
		err = fmt.Errorf("main: ошибка при инициализации FileJsonStorage: %v", err)
		println(err.Error())
		os.Exit(1)
	}

	shortenerApp := app.NewURLShortenerApp(flags.runAddr, flags.baseURL, storage)

	shortenerApp.RegisterHandlers()
	err = shortenerApp.Run()
	if err != nil {
		err = fmt.Errorf("main: ошибка при запуске urlshortener: %v", err)
		println(err.Error())
		os.Exit(1)
	}
}
