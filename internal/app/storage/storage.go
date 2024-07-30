package storage

import "github.com/vkhrushchev/urlshortener/internal/util"

type Storage interface {
	GetURLByShortURI(shortURI string) (longURL string, found bool)
	SaveURL(longURL string) (shortURI string)
}

type InMemoryStorage struct {
	storage map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		storage: make(map[string]string),
	}
}

func (s *InMemoryStorage) GetURLByShortURI(shortURI string) (longURL string, found bool) {
	longURL, found = s.storage[shortURI]
	return longURL, found
}

func (s *InMemoryStorage) SaveURL(longURL string) (shortURI string) {
	shortURI = util.RandStringRunes(10)
	for s.storage[shortURI] != "" {
		shortURI = util.RandStringRunes(10)
	}

	s.storage[shortURI] = longURL

	return shortURI
}
