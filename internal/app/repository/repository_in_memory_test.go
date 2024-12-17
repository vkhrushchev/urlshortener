package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/vkhrushchev/urlshortener/internal/app/entity"
	"github.com/vkhrushchev/urlshortener/internal/middleware"
	"github.com/vkhrushchev/urlshortener/internal/util"
)

type InMemoryRepositoryTestSuite struct {
	suite.Suite
	repository         *InMemoryShortURLRepository
	testUserIDFirst    string
	testUserIDSecond   string
	testShortURLFirst  entity.ShortURLEntity
	testShortURLSecond entity.ShortURLEntity
}

func (suite *InMemoryRepositoryTestSuite) SetupTest() {
	suite.testUserIDFirst = uuid.NewString()
	suite.testShortURLFirst = entity.ShortURLEntity{
		UUID:     uuid.NewString(),
		ShortURI: "abc",
		LongURL:  "https://ya.ru",
		UserID:   suite.testUserIDFirst,
		Deleted:  false,
	}
	suite.testUserIDSecond = uuid.NewString()
	suite.testShortURLSecond = entity.ShortURLEntity{
		UUID:     uuid.NewString(),
		ShortURI: "cde",
		LongURL:  "https://google.com",
		UserID:   suite.testUserIDSecond,
		Deleted:  false,
	}

	suite.repository = NewInMemoryShortURLRepository()
	suite.repository.storage = map[string]*entity.ShortURLEntity{
		suite.testShortURLFirst.ShortURI:  &suite.testShortURLFirst,
		suite.testShortURLSecond.ShortURI: &suite.testShortURLSecond,
	}
	suite.repository.storageByUserID = map[string][]*entity.ShortURLEntity{
		suite.testUserIDFirst:  {&suite.testShortURLFirst},
		suite.testUserIDSecond: {&suite.testShortURLSecond},
	}
}

func (suite *InMemoryRepositoryTestSuite) TestGetShortURLByShortURI_success() {
	shortURLEntity, err := suite.repository.GetShortURLByShortURI(context.Background(), "abc")
	if err != nil {
		suite.Error(err, "unexpected error when get short url by short uri")
	}

	suite.Equal(suite.testShortURLFirst, shortURLEntity)
}

func (suite *InMemoryRepositoryTestSuite) TestGetShortURLByShortURI_not_found() {
	_, err := suite.repository.GetShortURLByShortURI(context.Background(), "cba")
	if err != nil && !errors.Is(err, ErrNotFound) {
		suite.Error(err, "unexpected error when get short url by short uri")
	}
}

func (suite *InMemoryRepositoryTestSuite) TestSaveShortURL_success() {
	testUserID := uuid.NewString()
	testShortURL := &entity.ShortURLEntity{
		UUID:     uuid.NewString(),
		ShortURI: "ghi",
		LongURL:  "https://mail.ru",
		UserID:   testUserID,
		Deleted:  false,
	}

	testCtx := context.WithValue(context.Background(), middleware.UserIDContextKey, testUserID)
	savedShortURL, err := suite.repository.SaveShortURL(testCtx, testShortURL)
	if err != nil {
		suite.Error(err, "unexpected error when save ShortURLEntity")
	}

	suite.NotNil(savedShortURL, "savedShortURL should not be nil")
	suite.NotNil(suite.repository.storage[testShortURL.ShortURI], "testShortURL should be saved in storage")
	suite.NotNilf(suite.repository.storageByUserID[testUserID], "storageByUserID must be created")
	suite.Equalf(1, len(suite.repository.storageByUserID[testUserID]), "testShortURL must be saved in storageByUserID")
}

