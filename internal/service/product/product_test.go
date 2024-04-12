package product

import (
	"context"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/internal/errors"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"testing"
)

type ProductTestSuite struct {
	suite.Suite

	// Вспомогательные параметры
	ctx     context.Context
	logger  log.Logger
	service Service

	// Входные параметры
	create dto.CreateProduct
	get    dto.GetProduct
	update dto.UpdateProduct

	product  dto.Product
	category dto.Category
	products []dto.Product

	// Служебные параметры
	mock *Mockrepository
}

func TestSuiteCreate(t *testing.T) {
	suite.Run(t, &ProductTestSuite{})
}

func (s *ProductTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.logger = log.NewNullLogger()
}

func (s *ProductTestSuite) BeforeTest(_, _ string) {
	controller := gomock.NewController(s.T())

	s.setupMock(controller).setupUseCase()

	// Входные значения по-умолчанию
	s.setupCreateProduct("Продукт", 1).
		setupGetProduct(0, 10).
		setupUpdateProduct(1, "Продукт", 1, 2).
		setupProduct(1, "Продукт").
		setupCategory(1, "Категория")
}

func (s *ProductTestSuite) setupMock(
	controller *gomock.Controller,
) *ProductTestSuite {

	s.mock = NewMockrepository(controller)

	return s
}

func (s *ProductTestSuite) setupUseCase() *ProductTestSuite {
	s.service = New(s.mock, s.logger)

	return s
}

func (s *ProductTestSuite) setupCreateProduct(
	name string,
	categoryId int,
) *ProductTestSuite {

	s.create = dto.CreateProduct{
		Name:       name,
		CategoryId: categoryId,
	}

	return s
}

func (s *ProductTestSuite) setupGetProduct(
	offset, limit int,
) *ProductTestSuite {

	s.get = dto.GetProduct{
		Limit:  limit,
		Offset: offset,
	}

	return s
}

func (s *ProductTestSuite) setupUpdateProduct(
	id int,
	name string,
	old, new int,
) *ProductTestSuite {

	s.update = dto.UpdateProduct{
		ID:            id,
		Name:          name,
		OldCategoryId: old,
		NewCategoryId: new,
	}

	return s
}

func (s *ProductTestSuite) setupProduct(
	id int,
	name string,
) *ProductTestSuite {

	s.product = dto.Product{
		ID:   id,
		Name: name,
	}

	return s
}

func (s *ProductTestSuite) setupCategory(
	id int,
	name string,
) *ProductTestSuite {

	s.category = dto.Category{
		ID:   id,
		Name: name,
	}

	return s
}

func (s *ProductTestSuite) TestCreateSuccessful() {
	s.mock.
		EXPECT().
		Create(s.ctx, s.create, s.category).
		Return(1, nil)

	productId, err := s.service.Create(s.ctx, s.create, s.category)

	s.NoError(err)
	s.Equal(1, productId)
}

func (s *ProductTestSuite) TestGetSuccessful() {
	s.mock.
		EXPECT().
		Get(s.ctx, s.get).
		Return(s.products, nil)

	products, err := s.service.Get(s.ctx, s.get)

	s.NoError(err)
	s.Equal(s.products, products)
}

func (s *ProductTestSuite) TestGetByIdSuccessful() {
	s.mock.
		EXPECT().
		GetById(s.ctx, s.product.ID).
		Return(s.product, nil)

	product, err := s.service.GetById(s.ctx, s.product.ID)

	s.NoError(err)
	s.Equal(s.product, product)
}

func (s *ProductTestSuite) TestGetByCategoryIdSuccessful() {
	s.mock.
		EXPECT().
		GetByCategoryId(s.ctx, s.get, s.category).
		Return(s.products, nil).
		Times(1)

	products, err := s.service.GetByCategoryId(s.ctx, s.get, s.category)

	s.NoError(err)
	s.Equal(s.products, products)
}

func (s *ProductTestSuite) TestUpdateSuccessful() {
	s.mock.
		EXPECT().
		GetById(s.ctx, s.product.ID).
		Return(s.product, nil)

	s.mock.
		EXPECT().
		Update(s.ctx, s.update, s.product, s.category).
		Return(s.product.ID, nil).
		Times(1)

	productId, err := s.service.Update(s.ctx, s.update, s.category)

	s.NoError(err)
	s.Equal(1, productId)
}

func (s *ProductTestSuite) TestUpdateFailure() {
	const (
		expectedNotFoundErrorMsg = "product not found"
	)

	var (
		expectedError = errors.ErrNotFound.New(expectedNotFoundErrorMsg)
	)

	s.mock.
		EXPECT().
		GetById(s.ctx, s.product.ID).
		Return(dto.Product{}, expectedError)

	productId, err := s.service.Update(s.ctx, s.update, s.category)

	s.NotNil(err)
	s.Equal(expectedError.Error(), err.Error())
	s.Equal(0, productId)
}

func (s *ProductTestSuite) TestDeleteSuccessful() {
	s.mock.
		EXPECT().
		GetById(s.ctx, s.product.ID).
		Return(s.product, nil).
		Times(1)

	s.mock.
		EXPECT().
		Delete(s.ctx, s.product).
		Return(s.product.ID, nil).
		Times(1)

	productId, err := s.service.Delete(s.ctx, s.product.ID)

	s.NoError(err)
	s.Equal(1, productId)
}

func (s *ProductTestSuite) TestDeleteFailure() {
	const (
		expectedNotFoundErrorMsg = "product not found"
	)

	var (
		expectedError = errors.ErrNotFound.New(expectedNotFoundErrorMsg)
	)

	s.mock.
		EXPECT().
		GetById(s.ctx, s.product.ID).
		Return(dto.Product{}, expectedError).
		Times(1)

	productId, err := s.service.Delete(s.ctx, s.product.ID)

	s.NotNil(err)
	s.Equal(expectedError.Error(), err.Error())
	s.Equal(0, productId)
}
