package util

import (
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandStringRunes возвращает строку из случайного набора символов длинной n
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// GetShortURL возвращает короткую ссылку в виде конкатенации baseURL и shortURI
func GetShortURL(baseURL string, shortURI string) string {
	var shortURL string
	if strings.HasSuffix(baseURL, "/") {
		shortURL = baseURL + shortURI
	} else {
		shortURL = baseURL + "/" + shortURI
	}

	return shortURL
}
