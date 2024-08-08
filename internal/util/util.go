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

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func GetShortURL(baseURL string, urlID string) string {
	var shortURL string
	if strings.HasSuffix(baseURL, "/") {
		shortURL = baseURL + urlID
	} else {
		shortURL = baseURL + "/" + urlID
	}

	return shortURL
}
