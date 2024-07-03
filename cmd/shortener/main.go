package main

import (
	"fmt"
	"github.com/vkhrushchev/urlshortener/internal/app"
)

func main() {
	parseFlags()

	shortenerApp := app.NewURLShortenerApp(flags.runAddr, flags.baseURL)
	shortenerApp.RegisterHandlers()
	err := shortenerApp.Run()
	if err != nil {
		_ = fmt.Errorf("main: ошибка при запуске urlshortener: %v", err)
		return
	}
}
