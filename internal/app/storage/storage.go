package storage

import (
	"context"

	"github.com/vkhrushchev/urlshortener/internal/app/dto"
	"go.uber.org/zap"
)

var log = zap.Must(zap.NewProduction()).Sugar()

type Storage interface {
	GetURLByShortURI(ctx context.Context, shortURI string) (*dto.StorageShortURLEntry, error)
	SaveURL(ctx context.Context, longURL string) (*dto.StorageShortURLEntry, error)
	SaveURLBatch(ctx context.Context, entries []*dto.StorageShortURLEntry) ([]*dto.StorageShortURLEntry, error)
	GetURLByUserID(ctx context.Context, userID string) ([]*dto.StorageShortURLEntry, error)
	DeleteByShortURIs(ctx context.Context, shortURIs []string) (int, error)
}
