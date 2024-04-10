package category

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	pgerr "github.com/jackc/pgerrcode"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"testing"
)

type CreateTestSuite struct {
	suite.Suite

	// Вспомогательные параметры
	ctx        context.Context
	logger     log.Logger
	repository Repository

	// Входные параметры
	create   dto.CreateCategory
	category dto.Category

	// Служебные параметры
	db   *sqlx.DB
	mock sqlmock.Sqlmock
}

func convertArgs(args []any) []driver.Value {
	converted := make([]driver.Value, len(args))

	for i, arg := range args {
		converted[i] = arg
	}

	return converted
}

func TestSuiteCreate(t *testing.T) {
	suite.Run(t, &CreateTestSuite{})
}

func (s *CreateTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.logger = log.NewNullLogger()
}

func (s *CreateTestSuite) BeforeTest(_, _ string) {
	db, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
	)
	s.NoError(err)

	s.setupDatabase(db).setupMock(mock).setupRepository()

	// Входные значения по-умолчанию
	s.setupCreate("Категория").
		setupCategory(1, "Категория")
}

func (s *CreateTestSuite) setupDatabase(
	db *sql.DB,
) *CreateTestSuite {

	wrappedDb := sqlx.NewDb(db, "sqlmock")

	s.db = wrappedDb

	return s
}

func (s *CreateTestSuite) setupMock(
	mock sqlmock.Sqlmock,
) *CreateTestSuite {

	s.mock = mock

	return s
}

func (s *CreateTestSuite) setupRepository() {
	s.repository = New(s.db, s.logger)
}

func (s *CreateTestSuite) setupCreate(
	name string,
) *CreateTestSuite {

	s.create = dto.CreateCategory{
		Name: name,
	}

	return s
}

func (s *CreateTestSuite) setupCategory(
	id int,
	name string,
) *CreateTestSuite {

	s.category = dto.Category{
		ID:   id,
		Name: name,
	}

	return s
}

func (s *CreateTestSuite) TestSuccessful() {
	{
		query, args, err := sq.
			Insert("category").
			Columns("name").
			Values(s.category.Name).
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

	categoryId, err := s.repository.Create(s.ctx, s.create)

	s.NoError(err)
	s.Equal(1, categoryId)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *CreateTestSuite) TestFailed() {
	s.category.ID = 0

	const (
		expectedErrorMsg              = "unknown error on creating category"
		expectedAlreadyExistsErrorMsg = "category already exists"
	)

	testCases := []struct {
		testName              string
		expectedQueryError    error
		expectedQueryErrorMsg string
	}{
		{
			testName:              "Already exists",
			expectedQueryError:    &pq.Error{Code: pgerr.UniqueViolation},
			expectedQueryErrorMsg: expectedAlreadyExistsErrorMsg,
		},
		{
			testName:              "Unknown database error",
			expectedQueryError:    &pq.Error{Code: expectedErrorMsg},
			expectedQueryErrorMsg: expectedErrorMsg,
		},
		{
			testName:              "Unknown error",
			expectedQueryError:    errors.New(expectedErrorMsg),
			expectedQueryErrorMsg: expectedErrorMsg,
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.testName, func() {
			{
				query, args, err := sq.
					Insert("category").
					Columns("name").
					Values(s.create.Name).
					Suffix("RETURNING id").
					PlaceholderFormat(sq.Dollar).
					ToSql()

				s.NoError(err)

				s.mock.
					ExpectQuery(query).
					WithArgs(convertArgs(args)...).
					WillReturnError(
						testCase.expectedQueryError,
					)
			}

			categoryId, err := s.repository.Create(s.ctx, s.create)

			s.NotNil(err)
			s.Equal(err.Error(), testCase.expectedQueryErrorMsg)
			s.Equal(0, categoryId)
			s.NoError(s.mock.ExpectationsWereMet())
		})
	}
}
