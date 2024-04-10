package infrastructure

import (
	"context"
	"github.com/jackvonhouse/product-catalog/config"
	"github.com/jackvonhouse/product-catalog/internal/infrastructure/cache"
	"github.com/jackvonhouse/product-catalog/internal/infrastructure/postgres"
	"github.com/jackvonhouse/product-catalog/pkg/log"
)

type Infrastructure struct {
	Postgres postgres.Database
	Cache    cache.Database
}

func New(
	ctx context.Context,
	config config.Config,
	logger log.Logger,
) (Infrastructure, error) {

	infrastructureLog := logger.WithField("layer", "infrastructure")

	pg, err := postgres.New(ctx, config.Database, infrastructureLog)
	if err != nil {
		infrastructureLog.Warn(err)

		return Infrastructure{}, err
	}

	c, _ := cache.New(ctx, config.Cache, infrastructureLog)

	return Infrastructure{
		Postgres: pg,
		Cache:    c,
	}, nil
}
