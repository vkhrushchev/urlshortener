package repository

import (
	"context"
	"github.com/vkhrushchev/urlshortener/internal/app/entity"
	"github.com/vkhrushchev/urlshortener/internal/middleware"
)

type InMemoryShortURLRepository struct {
	storage         map[string]*entity.ShortURLEntity
	storageByUserID map[string][]*entity.ShortURLEntity
}

func NewInMemoryShortURLRepository() *InMemoryShortURLRepository {
	return &InMemoryShortURLRepository{
		storage:         make(map[string]*entity.ShortURLEntity),
		storageByUserID: make(map[string][]*entity.ShortURLEntity),
	}
}

func (r *InMemoryShortURLRepository) GetShortURLByShortURI(ctx context.Context, shortURI string) (entity.ShortURLEntity, error) {
	shortURLEntry := r.storage[shortURI]
	if shortURLEntry == nil {
		return entity.ShortURLEntity{}, ErrNotFound
	}

	return *shortURLEntry, nil
}

func (r *InMemoryShortURLRepository) SaveShortURL(ctx context.Context, shortURLEntity *entity.ShortURLEntity) (*entity.ShortURLEntity, error) {
	r.storage[shortURLEntity.ShortURI] = shortURLEntity

	userID := ctx.Value(middleware.UserIDContextKey).(string)
	shortURLEntitiesByUserID := r.storageByUserID[userID]
	if shortURLEntitiesByUserID == nil {
		shortURLEntitiesByUserID = make([]*entity.ShortURLEntity, 0)
	}

	shortURLEntitiesByUserID = append(shortURLEntitiesByUserID, r.storage[shortURLEntity.ShortURI])
	r.storageByUserID[userID] = shortURLEntitiesByUserID

	return r.storage[shortURLEntity.ShortURI], nil
}

func (r *InMemoryShortURLRepository) SaveShortURLs(ctx context.Context, shortURLEntities []entity.ShortURLEntity) ([]entity.ShortURLEntity, error) {
	result := make([]entity.ShortURLEntity, 0, len(shortURLEntities))
	for _, shortURLEntity := range shortURLEntities {
		savedShortURLEntity, err := r.SaveShortURL(ctx, &shortURLEntity)
		if err != nil {
			return nil, err
		}

		result = append(result, *savedShortURLEntity)
	}

	return result, nil
}

func (r *InMemoryShortURLRepository) GetShortURLsByUserID(ctx context.Context, userID string) ([]entity.ShortURLEntity, error) {
	shortURLEntitiesByUserID := r.storageByUserID[userID]
	result := make([]entity.ShortURLEntity, 0, len(shortURLEntitiesByUserID))

	for _, shortURLEntity := range shortURLEntitiesByUserID {
		result = append(result, *shortURLEntity)
	}

	return result, nil
}

func (r *InMemoryShortURLRepository) DeleteShortURLsByShortURIs(ctx context.Context, shortURIs []string) error {
	userID := ctx.Value(middleware.UserIDContextKey).(string)
	for _, shortURI := range shortURIs {
		shortURLEntry := r.storage[shortURI]
		if shortURLEntry != nil && shortURLEntry.UserID == userID {
			shortURLEntry.Deleted = true
		}
	}

	return nil
}
