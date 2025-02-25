package usecase

import (
	"context"
	"errors"
	"github.com/vkhrushchev/urlshortener/internal/common"

	"github.com/google/uuid"
	"github.com/vkhrushchev/urlshortener/internal/app/domain"
	"github.com/vkhrushchev/urlshortener/internal/app/entity"
	"github.com/vkhrushchev/urlshortener/internal/app/repository"
	"github.com/vkhrushchev/urlshortener/internal/util"
	"go.uber.org/zap"
)

var log = zap.Must(zap.NewDevelopment()).Sugar()

// ErrConflict - короткая ссылка уже существует
// ErrNotFound - короткая ссылка не найдена
// ErrUnexpected - непредвиденная ошибка
var (
	ErrConflict   = errors.New("conflict")
	ErrNotFound   = errors.New("entity not found")
	ErrUnexpected = errors.New("unexpected error")
)

// CreateShortURLUseCase реализует интерфейс ICreateShortURLUseCase
type CreateShortURLUseCase struct {
	repo repository.IShortURLRepository
}

// NewCreateShortURLUseCase создает экземпляр CreateShortURLUseCase
func NewCreateShortURLUseCase(repo repository.IShortURLRepository) *CreateShortURLUseCase {
	return &CreateShortURLUseCase{repo: repo}
}

// CreateShortURL создает короткую ссылку
func (uc *CreateShortURLUseCase) CreateShortURL(ctx context.Context, url string) (domain.ShortURLDomain, error) {
	userID := ctx.Value(common.UserIDContextKey).(string)
	log.Infow("use_case: CreateShortURL", "url", url, "userID", userID)

	shortURLEntity := &entity.ShortURLEntity{
		UUID:     uuid.NewString(),
		ShortURI: util.RandStringRunes(10),
		LongURL:  url,
		UserID:   userID,
		Deleted:  false,
	}

	shortURLEntity, err := uc.repo.SaveShortURL(ctx, shortURLEntity)
	if err != nil && errors.Is(err, repository.ErrConflict) {
		log.Infow("use_case: conflict with existed entity", "url", url, "userID", userID)
		return domain.ShortURLDomain(*shortURLEntity), ErrConflict
	} else if err != nil {
		log.Errorw("use_case: failed to save short url", "error", err)
		return domain.ShortURLDomain{}, ErrUnexpected
	}

	return domain.ShortURLDomain(*shortURLEntity), nil
}

// CreateShortURLBatch создает короткие ссылки пачкой
func (uc *CreateShortURLUseCase) CreateShortURLBatch(ctx context.Context, createShortURLBatchDomains []domain.CreateShortURLBatchDomain) ([]domain.CreateShortURLBatchResultDomain, error) {
	userID := ctx.Value(common.UserIDContextKey).(string)
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

	shortURLEntities, err := uc.repo.SaveShortURLs(ctx, shortURLEntities)
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

// GetShortURLUseCase реализует интерфейс IGetShortURLUseCase
type GetShortURLUseCase struct {
	repo repository.IShortURLRepository
}

// NewGetShortURLUseCase создает экземпляр GetShortURLUseCase
func NewGetShortURLUseCase(repo repository.IShortURLRepository) *GetShortURLUseCase {
	return &GetShortURLUseCase{repo: repo}
}

// GetShortURLByShortURI возвращает короткую ссылку по shortURI
func (uc *GetShortURLUseCase) GetShortURLByShortURI(ctx context.Context, shortURI string) (domain.ShortURLDomain, error) {
	log.Infow("use_case: get short URL", "shortURI", shortURI)

	shortURLEntity, err := uc.repo.GetShortURLByShortURI(ctx, shortURI)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		log.Infow("use_case: short url not found", "shortURI", shortURI)
		return domain.ShortURLDomain{}, ErrNotFound
	} else if err != nil {
		log.Errorw("use_case: failed to get short url", "error", err)
		return domain.ShortURLDomain{}, ErrUnexpected
	}

	return domain.ShortURLDomain(shortURLEntity), nil
}

// GetShortURLsByUserID возвращает список коротких ссылок по userID
func (uc *GetShortURLUseCase) GetShortURLsByUserID(ctx context.Context, userID string) ([]domain.ShortURLDomain, error) {
	log.Infow("use_case: get short URLs by userID", "userID", userID)

	shortURLEntities, err := uc.repo.GetShortURLsByUserID(ctx, userID)
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

// DeleteShortURLUseCase реализует IDeleteShortURLUseCase
type DeleteShortURLUseCase struct {
	repo repository.IShortURLRepository
}

// NewDeleteShortURLUseCase создает экземпляр DeleteShortURLUseCase
func NewDeleteShortURLUseCase(repo repository.IShortURLRepository) *DeleteShortURLUseCase {
	return &DeleteShortURLUseCase{repo: repo}
}

// DeleteShortURLsByShortURIs удаляет короткие ссылки по списку shortURIs
func (uc *DeleteShortURLUseCase) DeleteShortURLsByShortURIs(ctx context.Context, shortURIs []string) error {
	userID := ctx.Value(common.UserIDContextKey).(string)
	log.Infow("use_case: delete short URLs by shortURIs", "shortURIs", shortURIs, "userID", userID)

	err := uc.repo.DeleteShortURLsByShortURIs(ctx, shortURIs)
	if err != nil {
		log.Errorw("use_case: failed to delete short URLs by shortURIs", "shortURIs", shortURIs, "userID", userID, "error", err)
		return ErrUnexpected
	}

	return nil
}

// StatsUseCase структура реализующая интерфейс IStatsUseCase
type StatsUseCase struct {
	repo repository.IShortURLRepository
}

// NewStatsUseCase создает экземпляр StatsUseCase
func NewStatsUseCase(repo repository.IShortURLRepository) *StatsUseCase {
	return &StatsUseCase{repo: repo}
}

// GetStats возвращает статистику по сервису
// urlCount - количество коротких ссылок в сервисе
// userCount - количество пользователей в сервисе
func (uc *StatsUseCase) GetStats(ctx context.Context) (urlCount int, userCount int, err error) {
	log.Infow("use_case: get stats")

	urlCount, userCount, err = uc.repo.GetStats(ctx)
	if err != nil {
		log.Errorw("use_case: failed to get stats", "error", err)
		return 0, 0, ErrUnexpected
	}

	return urlCount, userCount, nil
}
