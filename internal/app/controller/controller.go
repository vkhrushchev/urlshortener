// Package controller
//
//	@Title			Shortener API
//	@Description	Сервис сокращения ссылок
//	@Version		1.0
package controller

import (
	"context"
	"github.com/vkhrushchev/urlshortener/internal/app/domain"
	"go.uber.org/zap"
)

var log = zap.Must(zap.NewDevelopment()).Sugar()

type shortURLCreator interface {
	CreateShortURL(ctx context.Context, url string) (domain.ShortURLDomain, error)
	CreateShortURLBatch(ctx context.Context, createShortURLBatchDomains []domain.CreateShortURLBatchDomain) ([]domain.CreateShortURLBatchResultDomain, error)
}

type shortURLProvider interface {
	GetShortURLByShortURI(ctx context.Context, shortURI string) (domain.ShortURLDomain, error)
	GetShortURLsByUserID(ctx context.Context, userID string) ([]domain.ShortURLDomain, error)
}

type shortURLDeleter interface {
	DeleteShortURLsByShortURIs(ctx context.Context, shortURIs []string) error
}
