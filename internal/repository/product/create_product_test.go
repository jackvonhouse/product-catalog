package product

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
	create   dto.CreateProduct
	product  dto.Product
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
	s.setupCreate("Продукт").
		setupCategory(1, "Категория").
		setupProduct(1, "Продукт")
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

	s.create = dto.CreateProduct{
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

func (s *CreateTestSuite) setupProduct(
	id int,
	name string,
) *CreateTestSuite {

	s.product = dto.Product{
		ID:   id,
		Name: name,
	}

	return s
}

func (s *CreateTestSuite) TestSuccessful() {
	{
		s.mock.ExpectBegin().WillReturnError(nil)
	}

	{
		query, args, err := sq.
			Insert("product").
			Columns("name").
			Values(s.create.Name).
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
					AddRow(s.product.ID),
			)
	}

	{
		query, args, err := sq.
			Insert("product_of_category").
			Columns("product_id", "category_id").
			Values(s.product.ID, s.category.ID).
			Suffix("RETURNING product_id").
			PlaceholderFormat(sq.Dollar).
			ToSql()

		s.NoError(err)

		s.mock.
			ExpectQuery(query).
			WithArgs(convertArgs(args)...).
			WillReturnRows(
				s.mock.
					NewRows([]string{"id"}).
					AddRow(1),
			)
	}

	{
		s.mock.ExpectCommit().WillReturnError(nil)
	}

	productId, err := s.repository.Create(s.ctx, s.create, s.category)

	s.NoError(err)
	s.Equal(1, productId)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *CreateTestSuite) TestBeginFailed() {
	const (
		expectedBeginErrorMsg = "unknown error on begin"
		expectedErrorMsg      = "unknown error on creating product"
	)

	{
		s.mock.
			ExpectBegin().
			WillReturnError(
				errors.New(expectedBeginErrorMsg),
			)
	}

	productId, err := s.repository.Create(s.ctx, s.create, s.category)

	s.NotNil(err)
	s.Equal(err.Error(), expectedErrorMsg)
	s.Equal(0, productId)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *CreateTestSuite) TestCommitFailed() {
	s.product.ID = 0

	const (
		expectedCommitErrorMsg = "unknown error on commit"
		expectedErrorMsg       = "unknown error on creating product"
	)

	{
		s.mock.ExpectBegin().WillReturnError(nil)
	}

	{
		query, args, err := sq.
			Insert("product").
			Columns("name").
			Values(s.create.Name).
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
					AddRow(s.product.ID),
			)
	}

	{
		query, args, err := sq.
			Insert("product_of_category").
			Columns("product_id", "category_id").
			Values(s.product.ID, s.category.ID).
			Suffix("RETURNING product_id").
			PlaceholderFormat(sq.Dollar).
			ToSql()

		s.NoError(err)

		s.mock.
			ExpectQuery(query).
			WithArgs(convertArgs(args)...).
			WillReturnRows(
				s.mock.
					NewRows([]string{"id"}).
					AddRow(0),
			)
	}

	{
		s.mock.ExpectCommit().WillReturnError(
			errors.New(expectedCommitErrorMsg),
		)
	}

	productId, err := s.repository.Create(s.ctx, s.create, s.category)

	s.NotNil(err)
	s.Equal(err.Error(), expectedErrorMsg)
	s.Equal(0, productId)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *CreateTestSuite) TestRollbackFailed() {
	s.product.ID = 0

	const (
		expectedQueryErrorMsg = "unknown error on query"
		expectedRollbackError = "unknown error on rollback"
		expectedErrorMsg      = "unknown error on creating product"
	)

	{
		s.mock.
			ExpectBegin().WillReturnError(nil)
	}

	{
		query, args, err := sq.
			Insert("product").
			Columns("name").
			Values(s.create.Name).
			Suffix("RETURNING id").
			PlaceholderFormat(sq.Dollar).
			ToSql()

		s.NoError(err)

		s.mock.
			ExpectQuery(query).
			WithArgs(convertArgs(args)...).
			WillReturnError(errors.New(expectedQueryErrorMsg))
	}

	{
		s.mock.ExpectRollback().
			WillReturnError(
				errors.New(expectedRollbackError),
			)
	}

	productId, err := s.repository.Create(s.ctx, s.create, s.category)

	s.NotNil(err)
	s.Equal(err.Error(), expectedErrorMsg)
	s.Equal(0, productId)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *CreateTestSuite) TestCreateProductFailed() {
	s.product.ID = 0

	const (
		expectedErrorMsg              = "unknown error on creating product"
		expectedAlreadyExistsErrorMsg = "product already exists"
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
				s.mock.
					ExpectBegin().WillReturnError(nil)
			}

			{
				query, args, err := sq.
					Insert("product").
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

			{
				s.mock.ExpectRollback().WillReturnError(nil)
			}

			productId, err := s.repository.Create(s.ctx, s.create, s.category)

			s.NotNil(err)
			s.Equal(err.Error(), testCase.expectedQueryErrorMsg)
			s.Equal(0, productId)
			s.NoError(s.mock.ExpectationsWereMet())
		})
	}
}

