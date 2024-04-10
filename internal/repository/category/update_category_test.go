package category

import (
	"context"
	"database/sql"
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

type UpdateTestSuite struct {
	suite.Suite

	// Вспомогательные параметры
	ctx        context.Context
	logger     log.Logger
	repository Repository

	// Входные параметры
	update   dto.UpdateCategory
	category dto.Category

	// Служебные параметры
	db   *sqlx.DB
	mock sqlmock.Sqlmock
}

func TestSuiteUpdate(t *testing.T) {
	suite.Run(t, &UpdateTestSuite{})
}

func (s *UpdateTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.logger = log.NewNullLogger()
}

func (s *UpdateTestSuite) BeforeTest(_, _ string) {
	db, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
	)
	s.NoError(err)

	s.setupDatabase(db).setupMock(mock).setupRepository()

	// Входные значения по-умолчанию
	s.setupUpdate("Категория").
		setupCategory(1, "Категория")
}

func (s *UpdateTestSuite) setupDatabase(
	db *sql.DB,
) *UpdateTestSuite {

	wrappedDb := sqlx.NewDb(db, "sqlmock")

	s.db = wrappedDb

	return s
}

func (s *UpdateTestSuite) setupMock(
	mock sqlmock.Sqlmock,
) *UpdateTestSuite {

	s.mock = mock

	return s
}

func (s *UpdateTestSuite) setupRepository() {
	s.repository = New(s.db, s.logger)
}

func (s *UpdateTestSuite) setupUpdate(
	name string,
) *UpdateTestSuite {

	s.update = dto.UpdateCategory{
		ID:   1,
		Name: name,
	}

	return s
}

func (s *UpdateTestSuite) setupCategory(
	id int,
	name string,
) *UpdateTestSuite {

	s.category = dto.Category{
		ID:   id,
		Name: name,
	}

	return s
}

func (s *UpdateTestSuite) TestSuccessful() {
	{
		query, args, err := sq.
			Update("category").
			SetMap(map[string]any{
				"name": s.update.Name,
			}).
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

	categoryId, err := s.repository.Update(s.ctx, s.update)

	s.NoError(err)
	s.Equal(1, categoryId)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *UpdateTestSuite) TestFailed() {
	const (
		expectedErrorMsg              = "unknown error on updating category"
		expectedInternalErrorMsg      = "unknown database error"
		expectedNotFoundErrorMsg      = "category not found"
		expectedAlreadyExistsErrorMsg = "category already exists"
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
			testName:         "Already exists",
			expectedError:    &pq.Error{Code: pgerr.UniqueViolation, Message: expectedAlreadyExistsErrorMsg},
			expectedErrorMsg: expectedAlreadyExistsErrorMsg,
		},
		{
			testName:         "Unknown database error",
			expectedError:    &pq.Error{Message: expectedInternalErrorMsg},
			expectedErrorMsg: expectedErrorMsg,
		},
		{
			testName:         "Unknown error",
			expectedError:    errors.New(expectedInternalErrorMsg),
			expectedErrorMsg: expectedErrorMsg,
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.testName, func() {
			{
				query, args, err := sq.
					Update("category").
					SetMap(map[string]any{
						"name": s.update.Name,
					}).
					Where(sq.Eq{"id": s.category.ID}).
					Suffix("RETURNING id").
					PlaceholderFormat(sq.Dollar).
					ToSql()

				s.NoError(err)

				s.mock.
					ExpectQuery(query).
					WithArgs(convertArgs(args)...).
					WillReturnError(testCase.expectedError)
			}

			categoryId, err := s.repository.Update(s.ctx, s.update)

			s.NotNil(err)
			s.Equal(err.Error(), testCase.expectedErrorMsg)
			s.Equal(0, categoryId)
			s.NoError(s.mock.ExpectationsWereMet())
		})
	}
}
