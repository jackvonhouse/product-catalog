package product

import (
	"bytes"
	"fmt"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type CreateTestSuite struct {
	suite.Suite

	// Вспомогательные параметры
	logger    log.Logger
	transport Transport

	// Входные параметры
	create dto.CreateProduct

	// Служебные параметры
	useCaseProductMock     *MockproductUseCase
	useCaseAccessTokenMock *MockuseCaseAccessToken
}

func TestSuiteCreate(t *testing.T) {
	suite.Run(t, &CreateTestSuite{})
}

func (s *CreateTestSuite) SetupTest() {
	s.logger = log.NewNullLogger()
}

func (s *CreateTestSuite) BeforeTest(_, _ string) {
	controller := gomock.NewController(s.T())

	// Входные значения по-умолчанию
	s.setupMock(controller).setupTransport().
		setupCreateProduct("Продукт")
}

func (s *CreateTestSuite) setupMock(
	controller *gomock.Controller,
) *CreateTestSuite {

	s.useCaseProductMock = NewMockproductUseCase(controller)
	s.useCaseAccessTokenMock = NewMockuseCaseAccessToken(controller)

	return s
}

func (s *CreateTestSuite) setupTransport() *CreateTestSuite {
	s.transport = New(s.useCaseProductMock, s.useCaseAccessTokenMock, s.logger)

	return s
}

func (s *CreateTestSuite) setupCreateProduct(
	name string,
) *CreateTestSuite {

	s.create = dto.CreateProduct{
		Name:       name,
		CategoryId: 1,
	}

	return s
}

func (s *CreateTestSuite) TestCreateSuccessful() {
	const (
		expectedBody   = `{"name": "Продукт","category_id":1}`
		expectedResult = `{"id":1}`
	)

	s.useCaseProductMock.
		EXPECT().
		Create(gomock.Any(), s.create).
		Return(1, nil).
		Times(1)

	r := httptest.NewRecorder()
	w, err := http.NewRequest(
		http.MethodPost,
		"",
		bytes.NewBufferString(expectedBody),
	)
	s.NoError(err)

	s.transport.Create(r, w)

	bodyResult := r.Result()
	defer bodyResult.Body.Close()

	byteResult, err := io.ReadAll(bodyResult.Body)
	s.NoError(err)

	result := string(byteResult)

	s.Equal(expectedResult, strings.Trim(result, " \n"))
}

func (s *CreateTestSuite) TestCreateDecodeFailed() {
	const (
		expectedBody   = `{wrong json}`
		expectedResult = `{"error":"invalid json structure"}`
	)

	r := httptest.NewRecorder()
	w, err := http.NewRequest(
		http.MethodPost,
		"",
		bytes.NewBufferString(expectedBody),
	)
	s.NoError(err)

	s.transport.Create(r, w)

	bodyResult := r.Result()
	defer bodyResult.Body.Close()

	byteResult, err := io.ReadAll(bodyResult.Body)
	s.NoError(err)

	result := string(byteResult)

	s.Equal(expectedResult, strings.Trim(result, " \n"))
}

func (s *CreateTestSuite) TestCreateEmptyNameFailed() {
	const (
		expectedBody   = `{"name":""}`
		expectedResult = `{"error":"name can't be empty"}`
	)

	r := httptest.NewRecorder()
	w, err := http.NewRequest(
		http.MethodPost,
		"",
		bytes.NewBufferString(expectedBody),
	)
	s.NoError(err)

	s.transport.Create(r, w)

	bodyResult := r.Result()
	defer bodyResult.Body.Close()

	byteResult, err := io.ReadAll(bodyResult.Body)
	s.NoError(err)

	result := string(byteResult)

	s.Equal(expectedResult, strings.Trim(result, " \n"))
}

func (s *CreateTestSuite) TestCreateFailed() {
	const (
		expectedInternalErrorMsg                       = "unknown error on creating product"
		expectedProductAlreadyExistsErrorMsg           = "product already exists"
		expectedProductInCategoryAlreadyExistsErrorMsg = "product already exists"
		expectedProductNotFoundErrorMsg                = "product not found"
		expectedCategoryNotFoundErrorMsg               = "category not found"
	)

	testCases := []struct {
		testName         string
		expectedErrorMsg string
		expectedBody     string
		expectedResult   string
	}{
		{
			testName:         "Internal error",
			expectedErrorMsg: expectedInternalErrorMsg,
			expectedBody:     `{"name":"Продукт","category_id":1}`,
			expectedResult:   fmt.Sprintf(`{"error":"%s"}`, expectedInternalErrorMsg),
		},
		{
			testName:         "Product already exists",
			expectedErrorMsg: expectedProductAlreadyExistsErrorMsg,
			expectedBody:     `{"name":"Продукт","category_id":1}`,
			expectedResult:   fmt.Sprintf(`{"error":"%s"}`, expectedProductAlreadyExistsErrorMsg),
		},
		{
			testName:         "Product in category already exists",
			expectedErrorMsg: expectedProductInCategoryAlreadyExistsErrorMsg,
			expectedBody:     `{"name":"Продукт","category_id":1}`,
			expectedResult:   fmt.Sprintf(`{"error":"%s"}`, expectedProductInCategoryAlreadyExistsErrorMsg),
		},
		{
			testName:         "Product not found",
			expectedErrorMsg: expectedProductNotFoundErrorMsg,
			expectedBody:     `{"name":"Продукт","category_id":1}`,
			expectedResult:   fmt.Sprintf(`{"error":"%s"}`, expectedProductNotFoundErrorMsg),
		},
		{
			testName:         "Category not found",
			expectedErrorMsg: expectedCategoryNotFoundErrorMsg,
			expectedBody:     `{"name":"Продукт","category_id":1}`,
			expectedResult:   fmt.Sprintf(`{"error":"%s"}`, expectedCategoryNotFoundErrorMsg),
		},
	}

	for _, testCase := range testCases {
		s.T().Run(testCase.testName, func(t *testing.T) {
			s.useCaseProductMock.
				EXPECT().
				Create(gomock.Any(), s.create).
				Return(0, fmt.Errorf(testCase.expectedErrorMsg)).
				Times(1)

			r := httptest.NewRecorder()
			w, err := http.NewRequest(
				http.MethodPost,
				"",
				bytes.NewBufferString(testCase.expectedBody),
			)
			s.NoError(err)

			s.transport.Create(r, w)

			bodyResult := r.Result()
			defer bodyResult.Body.Close()

			byteResult, err := io.ReadAll(bodyResult.Body)
			s.NoError(err)

			result := string(byteResult)

			s.Equal(testCase.expectedResult, strings.Trim(result, " \n"))
		})
	}
}
