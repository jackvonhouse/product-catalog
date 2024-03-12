package product

import (
	"context"
	"errors"
	pgerr "github.com/jackc/pgerrcode"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestCreate(t *testing.T) {
	controller := gomock.NewController(t)

	testCases := []struct {
		testName             string
		productName          string
		expectedProductId    int
		expectedProductError error
	}{
		{
			testName:             "Successful create",
			productName:          "Продукт",
			expectedProductId:    1,
			expectedProductError: nil,
		},
		{
			testName:             "Product already exists",
			productName:          "Продукт",
			expectedProductId:    0,
			expectedProductError: errors.New(pgerr.UniqueViolation),
		},
		{
			testName:             "Unknown error",
			productName:          "Продукт",
			expectedProductId:    0,
			expectedProductError: errors.New(""),
		},
	}

	ctx := context.Background()
	logger := log.NewLogNullLogger()
	repository := NewMockRepository(controller)
	service := New(repository, logger)

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			data := dto.CreateProduct{
				Name: testCase.productName,
			}

			repository.
				EXPECT().
				Create(ctx, data).
				Return(
					testCase.expectedProductId,
					testCase.expectedProductError,
				).
				Times(1)

			productId, err := service.Create(ctx, data)

			require.Equal(t, testCase.expectedProductId, productId)
			require.Equal(t, testCase.expectedProductError, err)
		})
	}
}
