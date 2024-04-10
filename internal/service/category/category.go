package category

import (
	"context"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
)

type repository interface {
	Create(context.Context, dto.CreateCategory) (int, error)

	Get(context.Context, dto.GetCategory) ([]dto.Category, error)
	GetById(context.Context, int) (dto.Category, error)

	Update(context.Context, dto.UpdateCategory) (int, error)

	Delete(context.Context, dto.Category) (int, error)
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
		logger:     logger.WithField("unit", "category"),
	}
}

func (s Service) Create(
	ctx context.Context,
	data dto.CreateCategory,
) (int, error) {

	return s.repository.Create(ctx, data)
}

func (s Service) Get(
	ctx context.Context,
	data dto.GetCategory,
) ([]dto.Category, error) {

	return s.repository.Get(ctx, data)
}

func (s Service) GetById(
	ctx context.Context,
	id int,
) (dto.Category, error) {

	return s.repository.GetById(ctx, id)
}

func (s Service) Update(
	ctx context.Context,
	data dto.UpdateCategory,
) (int, error) {

	return s.repository.Update(ctx, data)
}

func (s Service) Delete(
	ctx context.Context,
	id int,
) (int, error) {

	category, err := s.repository.GetById(ctx, id)
	if err != nil {
		s.logger.Warnf("category not found: %s", err)

		return 0, err
	}

	return s.repository.Delete(ctx, category)
}
