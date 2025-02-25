package usecase

import (
	"context"
	"errors"
	"github.com/vkhrushchev/urlshortener/internal/common"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/vkhrushchev/urlshortener/internal/app/domain"
	"github.com/vkhrushchev/urlshortener/internal/app/entity"
	"github.com/vkhrushchev/urlshortener/internal/app/repository"
	mock_usecase "github.com/vkhrushchev/urlshortener/internal/app/usecase/mocks"
	"github.com/vkhrushchev/urlshortener/internal/util"
)

type CreateShortURLUseCaseTestSuite struct {
	suite.Suite
	repositoryMock *mock_usecase.MockshortURLRepository
	useCase        *CreateShortURLUseCase
}

func (suite *CreateShortURLUseCaseTestSuite) SetupTest() {
	mockCtrl := gomock.NewController(suite.T())
	suite.repositoryMock = mock_usecase.NewMockshortURLRepository(mockCtrl)

	suite.useCase = NewCreateShortURLUseCase(suite.repositoryMock)
}

func (suite *CreateShortURLUseCaseTestSuite) TestCreateShortURL_success() {
	testUserID := uuid.NewString()
	testShortURLEntity := &entity.ShortURLEntity{
		UUID:     uuid.NewString(),
		ShortURI: "abc",
		LongURL:  "https://ya.ru",
		UserID:   testUserID,
		Deleted:  false,
	}

	suite.repositoryMock.EXPECT().
		SaveShortURL(gomock.Any(), gomock.Any()).
		Return(testShortURLEntity, nil)

	testCtx := context.WithValue(context.Background(), common.UserIDContextKey, testUserID)
	shortURLDomain, err := suite.useCase.CreateShortURL(testCtx, "https://ya.ru")
	if err != nil {
		log.Errorw("use_case: error when create shortURL", "error", err)
	}

	suite.NotNil(shortURLDomain, "shortURLDomain can not be nil")
}

func (suite *CreateShortURLUseCaseTestSuite) TestCreateShortURL_conflict() {
	testUserID := uuid.NewString()
	testShortURLEntity := &entity.ShortURLEntity{
		UUID:     uuid.NewString(),
		ShortURI: "abc",
		LongURL:  "https://ya.ru",
		UserID:   testUserID,
		Deleted:  false,
	}

	suite.repositoryMock.EXPECT().
		SaveShortURL(gomock.Any(), gomock.Any()).
		Return(testShortURLEntity, repository.ErrConflict)

	testCtx := context.WithValue(context.Background(), common.UserIDContextKey, uuid.NewString())
	shortURLDomain, err := suite.useCase.CreateShortURL(testCtx, "https://ya.ru")

	suite.NotNilf(err, "err cannot be nil")
	suite.True(errors.Is(err, ErrConflict), "err should be ErrConflict")
	suite.NotNilf(shortURLDomain, "shortURLDomain can not be nil")
	suite.NotEmptyf(shortURLDomain.ShortURI, "shortURLDomain.ShortURI can not be empty")
}

func (suite *CreateShortURLUseCaseTestSuite) TestCreateShortURL_unexpected_error() {
	suite.repositoryMock.EXPECT().
		SaveShortURL(gomock.Any(), gomock.Any()).
		Return(nil, repository.ErrUnexpected)

	testCtx := context.WithValue(context.Background(), common.UserIDContextKey, uuid.NewString())
	_, err := suite.useCase.CreateShortURL(testCtx, "https://ya.ru")

	suite.NotNilf(err, "err cannot be nil")
	suite.True(errors.Is(err, ErrUnexpected), "err should be ErrUnexpected")
}

