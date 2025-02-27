package grpc

import (
	"context"
	"errors"
	pb "github.com/vkhrushchev/urlshortener/grpc"
	"github.com/vkhrushchev/urlshortener/internal/app/db"
	"github.com/vkhrushchev/urlshortener/internal/app/domain"
	"github.com/vkhrushchev/urlshortener/internal/app/usecase"
	"github.com/vkhrushchev/urlshortener/internal/common"
	"github.com/vkhrushchev/urlshortener/internal/util"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

type statsProvider interface {
	GetStats(ctx context.Context) (urlCount int, userCount int, err error)
}

type ShortenerServiceServerImpl struct {
	pb.UnimplementedShortenerServiceServer
	shortURLCreator  shortURLCreator
	shortURLProvider shortURLProvider
	shortURLDeleter  shortURLDeleter
	statsProvider    statsProvider
	dbLookup         *db.DBLookup
	baseURL          string
}

func NewShortenerServiceServer(
	shortURLCreator shortURLCreator,
	shortURLProvider shortURLProvider,
	shortURLDeleter shortURLDeleter,
	statsProvider statsProvider,
	dbLookup *db.DBLookup,
	baseURL string) *ShortenerServiceServerImpl {
	return &ShortenerServiceServerImpl{
		shortURLCreator:  shortURLCreator,
		shortURLProvider: shortURLProvider,
		shortURLDeleter:  shortURLDeleter,
		statsProvider:    statsProvider,
		dbLookup:         dbLookup,
		baseURL:          baseURL,
	}
}

func (s *ShortenerServiceServerImpl) CreateShortURL(ctx context.Context, request *pb.CreateShortURLRequest) (*pb.CreateShortURLResponse, error) {
	log.Infow("grpc: CreateShortURL", "original_url", request.OriginalUrl)

	shortURLDomain, err := s.shortURLCreator.CreateShortURL(ctx, request.OriginalUrl)
	if err != nil && errors.Is(err, usecase.ErrConflict) {
		log.Infow("grpc: short URL already exists", "original_url", request.OriginalUrl)
		return nil, status.Errorf(codes.AlreadyExists, "short url already exists: %v", err)
	} else if err != nil {
		log.Errorw("grpc: CreateShortURL failed", "original_url", request.OriginalUrl, "error", err)
		return nil, status.Errorf(codes.Internal, "cannot create short url: %v", err)
	}

	response := &pb.CreateShortURLResponse{
		ShortUri: shortURLDomain.ShortURI,
		ShortUrl: util.GetShortURL(s.baseURL, shortURLDomain.ShortURI),
	}

	return response, nil
}

func (s *ShortenerServiceServerImpl) GetShortURL(ctx context.Context, request *pb.GetShortURLRequest) (*pb.GetShortURLResponse, error) {
	log.Infow("grpc: GetShortURL", "short_uri", request.ShortUri)

	shortURLDomain, err := s.shortURLProvider.GetShortURLByShortURI(ctx, request.ShortUri)
	if err != nil && errors.Is(err, usecase.ErrNotFound) {
		log.Infow("grpc: short URL not found", "short_uri", request.ShortUri)
		return nil, status.Errorf(codes.NotFound, "short url not found: %v", err)
	} else if err != nil {
		log.Errorw("grpc: GetShortURL failed", "error", err)
		return nil, status.Errorf(codes.Internal, "cannot get short url: %v", err)
	}

	response := &pb.GetShortURLResponse{
		ShortUri: shortURLDomain.ShortURI,
		ShortUrl: util.GetShortURL(s.baseURL, shortURLDomain.ShortURI),
	}

	return response, nil
}

func (s *ShortenerServiceServerImpl) CreateShortURLBatch(ctx context.Context, request *pb.CreateShortURLBatchRequest) (*pb.CreateShortURLBatchResponse, error) {
	log.Infow("gprc: CreateShortURLBatch", "batch_size", len(request.Entries))

	createShortURLBatchDomains := make([]domain.CreateShortURLBatchDomain, 0, len(request.Entries))
	for _, entry := range request.Entries {
		createShortURLBatchDomains = append(createShortURLBatchDomains, domain.CreateShortURLBatchDomain{
			CorrelationUUID: entry.CorrelationId,
			LongURL:         entry.OriginalUrl,
		})
	}

	createShortURLBatchResultDomains, err := s.shortURLCreator.CreateShortURLBatch(ctx, createShortURLBatchDomains)
	if err != nil {
		log.Errorw("grpc: CreateShortURLBatch failed", "error", err)
		return nil, status.Errorf(codes.Internal, "cannot CreateShortURLBatch: %v", err)
	}

	createShortURLBatchResponseEntries := make([]*pb.CreateShortURLBatchResponse_CreateShortURLBatchResponseEntry, 0, len(createShortURLBatchResultDomains))
	for _, createShortURLBatchResultDomain := range createShortURLBatchResultDomains {
		createShortURLBatchResponseEntries = append(createShortURLBatchResponseEntries, &pb.CreateShortURLBatchResponse_CreateShortURLBatchResponseEntry{
			CorrelationId: createShortURLBatchResultDomain.CorrelationUUID,
			ShortUrl:      util.GetShortURL(s.baseURL, createShortURLBatchResultDomain.ShortURI),
		})
	}
	createShortURLBatchResponse := &pb.CreateShortURLBatchResponse{
		Entries: createShortURLBatchResponseEntries,
	}

	return createShortURLBatchResponse, nil
}

func (s *ShortenerServiceServerImpl) GetShortURLByUserID(ctx context.Context, request *pb.GetShortURLsByUserIDRequest) (*pb.GetShortURLsByUserIDResponse, error) {
	userID := ctx.Value(common.UserIDContextKey).(string)
	log.Infow("gprc: GetShortURLByUserID", "user_id", userID)

	shortURLDomains, err := s.shortURLProvider.GetShortURLsByUserID(ctx, userID)
	if err != nil {
		log.Errorw("grpc: GetShortURLByUserID failed", "error", err)
		return nil, status.Errorf(codes.Internal, "cannot GetShortURLByUserID: %v", err)
	}

	getShortURLsByUserIDResponseEntries := make([]*pb.GetShortURLsByUserIDResponse_GetShortURLByUserIDResponseEntry, 0, len(shortURLDomains))
	for _, shortURLDomain := range shortURLDomains {
		getShortURLsByUserIDResponseEntries = append(getShortURLsByUserIDResponseEntries, &pb.GetShortURLsByUserIDResponse_GetShortURLByUserIDResponseEntry{
			ShortUrl:    util.GetShortURL(s.baseURL, shortURLDomain.ShortURI),
			OriginalUrl: shortURLDomain.LongURL,
		})
	}
	getShortURLsByUserIDResponse := &pb.GetShortURLsByUserIDResponse{
		Entries: getShortURLsByUserIDResponseEntries,
	}

	return getShortURLsByUserIDResponse, nil
}

func (s *ShortenerServiceServerImpl) DeleteShortURLsByShortURIs(ctx context.Context, request *pb.DeleteShortURLsByShortURIsRequest) (*pb.DeleteShortURLsByShortURIsResponse, error) {
	log.Infow("gprc: DeleteShortURLsByShortURIs", "batch_size", len(request.ShortURIs))

	err := s.shortURLDeleter.DeleteShortURLsByShortURIs(ctx, request.ShortURIs)
	if err != nil {
		log.Errorw("grpc: DeleteShortURLsByShortURIs failed", "error", err)
		return nil, status.Errorf(codes.Internal, "cannot DeleteShortURLsByShortURIs: %v", err)
	}

	deleteShortURLsByShortURIsResponse := &pb.DeleteShortURLsByShortURIsResponse{
		Accepted: true,
	}

	return deleteShortURLsByShortURIsResponse, nil
}

func (s *ShortenerServiceServerImpl) Ping(ctx context.Context, request *pb.PingRequest) (*pb.PingResponse, error) {
	log.Infow("gprc: Ping")
	isDBConnectionAlive := s.dbLookup.Ping(ctx)

	pingResponse := &pb.PingResponse{
		DatabaseActive: isDBConnectionAlive,
	}

	return pingResponse, nil
}

func (s *ShortenerServiceServerImpl) GetStats(ctx context.Context, request *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	log.Infow("gprc: GetStats")

	urlCount, userCount, err := s.statsProvider.GetStats(ctx)
	if err != nil {
		log.Errorw("grpc: GetStats failed", "error", err)
		return nil, status.Errorf(codes.Internal, "cannot GetStats: %v", err)
	}

	getStatsResponse := &pb.GetStatsResponse{
		UrlCount:  int64(urlCount),
		UserCount: int64(userCount),
	}

	return getStatsResponse, nil
}
