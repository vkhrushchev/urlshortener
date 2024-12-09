package repository

import (
	"context"
	"errors"

	"github.com/vkhrushchev/urlshortener/internal/app/entity"
	"go.uber.org/zap"
)

var log = zap.Must(zap.NewDevelopment()).Sugar()

var (
	ErrConflict   = errors.New("conflict")         // Короткая ссылка уже существует
	ErrNotFound   = errors.New("entity not found") // Короткая ссылка не найдена
	ErrUnexpected = errors.New("unexpected error") // Непредвиденная ошибка
)

// IShortURLRepository интерфейс репозитория для сущности ShortURL
type IShortURLRepository interface {
	GetShortURLByShortURI(ctx context.Context, shortURI string) (entity.ShortURLEntity, error)
	SaveShortURL(ctx context.Context, shortURLEntity *entity.ShortURLEntity) (*entity.ShortURLEntity, error)
	SaveShortURLs(ctx context.Context, shortURLEntities []entity.ShortURLEntity) ([]entity.ShortURLEntity, error)
	GetShortURLsByUserID(ctx context.Context, userID string) ([]entity.ShortURLEntity, error)
	DeleteShortURLsByShortURIs(ctx context.Context, shortURIs []string) error
}
