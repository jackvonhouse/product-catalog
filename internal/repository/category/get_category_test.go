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

type GetTestSuite struct {
	suite.Suite

	// Вспомогательные параметры
	ctx        context.Context
	logger     log.Logger
	repository Repository

	// Входные параметры
	get        dto.GetCategory
	category   dto.Category
	categories []dto.Category

	// Служебные параметры
	db   *sqlx.DB
	mock sqlmock.Sqlmock
}

func TestSuiteGet(t *testing.T) {
	suite.Run(t, &GetTestSuite{})
}

func (s *GetTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.logger = log.NewNullLogger()
}

func (s *GetTestSuite) BeforeTest(_, _ string) {
	db, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
	)
	s.NoError(err)

	s.setupDatabase(db).setupMock(mock).setupRepository()

	// Входные значения по-умолчанию
	s.setupGet(0, 10).
		setupCategory(1, "Категория").
		setupCategories(1, "Категория")
}

func (s *GetTestSuite) setupDatabase(
	db *sql.DB,
) *GetTestSuite {

	wrappedDb := sqlx.NewDb(db, "sqlmock")

	//wrappedDb, _ := sqlx.ConnectContext(s.ctx, "postgres", "user=catalog-admin password=catalog-admin-password dbname=catalog host=127.0.0.1 port=5432 sslmode=disable")

	s.db = wrappedDb

	return s
}

func (s *GetTestSuite) setupMock(
	mock sqlmock.Sqlmock,
) *GetTestSuite {

	s.mock = mock

	return s
}

func (s *GetTestSuite) setupRepository() {
	s.repository = New(s.db, s.logger)
}

func (s *GetTestSuite) setupGet(
	offset, limit int,
) *GetTestSuite {

	s.get = dto.GetCategory{
		Limit:  limit,
		Offset: offset,
	}

	return s
}

func (s *GetTestSuite) setupCategory(
	id int,
	name string,
) *GetTestSuite {

	s.category = dto.Category{
		ID:   id,
		Name: name,
	}

	return s
}

func (s *GetTestSuite) setupCategories(
	id int,
	name string,
) *GetTestSuite {

	s.categories = []dto.Category{
		{
			ID:   id,
			Name: name,
		},
	}

	return s
}

func (s *GetTestSuite) TestGetSuccessful() {
	{
		query, args, err := sq.
			Select("*").
			From("category").
			OrderBy("id ASC").
			Offset(uint64(s.get.Offset)).
			Limit(uint64(s.get.Limit)).
			PlaceholderFormat(sq.Dollar).
			ToSql()

		s.NoError(err)

		s.mock.
			ExpectQuery(query).
			WithArgs(convertArgs(args)...).
			WillReturnRows(
				s.mock.
					NewRows([]string{"id", "name"}).
					AddRow(
						s.category.ID,
						s.category.Name,
					),
			)
	}

	categories, err := s.repository.Get(s.ctx, s.get)

	s.NoError(err)
	s.Equal(categories, s.categories)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *GetTestSuite) TestGetFailure() {
	const (
		expectedInternalErrorMsg = "unknown database error"
		expectedErrorMsg         = "unknown error on getting categories"
		expectedNotFoundErrorMsg = "categories not found"
	)

	testCases := []struct {
		testName         string
		expectedError    error
		expectedErrorMsg string
	}{
		{
			testName:         "Unknown error",
			expectedError:    errors.New(expectedInternalErrorMsg),
			expectedErrorMsg: expectedErrorMsg,
		},
		{
			testName:         "Not found",
			expectedError:    sql.ErrNoRows,
			expectedErrorMsg: expectedNotFoundErrorMsg,
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.testName, func() {
			{
				query, args, err := sq.
					Select("*").
					From("category").
					OrderBy("id ASC").
					Offset(uint64(s.get.Offset)).
					Limit(uint64(s.get.Limit)).
					PlaceholderFormat(sq.Dollar).
					ToSql()

				s.NoError(err)

				s.mock.
					ExpectQuery(query).
					WithArgs(convertArgs(args)...).
					WillReturnError(testCase.expectedError)
			}

			categories, err := s.repository.Get(s.ctx, s.get)

			s.NotNil(err)
			s.Equal(err.Error(), testCase.expectedErrorMsg)
			s.Equal(categories, []dto.Category{})
			s.NoError(s.mock.ExpectationsWereMet())
		})
	}
}

func (s *GetTestSuite) TestGetByIdSuccessful() {
	{
		query, args, err := sq.
			Select("*").
			From("category").
			Where(sq.Eq{"id": s.category.ID}).
			PlaceholderFormat(sq.Dollar).
			ToSql()

		s.NoError(err)

		s.mock.
			ExpectQuery(query).
			WithArgs(convertArgs(args)...).
			WillReturnRows(
				s.mock.
					NewRows([]string{"id", "name"}).
					AddRow(
						s.category.ID,
						s.category.Name,
					),
			)
	}

	category, err := s.repository.GetById(s.ctx, s.category.ID)

	s.NoError(err)
	s.Equal(category, s.category)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *GetTestSuite) TestGetByIdFailure() {
	const (
		expectedInternalErrorMsg = "unknown database error"
		expectedErrorMsg         = "unknown error on getting category"
		expectedNotFoundErrorMsg = "category not found"
	)

	testCases := []struct {
		testName         string
		expectedError    error
		expectedErrorMsg string
	}{
		{
			testName:         "Unknown error",
			expectedError:    errors.New(expectedInternalErrorMsg),
			expectedErrorMsg: expectedErrorMsg,
		},
		{
			testName:         "Not found",
			expectedError:    sql.ErrNoRows,
			expectedErrorMsg: expectedNotFoundErrorMsg,
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.testName, func() {
			{
				query, args, err := sq.
					Select("*").
					From("category").
					Where(sq.Eq{"id": s.category.ID}).
					PlaceholderFormat(sq.Dollar).
					ToSql()

				s.NoError(err)

				s.mock.
					ExpectQuery(query).
					WithArgs(convertArgs(args)...).
					WillReturnError(testCase.expectedError)
			}

			category, err := s.repository.GetById(s.ctx, s.category.ID)

			s.NotNil(err)
			s.Equal(err.Error(), testCase.expectedErrorMsg)
			s.Equal(category, dto.Category{})
			s.NoError(s.mock.ExpectationsWereMet())
		})
	}
}
