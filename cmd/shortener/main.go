package main

import (
	"fmt"
	"github.com/vkhrushchev/urlshortener/internal/app"
)

func main() {
	shortenerApp := app.NewUrlShortenerApp()
	err := shortenerApp.Run()
	if err != nil {
		_ = fmt.Errorf("main: ошибка при запуске urlshortener: %v", err)
		return
	}
}
