package product

import (
	"context"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
)

type Repository interface {
	Create(context.Context, dto.CreateProduct) (int, error)

	Get(context.Context, dto.GetProduct) ([]dto.Product, error)
	GetByCategoryId(context.Context, dto.GetProduct, int) ([]dto.Product, error)

	Update(context.Context, dto.UpdateProduct) (int, error)

	Delete(context.Context, int) (int, error)
}

type Service struct {
	repository Repository

	logger log.Logger
}

func New(
	repository Repository,
	logger log.Logger,
) Service {

	return Service{
		repository: repository,
		logger:     logger.WithField("unit", "product"),
	}
}

func (s Service) Create(
	ctx context.Context,
	data dto.CreateProduct,
) (int, error) {

	return s.repository.Create(ctx, data)
}

func (s Service) Get(
	ctx context.Context,
	data dto.GetProduct,
) ([]dto.Product, error) {

	return s.repository.Get(ctx, data)
}

func (s Service) GetByCategoryId(
	ctx context.Context,
	data dto.GetProduct,
	categoryId int,
) ([]dto.Product, error) {

	return s.repository.GetByCategoryId(ctx, data, categoryId)
}

func (s Service) Update(
	ctx context.Context,
	data dto.UpdateProduct,
) (int, error) {

	return s.repository.Update(ctx, data)
}

func (s Service) Delete(
	ctx context.Context,
	id int,
) (int, error) {

	return s.repository.Delete(ctx, id)
}
