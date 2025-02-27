package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
)

func main() {
	endpoint := "http://localhost:8080/"
	// приглашение в консоли
	fmt.Println("Введите длинный URL")
	// открываем потоковое чтение из консоли
	reader := bufio.NewReader(os.Stdin)
	// читаем строку из консоли
	longURL, err := reader.ReadString('\n')
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	longURL = strings.TrimSuffix(longURL, "\n")

	// добавляем HTTP-клиент
	restyClient := resty.New()
	restyClient.SetBaseURL(endpoint)

	// пишем и выполняем запрос
	restyResponse, err := restyClient.R().
		SetHeader("Content-Type", "plain/text").
		SetBody(longURL).
		Post("/")
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	// выводим результат в консоль
	fmt.Println("========= STATUS CODE =========")
	fmt.Println(restyResponse.StatusCode())
	fmt.Println("========= BODY =========")
	fmt.Println(string(restyResponse.Body()))
}
