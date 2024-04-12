package category

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
	create dto.CreateCategory
	get    dto.GetCategory
	update dto.UpdateCategory

	product    dto.Product
	category   dto.Category
	categories []dto.Category

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
	s.setupCreateCategory("Категория").
		setupGetCategory(0, 10).
		setupUpdateCategory(1, "Продукт").
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

func (s *ProductTestSuite) setupCreateCategory(
	name string,
) *ProductTestSuite {

	s.create = dto.CreateCategory{
		Name: name,
	}

	return s
}

func (s *ProductTestSuite) setupGetCategory(
	offset, limit int,
) *ProductTestSuite {

	s.get = dto.GetCategory{
		Limit:  limit,
		Offset: offset,
	}

	return s
}

func (s *ProductTestSuite) setupUpdateCategory(
	id int,
	name string,
) *ProductTestSuite {

	s.update = dto.UpdateCategory{
		ID:   id,
		Name: name,
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
		Create(s.ctx, s.create).
		Return(1, nil)

	categoryId, err := s.service.Create(s.ctx, s.create)

	s.NoError(err)
	s.Equal(1, categoryId)
}

func (s *ProductTestSuite) TestGetSuccessful() {
	s.mock.
		EXPECT().
		Get(s.ctx, s.get).
		Return(s.categories, nil)

	categories, err := s.service.Get(s.ctx, s.get)

	s.NoError(err)
	s.Equal(s.categories, categories)
}

func (s *ProductTestSuite) TestGetByIdSuccessful() {
	s.mock.
		EXPECT().
		GetById(s.ctx, s.category.ID).
		Return(s.category, nil)

	category, err := s.service.GetById(s.ctx, s.product.ID)

	s.NoError(err)
	s.Equal(s.category, category)
}

func (s *ProductTestSuite) TestUpdateSuccessful() {
	s.mock.
		EXPECT().
		Update(s.ctx, s.update).
		Return(s.category.ID, nil).
		Times(1)

	categoryId, err := s.service.Update(s.ctx, s.update)

	s.NoError(err)
	s.Equal(1, categoryId)
}

func (s *ProductTestSuite) TestDeleteSuccessful() {
	s.mock.
		EXPECT().
		GetById(s.ctx, s.category.ID).
		Return(s.category, nil).
		Times(1)

	s.mock.
		EXPECT().
		Delete(s.ctx, s.category).
		Return(s.category.ID, nil).
		Times(1)

	categoryId, err := s.service.Delete(s.ctx, s.product.ID)

	s.NoError(err)
	s.Equal(1, categoryId)
}

func (s *ProductTestSuite) TestDeleteFailure() {
	const (
		expectedNotFoundErrorMsg = "category not found"
	)

	var (
		expectedError = errors.ErrNotFound.New(expectedNotFoundErrorMsg)
	)

	s.mock.
		EXPECT().
		GetById(s.ctx, s.category.ID).
		Return(dto.Category{}, expectedError).
		Times(1)

	categoryId, err := s.service.Delete(s.ctx, s.category.ID)

	s.NotNil(err)
	s.Equal(expectedError.Error(), err.Error())
	s.Equal(0, categoryId)
}
