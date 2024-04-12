package product

import (
	"context"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
)

type productService interface {
	Create(context.Context, dto.CreateProduct, dto.Category) (int, error)

	Get(context.Context, dto.GetProduct) ([]dto.Product, error)
	GetById(context.Context, int) (dto.Product, error)
	GetByCategoryId(context.Context, dto.GetProduct, dto.Category) ([]dto.Product, error)

	Update(context.Context, dto.UpdateProduct, dto.Category) (int, error)

	Delete(context.Context, int) (int, error)
}

type categoryService interface {
	GetById(context.Context, int) (dto.Category, error)
}

type UseCase struct {
	product  productService
	category categoryService

	logger log.Logger
}

func New(
	service productService,
	category categoryService,
	logger log.Logger,
) UseCase {

	return UseCase{
		product:  service,
		category: category,
		logger:   logger.WithField("unit", "product"),
	}
}

func (u UseCase) Create(
	ctx context.Context,
	data dto.CreateProduct,
) (int, error) {

	category, err := u.category.GetById(ctx, data.CategoryId)
	if err != nil {
		u.logger.Warnf("category not found: %s", err)

		return 0, err
	}

	return u.product.Create(ctx, data, category)
}

func (u UseCase) Get(
	ctx context.Context,
	data dto.GetProduct,
) ([]dto.Product, error) {

	return u.product.Get(ctx, data)
}

func (u UseCase) GetByCategoryId(
	ctx context.Context,
	data dto.GetProduct,
	categoryId int,
) ([]dto.Product, error) {

	category, err := u.category.GetById(ctx, categoryId)
	if err != nil {
		u.logger.Warnf("category not found: %s", err)

		return []dto.Product{}, err
	}

	return u.product.GetByCategoryId(ctx, data, category)
}

func (u UseCase) Update(
	ctx context.Context,
	data dto.UpdateProduct,
) (int, error) {

	category, err := u.category.GetById(ctx, data.NewCategoryId)
	if err != nil {
		u.logger.Warnf("category not found: %s", err)

		return 0, err
	}

	return u.product.Update(ctx, data, category)
}

func (u UseCase) Delete(
	ctx context.Context,
	id int,
) (int, error) {

	return u.product.Delete(ctx, id)
}
