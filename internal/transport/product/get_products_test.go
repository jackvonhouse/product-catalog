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

type GetTestSuite struct {
	suite.Suite

	// Вспомогательные параметры
	logger    log.Logger
	transport Transport

	// Входные параметры
	get dto.GetProduct

	category dto.Category
	products []dto.Product

	// Служебные параметры
	useCaseProductMock     *MockproductUseCase
	useCaseAccessTokenMock *MockuseCaseAccessToken
}

func TestSuiteGet(t *testing.T) {
	suite.Run(t, &GetTestSuite{})
}

func (s *GetTestSuite) SetupTest() {
	s.logger = log.NewNullLogger()
}

func (s *GetTestSuite) BeforeTest(_, _ string) {
	controller := gomock.NewController(s.T())

	// Входные значения по-умолчанию
	s.setupMock(controller).
		setupTransport().
		setupCategory(1, "Категория").
		setupProducts(1, "Продукт").
		setupGetProduct(0, 10)
}

func (s *GetTestSuite) setupMock(
	controller *gomock.Controller,
) *GetTestSuite {

	s.useCaseProductMock = NewMockproductUseCase(controller)
	s.useCaseAccessTokenMock = NewMockuseCaseAccessToken(controller)

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

func (s *GetTestSuite) setupTransport() *GetTestSuite {
	s.transport = New(s.useCaseProductMock, s.useCaseAccessTokenMock, s.logger)

	return s
}

func (s *GetTestSuite) setupGetProduct(
	offset, limit int,
) *GetTestSuite {

	s.get = dto.GetProduct{
		Limit:  limit,
		Offset: offset,
	}

	return s
}

func (s *GetTestSuite) TestGetSuccessful() {
	const (
		expectedBody   = ``
		expectedResult = `[{"id":1,"name":"Продукт"}]`
	)

	s.useCaseProductMock.
		EXPECT().
		Get(gomock.Any(), s.get).
		Return(s.products, nil).
		Times(1)

	r := httptest.NewRecorder()
	w, err := http.NewRequest(
		http.MethodGet,
		"",
		bytes.NewBufferString(expectedBody),
	)
	s.NoError(err)

	queries := w.URL.Query()
	queries.Add("limit", fmt.Sprintf("%d", s.get.Limit))
	queries.Add("offset", fmt.Sprintf("%d", s.get.Offset))

	w.URL.RawQuery = queries.Encode()

	s.transport.Get(r, w)

	bodyResult := r.Result()
	defer bodyResult.Body.Close()

	byteResult, err := io.ReadAll(bodyResult.Body)
	s.NoError(err)

	result := string(byteResult)

	s.Equal(expectedResult, strings.Trim(result, " \n"))
}

func (s *GetTestSuite) TestGetDefaultParamsSuccessful() {
	const (
		expectedBody   = ``
		expectedResult = `[{"id":1,"name":"Продукт"}]`
	)

	s.useCaseProductMock.
		EXPECT().
		Get(gomock.Any(), s.get).
		Return(s.products, nil).
		Times(1)

	r := httptest.NewRecorder()
	w, err := http.NewRequest(
		http.MethodGet,
		"",
		bytes.NewBufferString(expectedBody),
	)
	s.NoError(err)

	queries := w.URL.Query()
	queries.Add("limit", fmt.Sprintf("%s", "invalid limit"))
	queries.Add("offset", fmt.Sprintf("%s", "invalid offset"))

	w.URL.RawQuery = queries.Encode()

	s.transport.Get(r, w)

	bodyResult := r.Result()
	defer bodyResult.Body.Close()

	byteResult, err := io.ReadAll(bodyResult.Body)
	s.NoError(err)

	result := string(byteResult)

	s.Equal(expectedResult, strings.Trim(result, " \n"))
}

func (s *GetTestSuite) TestGetFailed() {
	const (
		expectedInternalErrorMsg        = "unknown error on getting product"
		expectedProductNotFoundErrorMsg = "product not found"
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
			expectedBody:     `{"name":"Продукт"}`,
			expectedResult:   fmt.Sprintf(`{"error":"%s"}`, expectedInternalErrorMsg),
		},
		{
			testName:         "Product not found",
			expectedErrorMsg: expectedProductNotFoundErrorMsg,
			expectedBody:     `{"name":"Продукт"}`,
			expectedResult:   fmt.Sprintf(`{"error":"%s"}`, expectedProductNotFoundErrorMsg),
		},
	}

	for _, testCase := range testCases {
		s.T().Run(testCase.testName, func(t *testing.T) {
			s.useCaseProductMock.
				EXPECT().
				Get(gomock.Any(), s.get).
				Return(s.products, fmt.Errorf(testCase.expectedErrorMsg)).
				Times(1)

			r := httptest.NewRecorder()
			w, err := http.NewRequest(
				http.MethodPost,
				"",
				bytes.NewBufferString(testCase.expectedBody),
			)
			s.NoError(err)

			s.transport.Get(r, w)

			bodyResult := r.Result()
			defer bodyResult.Body.Close()

			byteResult, err := io.ReadAll(bodyResult.Body)
			s.NoError(err)

			result := string(byteResult)

			s.Equal(testCase.expectedResult, strings.Trim(result, " \n"))
		})
	}
}

