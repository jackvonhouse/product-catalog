package product

import (
	"context"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
)

type service interface {
	Create(context.Context, dto.CreateProduct) (int, error)

	Get(context.Context, dto.GetProduct) ([]dto.Product, error)
	GetByCategoryId(context.Context, dto.GetProduct, int) ([]dto.Product, error)

	Update(context.Context, dto.UpdateProduct) (int, error)

	Delete(context.Context, int) (int, error)
}

type UseCase struct {
	product service

	logger log.Logger
}

func New(
	service service,
	logger log.Logger,
) UseCase {

	return UseCase{
		product: service,
		logger:  logger.WithField("unit", "product"),
	}
}

func (u UseCase) Create(
	ctx context.Context,
	data dto.CreateProduct,
) (int, error) {

	return u.product.Create(ctx, data)
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

	return u.product.GetByCategoryId(ctx, data, categoryId)
}

func (u UseCase) Update(
	ctx context.Context,
	data dto.UpdateProduct,
) (int, error) {

	return u.product.Update(ctx, data)
}

func (u UseCase) Delete(
	ctx context.Context,
	id int,
) (int, error) {

	return u.product.Delete(ctx, id)
}
