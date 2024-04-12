package category

import (
	"context"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
)

type categoryService interface {
	Create(context.Context, dto.CreateCategory) (int, error)

	Get(context.Context, dto.GetCategory) ([]dto.Category, error)

	Update(context.Context, dto.UpdateCategory) (int, error)

	Delete(context.Context, int) (int, error)
}

type UseCase struct {
	category categoryService

	logger log.Logger
}

func New(
	category categoryService,
	logger log.Logger,
) UseCase {

	return UseCase{
		category: category,
		logger:   logger.WithField("unit", "category"),
	}
}

func (u UseCase) Create(
	ctx context.Context,
	data dto.CreateCategory,
) (int, error) {

	return u.category.Create(ctx, data)
}

func (u UseCase) Get(
	ctx context.Context,
	data dto.GetCategory,
) ([]dto.Category, error) {

	return u.category.Get(ctx, data)
}

func (u UseCase) Update(
	ctx context.Context,
	data dto.UpdateCategory,
) (int, error) {

	return u.category.Update(ctx, data)
}

func (u UseCase) Delete(
	ctx context.Context,
	id int,
) (int, error) {

	return u.category.Delete(ctx, id)
}
