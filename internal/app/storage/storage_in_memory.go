package storage

import (
	"context"

	"github.com/vkhrushchev/urlshortener/internal/app/dto"
	"github.com/vkhrushchev/urlshortener/internal/util"
)

type InMemoryStorage struct {
	storage map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		storage: make(map[string]string),
	}
}

func (s *InMemoryStorage) GetURLByShortURI(ctx context.Context, shortURI string) (longURL string, found bool, err error) {
	longURL, found = s.storage[shortURI]
	return longURL, found, nil
}

func (s *InMemoryStorage) SaveURL(ctx context.Context, longURL string) (shortURI string, err error) {
	shortURI = util.RandStringRunes(10)
	for s.storage[shortURI] != "" {
		shortURI = util.RandStringRunes(10)
	}

	s.storage[shortURI] = longURL

	return shortURI, nil
}

func (s *InMemoryStorage) SaveURLBatch(ctx context.Context, entries []*dto.StorageShortURLEntry) ([]*dto.StorageShortURLEntry, error) {
	for _, entry := range entries {
		shortURI, err := s.SaveURL(ctx, entry.LongURL)
		if err != nil {
			return nil, err
		}

		entry.ShortURI = shortURI
	}

	return entries, nil
}
