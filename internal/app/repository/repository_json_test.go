package repository

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/vkhrushchev/urlshortener/internal/app/entity"
	"github.com/vkhrushchev/urlshortener/internal/middleware"
)

const TestDataFile = "json_short_url_test_data.json"

type JSONFileShortURLRepositoryTestSuite struct {
	suite.Suite
	repository *JSONFileShortURLRepository
}

func (s *JSONFileShortURLRepositoryTestSuite) SetupTest() {
	repository, err := NewJSONFileShortURLRepository(TestDataFile)
	if err != nil {
		s.Fail("repository: unexpected error when create JSONFileShortURLRepository: %v", err)
	}

	s.repository = repository
}

func (s *JSONFileShortURLRepositoryTestSuite) TearDownTest() {
	err := os.Remove(TestDataFile)
	if err != nil {
		s.Fail("repository: unexpected error when remove test data file for JSONFileShortURLRepository: %v", err)
	}
}

func (s *JSONFileShortURLRepositoryTestSuite) TestSaveShortURL() {
	testUserID := uuid.NewString()
	testShortURL := &entity.ShortURLEntity{
		UUID:     uuid.NewString(),
		ShortURI: "ghi",
		LongURL:  "https://mail.ru",
		UserID:   testUserID,
		Deleted:  false,
	}

	testCtx := context.WithValue(context.Background(), middleware.UserIDContextKey, testUserID)
	savedShortURL, err := s.repository.SaveShortURL(testCtx, testShortURL)
	if err != nil {
		s.Fail("unexpected error when save ShortURLEntity")
	}

	s.NotNil(savedShortURL, "savedShortURL should not be nil")
	s.NotNil(s.repository.storage[testShortURL.ShortURI], "testShortURL should be saved in storage")
	s.NotNil(s.repository.storageByUserID[testUserID], "storageByUserID must be created")
	s.Equal(1, len(s.repository.storageByUserID[testUserID]), "testShortURL must be saved in storageByUserID")
}

func TestJSONFileShortURLRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(JSONFileShortURLRepositoryTestSuite))
}
