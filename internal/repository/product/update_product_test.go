package product

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
	update   dto.UpdateProduct
	product  dto.Product
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
	s.setupUpdate("Продукт").
		setupCategory(1, "Категория").
		setupProduct(1, "Продукт")
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

	s.update = dto.UpdateProduct{
		ID:            1,
		Name:          name,
		OldCategoryId: 1,
		NewCategoryId: 2,
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

func (s *UpdateTestSuite) setupProduct(
	id int,
	name string,
) *UpdateTestSuite {

	s.product = dto.Product{
		ID:   id,
		Name: name,
	}

	return s
}

func (s *UpdateTestSuite) TestSuccessful() {
	{
		s.mock.ExpectBegin().WillReturnError(nil)
	}

	{
		query, args, err := sq.
			Update("product").
			SetMap(map[string]any{
				"name": s.update.Name,
			}).
			Where(sq.Eq{"id": s.update.OldCategoryId}).
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
			Update("product_of_category").
			SetMap(map[string]any{
				"category_id": s.category.ID,
			}).
			Where(sq.And{
				sq.Eq{"category_id": s.update.OldCategoryId},
				sq.Eq{"product_id": s.product.ID},
			}).
			PlaceholderFormat(sq.Dollar).
			ToSql()

		s.NoError(err)

		s.mock.
			ExpectExec(query).
			WithArgs(convertArgs(args)...).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}

	{
		s.mock.ExpectCommit().WillReturnError(nil)
	}

	productId, err := s.repository.Update(s.ctx, s.update, s.product, s.category)

	s.NoError(err)
	s.Equal(1, productId)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *UpdateTestSuite) TestWithoutCategoryUpdateSuccessful() {
	s.update.OldCategoryId = s.update.NewCategoryId
	s.category.ID = s.update.OldCategoryId

	{
		s.mock.ExpectBegin().WillReturnError(nil)
	}

	{
		query, args, err := sq.
			Update("product").
			SetMap(map[string]any{
				"name": s.update.Name,
			}).
			Where(sq.Eq{"id": s.product.ID}).
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
		s.mock.ExpectCommit().WillReturnError(nil)
	}

	productId, err := s.repository.Update(s.ctx, s.update, s.product, s.category)

	s.NoError(err)
	s.Equal(1, productId)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *UpdateTestSuite) TestBeginFailed() {
	const (
		expectedBeginErrorMsg = "unknown error on begin"
		expectedErrorMsg      = "unknown error on updating product"
	)

	{
		s.mock.
			ExpectBegin().
			WillReturnError(
				errors.New(expectedBeginErrorMsg),
			)
	}

	productId, err := s.repository.Update(s.ctx, s.update, s.product, s.category)

	s.NotNil(err)
	s.Equal(err.Error(), expectedErrorMsg)
	s.Equal(0, productId)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *UpdateTestSuite) TestUpdateProductFailed() {
	const (
		expectedErrorMsg              = "unknown error on updating product"
		expectedInternalErrorMsg      = "unknown database error"
		expectedNotFoundErrorMsg      = "product not found"
		expectedAlreadyExistsErrorMsg = "product already exists"
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
			testName: "Product not found",
			expectedError: &pq.Error{
				Code:    pgerr.ForeignKeyViolation,
				Message: expectedNotFoundErrorMsg,
				Detail:  `"product"`,
			},
			expectedErrorMsg: expectedNotFoundErrorMsg,
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
				s.mock.ExpectBegin().WillReturnError(nil)
			}

			{
				query, args, err := sq.
					Update("product").
					SetMap(map[string]any{
						"name": s.update.Name,
					}).
					Where(sq.Eq{"id": s.product.ID}).
					Suffix("RETURNING id").
					PlaceholderFormat(sq.Dollar).
					ToSql()

				s.NoError(err)

				s.mock.
					ExpectQuery(query).
					WithArgs(convertArgs(args)...).
					WillReturnError(testCase.expectedError)
			}

			{
				s.mock.ExpectRollback().WillReturnError(nil)
			}

			productId, err := s.repository.Update(s.ctx, s.update, s.product, s.category)

			s.NotNil(err)
			s.Equal(err.Error(), testCase.expectedErrorMsg)
			s.Equal(0, productId)
			s.NoError(s.mock.ExpectationsWereMet())
		})
	}
}

