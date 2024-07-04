package main

import (
	"fmt"
	"github.com/vkhrushchev/urlshortener/internal/app"
	"os"
)

func main() {
	parseFlags()

	shortenerApp := app.NewURLShortenerApp(flags.runAddr, flags.baseURL)
	shortenerApp.RegisterHandlers()
	err := shortenerApp.Run()
	if err != nil {
		err = fmt.Errorf("main: ошибка при запуске urlshortener: %v", err)
		if err != nil {
			println(err.Error())
		}
		os.Exit(1)
	}
}
