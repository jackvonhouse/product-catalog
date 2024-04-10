package product

import (
	"context"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
)

type repository interface {
	Create(context.Context, dto.CreateProduct, dto.Category) (int, error)

	Get(context.Context, dto.GetProduct) ([]dto.Product, error)
	GetById(context.Context, int) (dto.Product, error)
	GetByCategoryId(context.Context, dto.GetProduct, dto.Category) ([]dto.Product, error)

	Update(context.Context, dto.UpdateProduct, dto.Product, dto.Category) (int, error)

	Delete(context.Context, dto.Product) (int, error)
}

type Service struct {
	repository repository

	logger log.Logger
}

func New(
	repository repository,
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
	category dto.Category,
) (int, error) {

	return s.repository.Create(ctx, data, category)
}

func (s Service) Get(
	ctx context.Context,
	data dto.GetProduct,
) ([]dto.Product, error) {

	return s.repository.Get(ctx, data)
}

func (s Service) GetById(
	ctx context.Context,
	id int,
) (dto.Product, error) {

	return s.repository.GetById(ctx, id)
}

func (s Service) GetByCategoryId(
	ctx context.Context,
	data dto.GetProduct,
	category dto.Category,
) ([]dto.Product, error) {

	return s.repository.GetByCategoryId(ctx, data, category)
}

func (s Service) Update(
	ctx context.Context,
	data dto.UpdateProduct,
	category dto.Category,
) (int, error) {

	product, err := s.repository.GetById(ctx, data.ID)
	if err != nil {
		s.logger.Warnf("product not found: %s", err)

		return 0, err
	}

	return s.repository.Update(ctx, data, product, category)
}

func (s Service) Delete(
	ctx context.Context,
	id int,
) (int, error) {

	product, err := s.repository.GetById(ctx, id)
	if err != nil {
		s.logger.Warnf("product not found: %s", err)

		return 0, err
	}

	return s.repository.Delete(ctx, product)
}