func (s *UpdateTestSuite) TestUpdateProductInCategoryFailed() {
	const (
		expectedErrorMsg                 = "unknown error on updating product in category"
		expectedInternalErrorMsg         = "unknown database error"
		expectedProductNotFoundErrorMsg  = "product not found"
		expectedCategoryNotFoundErrorMsg = "category not found"
		expectedNotFoundErrorMsg         = "product in category not found"
		expectedAlreadyExistsErrorMsg    = "product in category already exists"
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
			testName: "Product not found",
			expectedError: &pq.Error{
				Code:    pgerr.ForeignKeyViolation,
				Message: expectedProductNotFoundErrorMsg,
				Detail:  `"product"`,
			},
			expectedErrorMsg: expectedProductNotFoundErrorMsg,
		},
		{
			testName: "Category not found",
			expectedError: &pq.Error{
				Code:    pgerr.ForeignKeyViolation,
				Message: expectedCategoryNotFoundErrorMsg,
				Detail:  `"category"`,
			},
			expectedErrorMsg: expectedCategoryNotFoundErrorMsg,
		},
		{
			testName:         "Unknown database error",
			expectedError:    &pq.Error{Message: expectedInternalErrorMsg},
			expectedErrorMsg: expectedErrorMsg,
		},
		{
			testName:         "Unknown database error",
			expectedError:    errors.New(expectedInternalErrorMsg),
			expectedErrorMsg: expectedErrorMsg,
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.testName, func() {
			{
				s.mock.ExpectBegin().WillReturnError(nil)
			}

			{
				query, args, err := sq.
					Update("product").
					SetMap(map[string]any{
						"name": s.update.Name,
					}).
					Where(sq.Eq{"id": s.product.ID}).
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
					Update("product_of_category").
					SetMap(map[string]any{
						"category_id": s.category.ID,
					}).
					Where(sq.And{
						sq.Eq{"category_id": s.update.OldCategoryId},
						sq.Eq{"product_id": s.product.ID},
					}).
					PlaceholderFormat(sq.Dollar).
					ToSql()

				s.NoError(err)

				s.mock.
					ExpectExec(query).
					WithArgs(convertArgs(args)...).
					WillReturnError(testCase.expectedError)
			}

			{
				s.mock.ExpectRollback().WillReturnError(nil)
			}

			productId, err := s.repository.Update(s.ctx, s.update, s.product, s.category)

			s.NotNil(err)
			s.Equal(err.Error(), testCase.expectedErrorMsg)
			s.Equal(0, productId)
			s.NoError(s.mock.ExpectationsWereMet())
		})
	}
}

func (s *UpdateTestSuite) TestUpdateProductRollbackFailed() {
	const (
		expectedQueryErrorMsg    = "unknown error on updating product"
		expectedRollbackErrorMsg = "unknown error on rollback"
	)

	{
		s.mock.ExpectBegin().WillReturnError(nil)
	}

	{
		query, args, err := sq.
			Update("product").
			SetMap(map[string]any{
				"name": s.update.Name,
			}).
			Where(sq.Eq{"id": s.product.ID}).
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
		s.mock.ExpectRollback().WillReturnError(
			errors.New(expectedRollbackErrorMsg),
		)
	}

	productId, err := s.repository.Update(s.ctx, s.update, s.product, s.category)

	s.NotNil(err)
	s.Equal(err.Error(), expectedQueryErrorMsg)
	s.Equal(0, productId)
	s.NoError(s.mock.ExpectationsWereMet())

}

func (s *UpdateTestSuite) TestUpdateProductInCategoryRollbackFailed() {
	const (
		expectedQueryErrorMsg    = "unknown error on updating product"
		expectedRollbackErrorMsg = "unknown error on rollback"
	)

	{
		s.mock.ExpectBegin().WillReturnError(nil)
	}

	{
		query, args, err := sq.
			Update("product").
			SetMap(map[string]any{
				"name": s.update.Name,
			}).
			Where(sq.Eq{"id": s.product.ID}).
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
			Update("product_of_category").
			SetMap(map[string]any{
				"category_id": s.category.ID,
			}).
			Where(sq.And{
				sq.Eq{"category_id": s.update.OldCategoryId},
				sq.Eq{"product_id": s.product.ID},
			}).
			PlaceholderFormat(sq.Dollar).
			ToSql()

		s.NoError(err)

		s.mock.
			ExpectExec(query).
			WithArgs(convertArgs(args)...).
			WillReturnError(errors.New(expectedQueryErrorMsg))
	}

	{
		s.mock.ExpectRollback().WillReturnError(
			errors.New(expectedRollbackErrorMsg),
		)
	}

	productId, err := s.repository.Update(s.ctx, s.update, s.product, s.category)

	s.NotNil(err)
	s.Equal(err.Error(), expectedQueryErrorMsg)
	s.Equal(0, productId)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *UpdateTestSuite) TestWithoutCategoryUpdateCommitFailed() {
	const (
		expectedCommitErrorMsg = "unknown error on commit"
		expectedErrorMsg       = "unknown error on updating product"
	)

	s.update.OldCategoryId = s.update.NewCategoryId
	s.category.ID = s.update.OldCategoryId

	{
		s.mock.ExpectBegin().WillReturnError(nil)
	}

	{
		query, args, err := sq.
			Update("product").
			SetMap(map[string]any{
				"name": s.update.Name,
			}).
			Where(sq.Eq{"id": s.product.ID}).
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
		s.mock.ExpectCommit().WillReturnError(
			errors.New(expectedCommitErrorMsg),
		)
	}

	productId, err := s.repository.Update(s.ctx, s.update, s.product, s.category)

	s.NotNil(err)
	s.Equal(err.Error(), expectedErrorMsg)
	s.Equal(1, productId)
	s.NoError(s.mock.ExpectationsWereMet())
}