func (s *GetTestSuite) TestGetByCategoryIdSuccessful() {
	const (
		expectedBody   = ``
		expectedResult = `[{"id":1,"name":"Продукт"}]`
	)

	s.useCaseProductMock.
		EXPECT().
		GetByCategoryId(gomock.Any(), s.get, s.category.ID).
		Return(s.products, nil).
		Times(1)

	r := httptest.NewRecorder()
	w, err := http.NewRequest(
		http.MethodGet,
		"",
		bytes.NewBufferString(expectedBody),
	)
	s.NoError(err)

	queries := w.URL.Query()
	queries.Add("limit", fmt.Sprintf("%d", s.get.Limit))
	queries.Add("offset", fmt.Sprintf("%d", s.get.Offset))
	queries.Add("category_id", fmt.Sprintf("%d", s.category.ID))

	w.URL.RawQuery = queries.Encode()

	s.transport.GetByCategoryId(r, w)

	bodyResult := r.Result()
	defer bodyResult.Body.Close()

	byteResult, err := io.ReadAll(bodyResult.Body)
	s.NoError(err)

	result := string(byteResult)

	s.Equal(expectedResult, strings.Trim(result, " \n"))
}

func (s *GetTestSuite) TestGetByCategoryIdInvalidFailed() {
	const (
		expectedBody   = ``
		expectedResult = `{"error":"invalid category id"}`
	)

	r := httptest.NewRecorder()
	w, err := http.NewRequest(
		http.MethodGet,
		"",
		bytes.NewBufferString(expectedBody),
	)
	s.NoError(err)

	queries := w.URL.Query()
	queries.Add("limit", fmt.Sprintf("%d", s.get.Limit))
	queries.Add("offset", fmt.Sprintf("%d", s.get.Offset))
	queries.Add("category_id", fmt.Sprintf("%s", "invalid category id"))

	w.URL.RawQuery = queries.Encode()

	s.transport.GetByCategoryId(r, w)

	bodyResult := r.Result()
	defer bodyResult.Body.Close()

	byteResult, err := io.ReadAll(bodyResult.Body)
	s.NoError(err)

	result := string(byteResult)

	s.Equal(expectedResult, strings.Trim(result, " \n"))
}

func (s *GetTestSuite) TestGetByCategoryIdDefaultParamsSuccessful() {
	const (
		expectedBody   = ``
		expectedResult = `[{"id":1,"name":"Продукт"}]`
	)

	s.useCaseProductMock.
		EXPECT().
		GetByCategoryId(gomock.Any(), s.get, s.category.ID).
		Return(s.products, nil).
		Times(1)

	r := httptest.NewRecorder()
	w, err := http.NewRequest(
		http.MethodGet,
		"",
		bytes.NewBufferString(expectedBody),
	)
	s.NoError(err)

	queries := w.URL.Query()
	queries.Add("limit", fmt.Sprintf("%s", "invalid limit"))
	queries.Add("offset", fmt.Sprintf("%s", "invalid offset"))
	queries.Add("category_id", fmt.Sprintf("%d", s.category.ID))

	w.URL.RawQuery = queries.Encode()

	s.transport.GetByCategoryId(r, w)

	bodyResult := r.Result()
	defer bodyResult.Body.Close()

	byteResult, err := io.ReadAll(bodyResult.Body)
	s.NoError(err)

	result := string(byteResult)

	s.Equal(expectedResult, strings.Trim(result, " \n"))
}

func (s *GetTestSuite) TestGetByCategoryIdFailed() {
	const (
		expectedInternalErrorMsg        = "unknown error on getting product"
		expectedProductNotFoundErrorMsg = "product not found"
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
			expectedBody:     `{"name":"Продукт"}`,
			expectedResult:   fmt.Sprintf(`{"error":"%s"}`, expectedInternalErrorMsg),
		},
		{
			testName:         "Product not found",
			expectedErrorMsg: expectedProductNotFoundErrorMsg,
			expectedBody:     `{"name":"Продукт"}`,
			expectedResult:   fmt.Sprintf(`{"error":"%s"}`, expectedProductNotFoundErrorMsg),
		},
	}

	for _, testCase := range testCases {
		s.T().Run(testCase.testName, func(t *testing.T) {
			s.useCaseProductMock.
				EXPECT().
				GetByCategoryId(gomock.Any(), s.get, s.category.ID).
				Return(s.products, fmt.Errorf(testCase.expectedErrorMsg)).
				Times(1)

			r := httptest.NewRecorder()
			w, err := http.NewRequest(
				http.MethodPost,
				"",
				bytes.NewBufferString(testCase.expectedBody),
			)
			s.NoError(err)

			queries := w.URL.Query()
			queries.Add("limit", fmt.Sprintf("%s", "invalid limit"))
			queries.Add("offset", fmt.Sprintf("%s", "invalid offset"))
			queries.Add("category_id", fmt.Sprintf("%d", s.category.ID))

			w.URL.RawQuery = queries.Encode()

			s.transport.GetByCategoryId(r, w)

			bodyResult := r.Result()
			defer bodyResult.Body.Close()

			byteResult, err := io.ReadAll(bodyResult.Body)
			s.NoError(err)

			result := string(byteResult)

			s.Equal(testCase.expectedResult, strings.Trim(result, " \n"))
		})
	}
}