func (s *CreateTestSuite) TestFailedToAttachProductToCategory() {
	const (
		errProductAlreadyExistsMsg = "product in category already exists"
		errProductNotFoundMsg      = "product not found"
		errCategoryNotFoundMsg     = "category not found"
		errUnknownMsg              = "unknown error on attaching product to category"
	)

	testCases := []struct {
		testName         string
		expectedError    error
		expectedErrorMsg string
	}{
		{
			testName:         "Already exists",
			expectedError:    &pq.Error{Code: pgerr.UniqueViolation},
			expectedErrorMsg: errProductAlreadyExistsMsg,
		},
		{
			testName:         "Product not found",
			expectedError:    &pq.Error{Code: pgerr.ForeignKeyViolation, Detail: `"product"`},
			expectedErrorMsg: errProductNotFoundMsg,
		},
		{
			testName:         "Category not found",
			expectedError:    &pq.Error{Code: pgerr.ForeignKeyViolation, Detail: `"category"`},
			expectedErrorMsg: errCategoryNotFoundMsg,
		},
		{
			testName:         "Unknown database error",
			expectedError:    &pq.Error{Code: errUnknownMsg},
			expectedErrorMsg: errUnknownMsg,
		},
		{
			testName:         "Unknown error",
			expectedError:    errors.New(errUnknownMsg),
			expectedErrorMsg: errUnknownMsg,
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.testName, func() {
			{
				s.mock.
					ExpectBegin().
					WillReturnError(nil)
			}

			{
				query, args, err := sq.
					Insert("product").
					Columns("name").
					Values(s.create.Name).
					Suffix("RETURNING id").
					PlaceholderFormat(sq.Dollar).
					ToSql()

				s.NoError(err)

				s.mock.
					ExpectQuery(query).
					WithArgs(convertArgs(args)...).
					WillReturnRows(
						sqlmock.NewRows([]string{"id"}).
							AddRow(s.product.ID),
					)
			}

			{
				query, args, err := sq.
					Insert("product_of_category").
					Columns("product_id", "category_id").
					Values(s.product.ID, s.category.ID).
					Suffix("RETURNING product_id").
					PlaceholderFormat(sq.Dollar).
					ToSql()

				s.NoError(err)

				s.mock.
					ExpectQuery(query).
					WithArgs(convertArgs(args)...).
					WillReturnError(
						testCase.expectedError,
					)
			}

			{
				s.mock.ExpectRollback().WillReturnError(nil)
			}

			productId, err := s.repository.Create(s.ctx, s.create, s.category)

			s.NotNil(err)
			s.Equal(err.Error(), testCase.expectedErrorMsg)
			s.Equal(0, productId)
			s.NoError(s.mock.ExpectationsWereMet())
		})
	}
}

func (s *CreateTestSuite) TestFailedRollbackAfterAttachProductToCategory() {
	const (
		errUnknownMsg       = "unknown error on creating product"
		errUnknownAttachMsg = "unknown error on attaching product to category"
	)

	{
		s.mock.
			ExpectBegin().
			WillReturnError(nil)
	}

	{
		query, args, err := sq.
			Insert("product").
			Columns("name").
			Values(s.create.Name).
			Suffix("RETURNING id").
			PlaceholderFormat(sq.Dollar).
			ToSql()

		s.NoError(err)

		s.mock.
			ExpectQuery(query).
			WithArgs(convertArgs(args)...).
			WillReturnRows(
				sqlmock.NewRows([]string{"id"}).
					AddRow(s.product.ID),
			)
	}

	{
		query, args, err := sq.
			Insert("product_of_category").
			Columns("product_id", "category_id").
			Values(s.product.ID, s.category.ID).
			Suffix("RETURNING product_id").
			PlaceholderFormat(sq.Dollar).
			ToSql()

		s.NoError(err)

		s.mock.
			ExpectQuery(query).
			WithArgs(convertArgs(args)...).
			WillReturnError(
				errors.New(errUnknownAttachMsg),
			)
	}

	{
		s.mock.
			ExpectRollback().
			WillReturnError(
				errors.New(errUnknownMsg),
			)
	}

	productId, err := s.repository.Create(s.ctx, s.create, s.category)

	s.NotNil(err)
	s.Equal(err.Error(), errUnknownMsg)
	s.Equal(0, productId)
	s.NoError(s.mock.ExpectationsWereMet())
}