func (suite *CreateShortURLUseCaseTestSuite) TestCreateShortURLBatch_success() {
	testUserID := uuid.NewString()

	suite.repositoryMock.EXPECT().
		SaveShortURLs(gomock.Any(), gomock.Any()).
		Return(
			[]entity.ShortURLEntity{
				{
					UUID:     uuid.NewString(),
					ShortURI: util.RandStringRunes(10),
					LongURL:  "https://ya.ru",
					UserID:   testUserID,
					Deleted:  false,
				},
				{
					UUID:     uuid.NewString(),
					ShortURI: util.RandStringRunes(10),
					LongURL:  "https://mail.ru",
					UserID:   testUserID,
					Deleted:  false,
				},
			},
			nil,
		)

	testCtx := context.WithValue(context.Background(), common.UserIDContextKey, testUserID)
	testCreateShortURLBatchDomains := []domain.CreateShortURLBatchDomain{
		{
			CorrelationUUID: uuid.NewString(),
			LongURL:         "https://ya.ru",
		},
		{
			CorrelationUUID: uuid.NewString(),
			LongURL:         "https://mail.ru",
		},
	}

	createShortURLBatchResultDomains, err := suite.useCase.CreateShortURLBatch(testCtx, testCreateShortURLBatchDomains)
	if err != nil {
		log.Errorw("use_case: error when create shortURL batch", "error", err)
	}

	suite.NotNil(createShortURLBatchResultDomains, "createShortURLBatchResultDomains can not be nil")
	suite.Equal(2, len(createShortURLBatchResultDomains))
}

func (suite *CreateShortURLUseCaseTestSuite) TestCreateShortURLBatch_unexpected_error() {
	suite.repositoryMock.EXPECT().
		SaveShortURLs(gomock.Any(), gomock.Any()).
		Return([]entity.ShortURLEntity{}, repository.ErrUnexpected)

	testUserID := uuid.NewString()
	testCtx := context.WithValue(context.Background(), common.UserIDContextKey, testUserID)
	testCreateShortURLBatchDomains := []domain.CreateShortURLBatchDomain{
		{
			CorrelationUUID: uuid.NewString(),
			LongURL:         "https://ya.ru",
		},
		{
			CorrelationUUID: uuid.NewString(),
			LongURL:         "https://mail.ru",
		},
	}

	createShortURLBatchResultDomains, err := suite.useCase.CreateShortURLBatch(testCtx, testCreateShortURLBatchDomains)
	if err != nil {
		log.Errorw("use_case: error when create shortURL batch", "error", err)
	}

	suite.NotNilf(err, "err cannot be nil")
	suite.Nil(createShortURLBatchResultDomains, "createShortURLBatchResultDomains must be nil")
	suite.True(errors.Is(err, ErrUnexpected), "err should be ErrUnexpected")
}

func TestCreateShortURLUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(CreateShortURLUseCaseTestSuite))
}

type GetShortURLUseCaseTestSuite struct {
	suite.Suite
	repositoryMock *mock_usecase.MockshortURLRepository
	useCase        *GetShortURLUseCase
}

func (suite *GetShortURLUseCaseTestSuite) SetupTest() {
	mockCtrl := gomock.NewController(suite.T())
	suite.repositoryMock = mock_usecase.NewMockshortURLRepository(mockCtrl)

	suite.useCase = NewGetShortURLUseCase(suite.repositoryMock)
}

func (suite *GetShortURLUseCaseTestSuite) TestGetShortURL_success() {
	testUserID := uuid.NewString()
	testShortURLEntity := entity.ShortURLEntity{
		UUID:     uuid.NewString(),
		ShortURI: "abc",
		LongURL:  "https://ya.ru",
		UserID:   testUserID,
		Deleted:  false,
	}

	suite.repositoryMock.EXPECT().
		GetShortURLByShortURI(gomock.Any(), gomock.Any()).
		Return(testShortURLEntity, nil)

	testCtx := context.WithValue(context.Background(), common.UserIDContextKey, testUserID)
	shortURLDomain, err := suite.useCase.GetShortURLByShortURI(testCtx, "abc")
	if err != nil {
		log.Errorw("use_case: error when get short url", "error", err)
	}

	suite.NotNilf(shortURLDomain, "shortURLDomain can not be nil")
	suite.Equalf(domain.ShortURLDomain(testShortURLEntity), shortURLDomain, "must be equal")
}

