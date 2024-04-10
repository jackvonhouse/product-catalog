package category

import (
	"context"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"testing"
)

type CategoryTestSuite struct {
	suite.Suite

	// Вспомогательные параметры
	ctx     context.Context
	logger  log.Logger
	useCase UseCase

	// Входные параметры
	create dto.CreateCategory
	get    dto.GetCategory
	update dto.UpdateCategory

	product    dto.Product
	category   dto.Category
	categories []dto.Category

	// Служебные параметры
	categoryMock *MockcategoryService
}

func TestSuiteCreate(t *testing.T) {
	suite.Run(t, &CategoryTestSuite{})
}

func (s *CategoryTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.logger = log.NewNullLogger()
}

func (s *CategoryTestSuite) BeforeTest(_, _ string) {
	controller := gomock.NewController(s.T())

	s.setupMock(controller).setupUseCase()

	// Входные значения по-умолчанию
	s.setupCreateCategory("Категория").
		setupGetCategory(0, 10).
		setupUpdateCategory(1, "Продукт").
		setupProduct(1, "Продукт").
		setupCategory(1, "Категория")
}

func (s *CategoryTestSuite) setupMock(
	controller *gomock.Controller,
) *CategoryTestSuite {

	s.categoryMock = NewMockcategoryService(controller)

	return s
}

func (s *CategoryTestSuite) setupUseCase() *CategoryTestSuite {
	s.useCase = New(s.categoryMock, s.logger)

	return s
}

func (s *CategoryTestSuite) setupCreateCategory(
	name string,
) *CategoryTestSuite {

	s.create = dto.CreateCategory{
		Name: name,
	}

	return s
}

func (s *CategoryTestSuite) setupGetCategory(
	offset, limit int,
) *CategoryTestSuite {

	s.get = dto.GetCategory{
		Limit:  limit,
		Offset: offset,
	}

	return s
}

func (s *CategoryTestSuite) setupUpdateCategory(
	id int,
	name string,
) *CategoryTestSuite {

	s.update = dto.UpdateCategory{
		ID:   id,
		Name: name,
	}

	return s
}

func (s *CategoryTestSuite) setupProduct(
	id int,
	name string,
) *CategoryTestSuite {

	s.product = dto.Product{
		ID:   id,
		Name: name,
	}

	return s
}

func (s *CategoryTestSuite) setupCategory(
	id int,
	name string,
) *CategoryTestSuite {

	s.category = dto.Category{
		ID:   id,
		Name: name,
	}

	return s
}

func (s *CategoryTestSuite) TestCreateSuccessful() {
	s.categoryMock.
		EXPECT().
		Create(s.ctx, s.create).
		Return(s.category.ID, nil).
		Times(1)

	categoryId, err := s.useCase.Create(s.ctx, s.create)

	s.NoError(err)
	s.Equal(1, categoryId)
}

func (s *CategoryTestSuite) TestGetSuccessful() {
	s.categoryMock.
		EXPECT().
		Get(s.ctx, s.get).
		Return(s.categories, nil).
		Times(1)

	categories, err := s.useCase.Get(s.ctx, s.get)

	s.NoError(err)
	s.Equal(s.categories, categories)
}

func (s *CategoryTestSuite) TestUpdateSuccessful() {
	s.categoryMock.
		EXPECT().
		Update(s.ctx, s.update).
		Return(s.category.ID, nil).
		Times(1)

	categoryId, err := s.useCase.Update(s.ctx, s.update)

	s.NoError(err)
	s.Equal(1, categoryId)
}

func (s *CategoryTestSuite) TestDeleteSuccessful() {
	s.categoryMock.
		EXPECT().
		Delete(s.ctx, s.category.ID).
		Return(s.category.ID, nil).
		Times(1)

	categoryId, err := s.useCase.Delete(s.ctx, s.category.ID)

	s.NoError(err)
	s.Equal(1, categoryId)
}
