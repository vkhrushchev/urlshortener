package storage

import (
	"context"

	"github.com/vkhrushchev/urlshortener/internal/app/dto"
	"go.uber.org/zap"
)

var log = zap.Must(zap.NewProduction()).Sugar()

type Storage interface {
	GetURLByShortURI(ctx context.Context, shortURI string) (longURL string, found bool, err error)
	SaveURL(ctx context.Context, longURL string) (shortURI string, err error)
	SaveURLBatch(ctx context.Context, entries []dto.StorageShortURLEntry) ([]dto.StorageShortURLEntry, error)
}
