package usecase

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/vkhrushchev/urlshortener/internal/app/domain"
	"github.com/vkhrushchev/urlshortener/internal/app/entity"
	"github.com/vkhrushchev/urlshortener/internal/app/repository"
	"github.com/vkhrushchev/urlshortener/internal/middleware"
	"github.com/vkhrushchev/urlshortener/internal/util"
	"go.uber.org/zap"
)

var log = zap.Must(zap.NewDevelopment()).Sugar()

var (
	ErrConflict   = errors.New("conflict")
	ErrNotFound   = errors.New("not found")
	ErrUnexpected = errors.New("unexpected error")
)

type ICreateShortURLUseCase interface {
	CreateShortURL(ctx context.Context, url string) (domain.ShortURLDomain, error)
	CreateShortURLBatch(ctx context.Context, createShortURLBatchDomains []domain.CreateShortURLBatchDomain) ([]domain.CreateShortURLBatchResultDomain, error)
}

type CreateShortURLUseCase struct {
	repository repository.IShortURLRepository
}

func NewCreateShortURLUseCase(r repository.IShortURLRepository) *CreateShortURLUseCase {
	return &CreateShortURLUseCase{repository: r}
}

func (uc *CreateShortURLUseCase) CreateShortURL(ctx context.Context, url string) (domain.ShortURLDomain, error) {
	userID := ctx.Value(middleware.UserIDContextKey).(string)
	log.Infow("use_case: create short URL", "url", url, "userID", userID)

	shortURLEntity := entity.ShortURLEntity{
		UUID:     uuid.NewString(),
		ShortURI: util.RandStringRunes(10),
		LongURL:  url,
		UserID:   userID,
		Deleted:  false,
	}

	shortURLEntity, err := uc.repository.SaveShortURL(ctx, shortURLEntity)
	if err != nil && errors.Is(err, repository.ErrConflict) {
		log.Infow("use_case: conflict with existed entity", "url", url, "userID", userID)
		return domain.ShortURLDomain{}, ErrConflict
	} else if err != nil {
		log.Errorw("use_case: failed to save short url", "error", err)
		return domain.ShortURLDomain{}, ErrUnexpected
	}

	return domain.ShortURLDomain(shortURLEntity), nil
}

func (uc *CreateShortURLUseCase) CreateShortURLBatch(ctx context.Context, createShortURLBatchDomains []domain.CreateShortURLBatchDomain) ([]domain.CreateShortURLBatchResultDomain, error) {
	userID := ctx.Value(middleware.UserIDContextKey).(string)
	log.Infow("use_case: create short URL batch", "userID", userID)

	shortURLEntities := make([]entity.ShortURLEntity, 0, len(createShortURLBatchDomains))
	for _, createShortURLBatchDomain := range createShortURLBatchDomains {
		shortURLEntity := entity.ShortURLEntity{
			UUID:     createShortURLBatchDomain.CorrelationUUID,
			ShortURI: util.RandStringRunes(10),
			LongURL:  createShortURLBatchDomain.LongURL,
			UserID:   userID,
			Deleted:  false,
		}

		shortURLEntities = append(shortURLEntities, shortURLEntity)
	}

	shortURLEntities, err := uc.repository.SaveShortURLs(ctx, shortURLEntities)
	if err != nil {
		log.Errorw("use_case: failed to save short URL batch", "error", err)
		return nil, ErrUnexpected
	}

	result := make([]domain.CreateShortURLBatchResultDomain, 0, len(createShortURLBatchDomains))
	for _, shortURLEntity := range shortURLEntities {
		createShortURLBatchResultDomain := domain.CreateShortURLBatchResultDomain{
			CorrelationUUID: shortURLEntity.UUID,
			ShortURI:        shortURLEntity.ShortURI,
		}

		result = append(result, createShortURLBatchResultDomain)
	}

	return result, nil
}

type IGetShortURLUseCase interface {
	GetShortURL(ctx context.Context, shortURI string) (domain.ShortURLDomain, error)
	GetShortURLsByUserID(ctx context.Context, userID string) ([]domain.ShortURLDomain, error)
}

type GetShortURLUseCase struct {
	repository repository.IShortURLRepository
}

func NewGetShortURLUseCase(r repository.IShortURLRepository) *GetShortURLUseCase {
	return &GetShortURLUseCase{repository: r}
}

func (uc *GetShortURLUseCase) GetShortURL(ctx context.Context, shortURI string) (domain.ShortURLDomain, error) {
	log.Infow("use_case: get short URL", "shortURI", shortURI)

	shortURLEntity, err := uc.repository.GetShortURLByShortURI(ctx, shortURI)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		log.Infow("use_case: short url not found", "shortURI", shortURI)
		return domain.ShortURLDomain{}, ErrNotFound
	} else if err != nil {
		log.Errorw("use_case: failed to get short url", "error", err)
		return domain.ShortURLDomain{}, ErrUnexpected
	}

	return domain.ShortURLDomain(shortURLEntity), nil
}

func (uc *GetShortURLUseCase) GetShortURLsByUserID(ctx context.Context, userID string) ([]domain.ShortURLDomain, error) {
	log.Infow("use_case: get short URLs by userID", "userID", userID)

	shortURLEntities, err := uc.repository.GetShortURLsByUserID(ctx, userID)
	if err != nil {
		log.Errorw("use_case: failed to get short urls by userID", "userID", userID, "error", err)
		return nil, ErrUnexpected
	}

	result := make([]domain.ShortURLDomain, 0, len(shortURLEntities))
	for _, shortURLEntity := range shortURLEntities {
		shortURLDomain := domain.ShortURLDomain(shortURLEntity)

		result = append(result, shortURLDomain)
	}

	return result, nil
}

type IDeleteShortURLUseCase interface {
	DeleteShortURLsByShortURIs(ctx context.Context, shortURIs []string) error
}

type DeleteShortURLUseCase struct {
	repository repository.IShortURLRepository
}

func NewDeleteShortURLUseCase(r repository.IShortURLRepository) *DeleteShortURLUseCase {
	return &DeleteShortURLUseCase{repository: r}
}

func (uc *DeleteShortURLUseCase) DeleteShortURLsByShortURIs(ctx context.Context, shortURIs []string) error {
	userID := ctx.Value(middleware.UserIDContextKey).(string)
	log.Infow("use_case: delete short URLs by shortURIs", "shortURIs", shortURIs, "userID", userID)

	err := uc.repository.DeleteShortURLsByShortURIs(ctx, shortURIs)
	if err != nil {
		log.Errorw("use_case: failed to delete short URLs by shortURIs", "shortURIs", shortURIs, "userID", userID, "error", err)
		return ErrUnexpected
	}

	return nil
}
