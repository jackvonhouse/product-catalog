package product

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"testing"
)

type IntArrayConverter struct{}

func (s IntArrayConverter) ConvertValue(v any) (driver.Value, error) {
	switch x := v.(type) {
	case []int:
		return x, nil
	default:
		return driver.DefaultParameterConverter.ConvertValue(v)
	}
}

type GetTestSuite struct {
	suite.Suite

	// Вспомогательные параметры
	ctx        context.Context
	logger     log.Logger
	repository Repository

	// Входные параметры
	get      dto.GetProduct
	category dto.Category
	product  dto.Product
	products []dto.Product

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
		sqlmock.ValueConverterOption(IntArrayConverter{}),
	)
	s.NoError(err)

	s.setupDatabase(db).setupMock(mock).setupRepository()

	// Входные значения по-умолчанию
	s.setupCategory(1, "Категория").
		setupGet(0, 10).
		setupProduct(1, "Продукт").
		setupProducts(1, "Продукт")
}

func (s *GetTestSuite) setupDatabase(
	db *sql.DB,
) *GetTestSuite {

	wrappedDb := sqlx.NewDb(db, "sqlmock")

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

	s.get = dto.GetProduct{
		Limit:  limit,
		Offset: offset,
	}

	return s
}

func (s *GetTestSuite) setupProduct(
	id int,
	name string,
) *GetTestSuite {

	s.product = dto.Product{
		ID:   id,
		Name: name,
	}

	return s
}

func (s *GetTestSuite) setupProducts(
	id int,
	name string,
) *GetTestSuite {

	s.products = []dto.Product{
		{
			ID:   id,
			Name: name,
		},
	}

	return s
}

func (s *GetTestSuite) setupCategory(
	id int,
	name string,
) *GetTestSuite {

	s.category = dto.Category{
		ID:   1,
		Name: "Категория",
	}

	return s
}

func (s *GetTestSuite) TestGetSuccessful() {
	{
		query, args, err := sq.
			Select("*").
			From("product").
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
						s.product.ID,
						s.product.Name,
					),
			)
	}

	products, err := s.repository.Get(s.ctx, s.get)

	s.NoError(err)
	s.Equal(products, s.products)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *GetTestSuite) TestGetFailure() {
	testCases := []struct {
		testName         string
		expectedError    error
		expectedErrorMsg string
	}{
		{
			testName:         "Unknown database error",
			expectedError:    errors.New("unknown error on getting products"),
			expectedErrorMsg: "unknown error on getting products",
		},
		{
			testName:         "Not found",
			expectedError:    sql.ErrNoRows,
			expectedErrorMsg: "products not found",
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.testName, func() {
			{
				query, args, err := sq.
					Select("*").
					From("product").
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

			products, err := s.repository.Get(s.ctx, s.get)

			s.NotNil(err)
			s.Equal(err.Error(), testCase.expectedErrorMsg)
			s.Equal(products, []dto.Product{})
			s.NoError(s.mock.ExpectationsWereMet())
		})
	}
}

func (s *GetTestSuite) TestGetByCategoryIdSuccessful() {
	{
		query, args, err := sq.
			Select("p.*").
			LeftJoin("product_of_category pc ON pc.product_id = p.id").
			LeftJoin("category c ON c.id = pc.category_id").
			From("product p").
			Where(sq.Eq{"pc.category_id": 1}).
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
						s.product.ID,
						s.product.Name,
					),
			)
	}

	products, err := s.repository.GetByCategoryId(s.ctx, s.get, s.category)

	s.NoError(err)
	s.Equal(products, s.products)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *GetTestSuite) TestGetByCategoryIdFailure() {
	testCases := []struct {
		testName         string
		expectedError    error
		expectedErrorMsg string
	}{
		{
			testName:         "Unknown database error",
			expectedError:    errors.New("unknown error on getting products"),
			expectedErrorMsg: "unknown error on getting products",
		},
		{
			testName:         "Not found",
			expectedError:    sql.ErrNoRows,
			expectedErrorMsg: "products not found",
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.testName, func() {
			{
				query, args, err := sq.
					Select("p.*").
					LeftJoin("product_of_category pc ON pc.product_id = p.id").
					LeftJoin("category c ON c.id = pc.category_id").
					From("product p").
					Where(sq.Eq{"pc.category_id": 1}).
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

			products, err := s.repository.GetByCategoryId(s.ctx, s.get, s.category)

			s.NotNil(err)
			s.Equal(err.Error(), testCase.expectedErrorMsg)
			s.Equal(products, []dto.Product{})
			s.NoError(s.mock.ExpectationsWereMet())
		})
	}
}
func (s *GetTestSuite) TestGetByIdSuccessful() {
	{
		query, args, err := sq.
			Select("*").
			From("product").
			Where(sq.Eq{"id": s.product.ID}).
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
						s.product.ID,
						s.product.Name,
					),
			)
	}

	product, err := s.repository.GetById(s.ctx, s.product.ID)

	s.NoError(err)
	s.Equal(product, s.product)
	s.NoError(s.mock.ExpectationsWereMet())
}

func (s *GetTestSuite) TestGetByIdFailure() {
	testCases := []struct {
		testName         string
		expectedError    error
		expectedErrorMsg string
	}{
		{
			testName:         "Unknown database error",
			expectedError:    errors.New("unknown error on getting product"),
			expectedErrorMsg: "unknown error on getting product",
		},
		{
			testName:         "Not found",
			expectedError:    sql.ErrNoRows,
			expectedErrorMsg: "product not found",
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.testName, func() {
			{
				query, args, err := sq.
					Select("*").
					From("product").
					Where(sq.Eq{"id": s.product.ID}).
					PlaceholderFormat(sq.Dollar).
					ToSql()

				s.NoError(err)

				s.mock.
					ExpectQuery(query).
					WithArgs(convertArgs(args)...).
					WillReturnError(testCase.expectedError)
			}

			product, err := s.repository.GetById(s.ctx, s.product.ID)

			s.NotNil(err)
			s.Equal(err.Error(), testCase.expectedErrorMsg)
			s.Equal(product, dto.Product{})
			s.NoError(s.mock.ExpectationsWereMet())
		})
	}
}
