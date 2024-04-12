package category

import (
	"context"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"testing"
)

type DeleteTestSuite struct {
	suite.Suite

	// Вспомогательные параметры
	ctx        context.Context
	logger     log.Logger
	repository Repository

	// Входные параметры
	category dto.Category

	// Служебные параметры
	db   *sqlx.DB
	mock sqlmock.Sqlmock
}

func TestSuiteDelete(t *testing.T) {
	suite.Run(t, &DeleteTestSuite{})
}

func (s *DeleteTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.logger = log.NewNullLogger()
}

func (s *DeleteTestSuite) BeforeTest(_, _ string) {
	db, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
	)
	s.NoError(err)

	s.setupDatabase(db).setupMock(mock).setupRepository()

	// Входные значения по-умолчанию
	s.setupCategory(1)
}

func (s *DeleteTestSuite) setupDatabase(
	db *sql.DB,
) *DeleteTestSuite {

	wrappedDb := sqlx.NewDb(db, "sqlmock")

	s.db = wrappedDb

	return s
}

func (s *DeleteTestSuite) setupMock(
	mock sqlmock.Sqlmock,
) *DeleteTestSuite {

	s.mock = mock

	return s
}

func (s *DeleteTestSuite) setupRepository() {
	s.repository = New(s.db, s.logger)
}

func (s *DeleteTestSuite) setupCategory(
	id int,
) *DeleteTestSuite {

	s.category = dto.Category{
		ID: id,
	}

	return s
}

func (s *DeleteTestSuite) TestSuccessful() {
	{
		query, args, err := sq.
			Delete("category").
			Where(sq.Eq{"id": s.category.ID}).
			Suffix("RETURNING id").
			PlaceholderFormat(sq.Dollar).
			ToSql()

		s.NoError(err)

		s.mock.
			ExpectQuery(query).
			WithArgs(convertArgs(args)...).
			WillReturnRows(
				s.mock.
					NewRows([]string{"id"}).
					AddRow(s.category.ID),
			)
	}

	categoryId, err := s.repository.Delete(s.ctx, s.category)

	s.NoError(err)
	s.Equal(1, categoryId)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *DeleteTestSuite) TestFailed() {
	const (
		expectedErrorMsg         = "unknown error on deleting category"
		expectedInternalErrorMsg = "unknown error"
		expectedNotFoundErrorMsg = "category not found"
	)

	testCases := []struct {
		testName         string
		expectedError    error
		expectedErrorMsg string
	}{
		{
			testName:         "Not found",
			expectedError:    sql.ErrNoRows,
			expectedErrorMsg: expectedNotFoundErrorMsg,
		},
		{
			testName:         "Unknown error",
			expectedError:    errors.New(expectedInternalErrorMsg),
			expectedErrorMsg: expectedErrorMsg,
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.testName, func() {
			query, args, err := sq.
				Delete("category").
				Where(sq.Eq{"id": s.category.ID}).
				Suffix("RETURNING id").
				PlaceholderFormat(sq.Dollar).
				ToSql()

			s.NoError(err)

			s.mock.
				ExpectQuery(query).
				WithArgs(convertArgs(args)...).
				WillReturnError(testCase.expectedError)

			categoryId, err := s.repository.Delete(s.ctx, s.category)

			s.NotNil(err)
			s.Equal(err.Error(), testCase.expectedErrorMsg)
			s.Equal(0, categoryId)
			s.NoError(s.mock.ExpectationsWereMet())
		})
	}
}