func (suite *InMemoryRepositoryTestSuite) TestSaveShortURL_existed_user() {
	testShortURL := &entity.ShortURLEntity{
		UUID:     uuid.NewString(),
		ShortURI: "ghi",
		LongURL:  "https://mail.ru",
		UserID:   suite.testUserIDFirst,
		Deleted:  false,
	}

	testCtx := context.WithValue(context.Background(), middleware.UserIDContextKey, suite.testUserIDFirst)
	savedShortURL, err := suite.repository.SaveShortURL(testCtx, testShortURL)
	if err != nil {
		suite.Error(err, "unexpected error when save ShortURLEntity")
	}

	suite.NotNil(savedShortURL, "savedShortURL should not be nil")
	suite.NotNil(suite.repository.storage[testShortURL.ShortURI], "testShortURL should be saved in storage")
	suite.NotNilf(suite.repository.storageByUserID[suite.testUserIDFirst], "storageByUserID must be created")
	suite.Equalf(2, len(suite.repository.storageByUserID[suite.testUserIDFirst]), "testShortURL must be saved in storageByUserID")
}

func (suite *InMemoryRepositoryTestSuite) TestSaveShortURLs_success() {
	testShortURLEntities := []entity.ShortURLEntity{
		{
			UUID:     uuid.NewString(),
			ShortURI: util.RandStringRunes(10),
			LongURL:  "https://mail.ru",
			UserID:   suite.testUserIDFirst,
			Deleted:  false,
		},
		{
			UUID:     uuid.NewString(),
			ShortURI: util.RandStringRunes(10),
			LongURL:  "https://vk.com",
			UserID:   suite.testUserIDFirst,
			Deleted:  false,
		},
	}

	testCtx := context.WithValue(context.Background(), middleware.UserIDContextKey, suite.testUserIDFirst)
	savedShortURLEntities, err := suite.repository.SaveShortURLs(testCtx, testShortURLEntities)
	if err != nil {
		suite.Error(err, "unexpected error when save ShortURLEntities")
	}

	suite.Equal(2, len(savedShortURLEntities), "not all testShortURLEntities saved")
	suite.Equal(4, len(suite.repository.storage), "not all testShortURLEntities saved")
	suite.Equal(3, len(suite.repository.storageByUserID[suite.testUserIDFirst]), "not expected count of shortURLEntities for testUserIDFirst")
	suite.Equal(1, len(suite.repository.storageByUserID[suite.testUserIDSecond]), "not expected count of shortURLEntities for testUserIDSecond")
}

func (suite *InMemoryRepositoryTestSuite) TestGetShortURLsByUserID_success() {
	shortURLEntities, err := suite.repository.GetShortURLsByUserID(context.Background(), suite.testUserIDFirst)
	if err != nil {
		suite.Error(err, "unexpected error when get shortULREntities by userID")
	}

	suite.Equal(1, len(shortURLEntities), "not expected count of shortURLEntities")
}

func (suite *InMemoryRepositoryTestSuite) TestGetShortURLsByUserID_not_existed_user() {
	shortURLEntities, err := suite.repository.GetShortURLsByUserID(context.Background(), uuid.NewString())
	if err != nil {
		suite.Error(err, "unexpected error when get shortULREntities by userID")
	}

	suite.Equal(0, len(shortURLEntities), "not expected count of shortURLEntities")
}

func (suite *InMemoryRepositoryTestSuite) TestDeleteShortURLsByShortURIs_success() {
	testCtx := context.WithValue(context.Background(), middleware.UserIDContextKey, suite.testUserIDFirst)
	err := suite.repository.DeleteShortURLsByShortURIs(testCtx, []string{suite.testShortURLFirst.ShortURI})
	if err != nil {
		suite.Error(err, "unexpected error when delete shortURLs by shortURIs")
	}

	suite.Equal(true, suite.testShortURLFirst.Deleted, "testShortURLFirst must be deleted")
}

func (suite *InMemoryRepositoryTestSuite) TestDeleteShortURLsByShortURIs_not_expected_user() {
	testCtx := context.WithValue(context.Background(), middleware.UserIDContextKey, suite.testUserIDSecond)
	err := suite.repository.DeleteShortURLsByShortURIs(testCtx, []string{suite.testShortURLFirst.ShortURI})
	if err != nil {
		suite.Error(err, "unexpected error when delete shortURLs by shortURIs")
	}

	suite.Equal(false, suite.testShortURLFirst.Deleted, "testShortURLFirst must be deleted")
}

func TestInMemoryRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(InMemoryRepositoryTestSuite))
}