func (suite *GetShortURLUseCaseTestSuite) TestGetShortURL_not_found() {
	suite.repositoryMock.EXPECT().
		GetShortURLByShortURI(gomock.Any(), gomock.Any()).
		Return(entity.ShortURLEntity{}, repository.ErrNotFound)

	testUserID := uuid.NewString()
	testCtx := context.WithValue(context.Background(), common.UserIDContextKey, testUserID)
	_, err := suite.useCase.GetShortURLByShortURI(testCtx, "abc")
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		log.Errorw("use_case: error when get short url", "error", err)
	}

	suite.NotNilf(err, "err cannot be nil")
	suite.True(errors.Is(err, ErrNotFound), "err should be ErrNotFound")
}

func (suite *GetShortURLUseCaseTestSuite) TestGetShortURL_unexpected_error() {
	suite.repositoryMock.EXPECT().
		GetShortURLByShortURI(gomock.Any(), gomock.Any()).
		Return(entity.ShortURLEntity{}, repository.ErrUnexpected)

	testUserID := uuid.NewString()
	testCtx := context.WithValue(context.Background(), common.UserIDContextKey, testUserID)
	_, err := suite.useCase.GetShortURLByShortURI(testCtx, "abc")
	if err != nil && !errors.Is(err, repository.ErrUnexpected) {
		log.Errorw("use_case: error when get short url", "error", err)
	}

	suite.NotNilf(err, "err cannot be nil")
	suite.True(errors.Is(err, ErrUnexpected), "err should be ErrUnexpected")
}

func (suite *GetShortURLUseCaseTestSuite) TestGetShortURLsByUserID_success() {
	testUserID := uuid.NewString()
	testShortURLEntity := entity.ShortURLEntity{
		UUID:     uuid.NewString(),
		ShortURI: "abc",
		LongURL:  "https://ya.ru",
		UserID:   testUserID,
		Deleted:  false,
	}

	suite.repositoryMock.EXPECT().
		GetShortURLsByUserID(gomock.Any(), gomock.Any()).
		Return([]entity.ShortURLEntity{testShortURLEntity}, nil)

	testCtx := context.WithValue(context.Background(), common.UserIDContextKey, testUserID)
	shortURLDomains, err := suite.useCase.GetShortURLsByUserID(testCtx, testUserID)
	if err != nil {
		log.Errorw("use_case: error when get short urls by userID", "error", err)
	}

	suite.NotNilf(shortURLDomains, "shortURLDomains can not be nil")
	suite.Equalf(1, len(shortURLDomains), "len of shortURLDomains must be equal 1")
}

func (suite *GetShortURLUseCaseTestSuite) TestGetShortURLsByUserID_unexpected_error() {
	suite.repositoryMock.EXPECT().
		GetShortURLsByUserID(gomock.Any(), gomock.Any()).
		Return(nil, repository.ErrUnexpected)

	testUserID := uuid.NewString()
	testCtx := context.WithValue(context.Background(), common.UserIDContextKey, testUserID)
	shortURLDomains, err := suite.useCase.GetShortURLsByUserID(testCtx, testUserID)
	if err != nil && !errors.Is(err, repository.ErrUnexpected) {
		suite.Errorf(err, "use_case: error when get short urls by userID")
	}

	suite.NotNilf(err, "err cannot be nil")
	suite.Nilf(shortURLDomains, "shortURLDomains must be nil")
	suite.True(errors.Is(err, ErrUnexpected), "err should be ErrUnexpected")
}

func TestGetShortURLUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(GetShortURLUseCaseTestSuite))
}

type DeleteShortURLUseCaseTestSuite struct {
	suite.Suite
	repositoryMock *mock_usecase.MockshortURLRepository
	useCase        *DeleteShortURLUseCase
}

func (suite *DeleteShortURLUseCaseTestSuite) SetupTest() {
	mockCtrl := gomock.NewController(suite.T())
	suite.repositoryMock = mock_usecase.NewMockshortURLRepository(mockCtrl)

	suite.useCase = NewDeleteShortURLUseCase(suite.repositoryMock)
}

