package repository

import (
	"context"
	"github.com/jackvonhouse/product-catalog/app/infrastructure"
	"github.com/jackvonhouse/product-catalog/internal/infrastructure/postgres"
	"github.com/jackvonhouse/product-catalog/internal/repository/category"
	"github.com/jackvonhouse/product-catalog/internal/repository/jwt/refresh"
	"github.com/jackvonhouse/product-catalog/internal/repository/product"
	"github.com/jackvonhouse/product-catalog/internal/repository/user"
	"github.com/jackvonhouse/product-catalog/pkg/log"
)

type Repository struct {
	Product      product.Repository
	Category     category.Repository
	RefreshToken refresh.Repository
	User         user.Repository

	storage postgres.Database
}

func New(
	infrastructure infrastructure.Infrastructure,
	logger log.Logger,
) Repository {

	repositoryLogger := logger.WithField("layer", "repository")

	return Repository{
		Product: product.New(
			infrastructure.Postgres.Database(),
			repositoryLogger,
		),
		Category: category.New(
			infrastructure.Postgres.Database(),
			repositoryLogger,
		),
		RefreshToken: refresh.New(
			infrastructure.Postgres.Database(),
			repositoryLogger,
		),
		User: user.New(
			infrastructure.Postgres.Database(),
			repositoryLogger,
		),

		storage: infrastructure.Postgres,
	}
}

func (r Repository) Shutdown(
	_ context.Context,
) error {

	return r.storage.Database().Close()
}
