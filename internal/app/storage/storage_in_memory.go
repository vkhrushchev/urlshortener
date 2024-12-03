package storage

import (
	"context"

	"github.com/google/uuid"
	"github.com/vkhrushchev/urlshortener/internal/app/dto"
	"github.com/vkhrushchev/urlshortener/internal/middleware"
	"github.com/vkhrushchev/urlshortener/internal/util"
)

type InMemoryStorage struct {
	storage         map[string]*dto.StorageShortURLEntry
	storageByUserID map[string][]*dto.StorageShortURLEntry
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		storage:         make(map[string]*dto.StorageShortURLEntry),
		storageByUserID: make(map[string][]*dto.StorageShortURLEntry),
	}
}

func (s *InMemoryStorage) GetURLByShortURI(ctx context.Context, shortURI string) (*dto.StorageShortURLEntry, error) {
	shortURLEntry := s.storage[shortURI]
	return shortURLEntry, nil
}

func (s *InMemoryStorage) SaveURL(ctx context.Context, longURL string) (*dto.StorageShortURLEntry, error) {
	shortURI := util.RandStringRunes(10)
	for s.storage[shortURI] != nil {
		shortURI = util.RandStringRunes(10)
	}

	userID := ctx.Value(middleware.UserIDContextKey).(string)
	s.storage[shortURI] = &dto.StorageShortURLEntry{
		UUID:     uuid.NewString(),
		ShortURI: shortURI,
		LongURL:  longURL,
		UserID:   userID,
		Deleted:  false,
	}

	shortURLEntriesByUserID := s.storageByUserID[userID]
	if shortURLEntriesByUserID != nil {
		shortURLEntriesByUserID = make([]*dto.StorageShortURLEntry, 0)
	}

	shortURLEntriesByUserID = append(shortURLEntriesByUserID, s.storage[shortURI])
	s.storageByUserID[userID] = shortURLEntriesByUserID

	return s.storage[shortURI], nil
}

func (s *InMemoryStorage) SaveURLBatch(ctx context.Context, entries []*dto.StorageShortURLEntry) ([]*dto.StorageShortURLEntry, error) {
	for _, entry := range entries {
		savedShortURLEntry, err := s.SaveURL(ctx, entry.LongURL)
		if err != nil {
			return nil, err
		}

		entry.ShortURI = savedShortURLEntry.ShortURI
	}

	return entries, nil
}

func (s *InMemoryStorage) GetURLByUserID(ctx context.Context, userID string) ([]*dto.StorageShortURLEntry, error) {
	return s.storageByUserID[userID], nil
}

func (s *InMemoryStorage) DeleteByShortURIs(ctx context.Context, shortURIs []string) error {
	userID := ctx.Value(middleware.UserIDContextKey).(string)
	for _, shortURI := range shortURIs {
		shortURLEntry := s.storage[shortURI]
		if shortURLEntry != nil && shortURLEntry.UserID == userID {
			shortURLEntry.Deleted = true
		}
	}

	return nil
}
