package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vkhrushchev/urlshortener/internal/app/db"
	"github.com/vkhrushchev/urlshortener/internal/app/entity"
	"github.com/vkhrushchev/urlshortener/internal/middleware"
	"github.com/vkhrushchev/urlshortener/internal/util"
)

type DBShortURLRepositoryTestSuite struct {
	suite.Suite
	postgresContainer *postgres.PostgresContainer
	repository        *DBShortURLRepository
}

func (s *DBShortURLRepositoryTestSuite) SetupSuite() {
	ctx := context.Background()
	postgresContainer, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("urlshortener"),
		postgres.WithUsername("urlshortener"),
		postgres.WithPassword("urlshortener"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		s.Fail("repository: failed to start postgres container: %v", err)
	}

	s.postgresContainer = postgresContainer

	dsnString := s.postgresContainer.MustConnectionString(ctx, "sslmode=disable")
	dbLookup, err := db.NewDBLookup(dsnString)
	if err != nil {
		s.Fail("repository: failed to create dbLookup: %v", err)
	}

	if err := dbLookup.InitDB(ctx); err != nil {
		s.Fail("repository: failed to init dbLookup: %v", err)
	}

	s.repository = NewDBShortURLRepository(dbLookup)
}

func (s *DBShortURLRepositoryTestSuite) TearDownSuite() {
	err := s.postgresContainer.Terminate(context.Background())
	if err != nil {
		s.Fail("repository: failed to terminate postgres container: %v", err)
	}
}

func (s *DBShortURLRepositoryTestSuite) TestGetShortURLByShortURI_not_found() {
	testUserID := uuid.NewString()
	testCtx := context.WithValue(context.Background(), middleware.UserIDContextKey, testUserID)

	_, err := s.repository.GetShortURLByShortURI(testCtx, "not_existed_shortURL")
	s.ErrorIs(err, ErrNotFound, "expected ErrNotFound, got %v", err)
}

func (s *DBShortURLRepositoryTestSuite) TestSaveShortURL_conflict() {
	testUserID := uuid.NewString()
	testCtx := context.WithValue(context.Background(), middleware.UserIDContextKey, testUserID)
	testShortURL := &entity.ShortURLEntity{
		UUID:     uuid.NewString(),
		ShortURI: util.RandStringRunes(10),
		LongURL:  "https://mail.ru/" + util.RandStringRunes(10),
		UserID:   testUserID,
		Deleted:  false,
	}

	_, err := s.repository.SaveShortURL(testCtx, testShortURL)
	if err != nil {
		s.Failf("failed to save shortURL", "error: %v", err)
	}

	savedShortURL, err := s.repository.SaveShortURL(testCtx, testShortURL)
	s.ErrorIs(err, ErrConflict, "expected ErrConflict, got %v", err)
	s.NotNil(savedShortURL, "savedShortURL should not be nil")
}

func (s *DBShortURLRepositoryTestSuite) TestSaveShortURLs_success() {
	testUserID := uuid.NewString()
	testCtx := context.WithValue(context.Background(), middleware.UserIDContextKey, testUserID)
	testShortURLFirst := entity.ShortURLEntity{
		UUID:     uuid.NewString(),
		ShortURI: util.RandStringRunes(10),
		LongURL:  "https://mail.ru/" + util.RandStringRunes(10),
		UserID:   testUserID,
		Deleted:  false,
	}
	testShortURLSecond := entity.ShortURLEntity{
		UUID:     uuid.NewString(),
		ShortURI: util.RandStringRunes(10),
		LongURL:  "https://mail.ru/" + util.RandStringRunes(10),
		UserID:   testUserID,
		Deleted:  false,
	}

	savedShortURLEntities, err := s.repository.SaveShortURLs(
		testCtx,
		[]entity.ShortURLEntity{testShortURLFirst, testShortURLSecond},
	)
	if err != nil {
		s.Failf("failed to save shortURLs", "error: %v", err)
	}

	s.NotNil(savedShortURLEntities, "savedShortURLEntities should not be nil")
	s.Equal(2, len(savedShortURLEntities), "savedShortURLEntities should contain 2 entities")
}

func (s *DBShortURLRepositoryTestSuite) TestFull() {
	testUserID := uuid.NewString()
	testCtx := context.WithValue(context.Background(), middleware.UserIDContextKey, testUserID)
	testShortURL := &entity.ShortURLEntity{
		UUID:     uuid.NewString(),
		ShortURI: util.RandStringRunes(10),
		LongURL:  "https://mail.ru/" + util.RandStringRunes(10),
		UserID:   testUserID,
		Deleted:  false,
	}

	savedShortURL, err := s.repository.SaveShortURL(testCtx, testShortURL)
	if err != nil {
		s.Fail("unexpected error when save ShortURLEntity: %v", err)
	}
	s.NotNil(savedShortURL, "shortURL should not be nil")

	shortURL, err := s.repository.GetShortURLByShortURI(testCtx, savedShortURL.ShortURI)
	if err != nil {
		s.Fail("unexpected error when get ShortURLEntity by shortURI: %v", err)
	}
	s.NotNil(shortURL, "shortURL should not be nil")

	savedShortURLsByUserID, err := s.repository.GetShortURLsByUserID(testCtx, shortURL.UserID)
	if err != nil {
		s.Fail("unexpected error when get ShortURLEntity by shortURI: %v", err)
	}
	s.NotNil(savedShortURLsByUserID, "savedShortURLsByUserID should not be nil")
	s.Equal(1, len(savedShortURLsByUserID), "savedShortURLsByUserID len mast equal 1")

	s.repository.DeleteShortURLsByShortURIs(testCtx, []string{shortURL.ShortURI, "not_existed_shortURL"})
	if err != nil {
		s.Fail("unexpected error when delete ShortURLEntities by shortURIs: %v", err)
	}
}

func (s *DBShortURLRepositoryTestSuite) TestGetStats() {
	urlCount, userCount, err := s.repository.GetStats(context.Background())
	if err != nil {
		s.Fail("unexpected error when get stats", "error: %v", err)
	}

	s.Equal(1, urlCount, "urlCount should be 1")
	s.Equal(1, userCount, "userCount should be 1")
}

func TestDBShortURLRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(DBShortURLRepositoryTestSuite))
}
