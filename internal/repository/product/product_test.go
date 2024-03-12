package product

import (
	"context"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSuccessfulCreate(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogNullLogger()

	db, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
	)
	require.NoError(t, err)

	wrappedDb := sqlx.NewDb(db, "sqlmock")
	repository := New(wrappedDb, logger)

	data := dto.CreateProduct{
		Name: "Продукт",
	}

	query, _, err := sq.
		Insert("product").
		Columns("name").
		Values(data.Name).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	require.NoError(t, err)

	mock.
		ExpectQuery(query).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(
			mock.
				NewRows([]string{"id"}).
				AddRow(1),
		)

	productId, err := repository.Create(ctx, data)

	require.NoError(t, err)
	require.Equal(t, 1, productId)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestFailureCreate(t *testing.T) {
	ctx := context.Background()
	logger := log.NewLogNullLogger()

	db, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
	)
	require.NoError(t, err)

	wrappedDb := sqlx.NewDb(db, "sqlmock")

	repository := New(wrappedDb, logger)

	testCases := []struct {
		testName                string
		productName             string
		expectedDatabaseError   error
		expectedRepositoryError error
	}{
		{
			testName:                "Product already exists",
			productName:             "Продукт",
			expectedDatabaseError:   &pq.Error{Code: "23505"},
			expectedRepositoryError: errors.New("product already exists"),
		},
		{
			testName:                "Unexpected postgresql error",
			productName:             "Продукт",
			expectedDatabaseError:   &pq.Error{Code: ""},
			expectedRepositoryError: errors.New("unknown error on creating product"),
		},
		{
			testName:                "Unknown error",
			productName:             "Продукт",
			expectedDatabaseError:   errors.New("unexpected error"),
			expectedRepositoryError: errors.New("unknown error on creating product"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			data := dto.CreateProduct{
				Name: testCase.productName,
			}

			query, _, err := sq.
				Insert("product").
				Columns("name").
				Values(data.Name).
				Suffix("RETURNING id").
				PlaceholderFormat(sq.Dollar).
				ToSql()

			require.NoError(t, err)

			mock.
				ExpectQuery(query).
				WithArgs(sqlmock.AnyArg()).
				WillReturnError(testCase.expectedDatabaseError)

			productId, err := repository.Create(ctx, data)

			require.Equal(t, 0, productId)
			require.Equal(t, testCase.expectedRepositoryError.Error(), err.Error())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