func (suite *DeleteShortURLUseCaseTestSuite) TestDeleteShortURLsByShortURIs_success() {
	suite.repositoryMock.EXPECT().
		DeleteShortURLsByShortURIs(gomock.Any(), gomock.Any()).
		Return(nil)

	testUserID := uuid.NewString()
	testCtx := context.WithValue(context.Background(), common.UserIDContextKey, testUserID)
	err := suite.useCase.DeleteShortURLsByShortURIs(testCtx, []string{"abc"})
	if err != nil {
		suite.Errorf(err, "use_case: error when delete shortURLs by shortURIs")
	}
}

func (suite *DeleteShortURLUseCaseTestSuite) TestDeleteShortURLsByShortURIs_unexpected_error() {
	suite.repositoryMock.EXPECT().
		DeleteShortURLsByShortURIs(gomock.Any(), gomock.Any()).
		Return(repository.ErrUnexpected)

	testUserID := uuid.NewString()
	testCtx := context.WithValue(context.Background(), common.UserIDContextKey, testUserID)
	err := suite.useCase.DeleteShortURLsByShortURIs(testCtx, []string{"abc"})
	if err != nil && !errors.Is(err, ErrUnexpected) {
		suite.Errorf(err, "use_case: error when delete shortURLs by shortURIs")
	}
}

func TestDeleteShortURLUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(DeleteShortURLUseCaseTestSuite))
}

func BenchmarkCreateShortURLUseCase_CreateShortURL(b *testing.B) {
	repo := repository.NewInMemoryShortURLRepository()
	useCase := NewCreateShortURLUseCase(repo)

	testUserID := uuid.NewString()
	testCtx := context.WithValue(context.Background(), common.UserIDContextKey, testUserID)

	for i := 0; i < b.N; i++ {
		_, err := useCase.CreateShortURL(testCtx, "https://ya.ru")
		if err != nil {
			b.Fatal(err)
		}
	}
}

type StatsUseCaseTestSuite struct {
	suite.Suite
	repositoryMock *mock_usecase.MockstatsRepository
	useCase        *StatsUseCase
}

func (suite *StatsUseCaseTestSuite) SetupTest() {
	mockCtrl := gomock.NewController(suite.T())
	suite.repositoryMock = mock_usecase.NewMockstatsRepository(mockCtrl)

	suite.useCase = NewStatsUseCase(suite.repositoryMock)
}

func (suite *StatsUseCaseTestSuite) TestGetStats_success() {
	suite.repositoryMock.EXPECT().
		GetStats(gomock.Any()).
		Return(5, 2, nil)

	testUserID := uuid.NewString()
	testCtx := context.WithValue(context.Background(), common.UserIDContextKey, testUserID)
	urlCount, userCount, err := suite.useCase.GetStats(testCtx)
	if err != nil {
		suite.Errorf(err, "use_case: error when get stats")
	}

	suite.Equal(5, urlCount, "urlCount must be equal 5")
	suite.Equal(2, userCount, "userCount should be equal 2")
}

func (suite *StatsUseCaseTestSuite) TestGetStats_unexpected_error() {
	suite.repositoryMock.EXPECT().
		GetStats(gomock.Any()).
		Return(0, 0, repository.ErrUnexpected)

	testUserID := uuid.NewString()
	testCtx := context.WithValue(context.Background(), common.UserIDContextKey, testUserID)
	urlCount, userCount, err := suite.useCase.GetStats(testCtx)
	if err != nil && !errors.Is(err, repository.ErrUnexpected) {
		suite.Errorf(err, "use_case: unexpected error when get stats")
	}

	suite.Equal(0, urlCount, "urlCount must be equal 0")
	suite.Equal(0, userCount, "userCount should be equal 0")
	suite.True(errors.Is(err, ErrUnexpected), "err should be ErrUnexpected")
}

func TestStatsUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(StatsUseCaseTestSuite))
}
