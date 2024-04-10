package app

import (
	"context"
	"github.com/jackvonhouse/product-catalog/app/infrastructure"
	"github.com/jackvonhouse/product-catalog/app/repository"
	"github.com/jackvonhouse/product-catalog/app/service"
	"github.com/jackvonhouse/product-catalog/app/transport"
	"github.com/jackvonhouse/product-catalog/app/usecase"
	"github.com/jackvonhouse/product-catalog/config"
	"github.com/jackvonhouse/product-catalog/internal/infrastructure/server/http"
	"github.com/jackvonhouse/product-catalog/pkg/log"
)

type App struct {
	infrastructure infrastructure.Infrastructure
	repository     repository.Repository
	service        service.Service
	useCase        usecase.UseCase
	transport      transport.Transport

	config config.Config
	logger log.Logger
	server http.Server
}

func New(
	ctx context.Context,
	config config.Config,
	logger log.Logger,
) (App, error) {

	i, err := infrastructure.New(ctx, config, logger)
	if err != nil {
		return App{}, err
	}

	r := repository.New(i, logger)
	s := service.New(r, config.JWT, logger)
	u := usecase.New(s, logger)
	t := transport.New(u, logger)

	httpServer := http.New(t.Router(), config.Server)

	return App{
		infrastructure: i,
		repository:     r,
		service:        s,
		useCase:        u,
		transport:      t,
		config:         config,
		logger:         logger,
		server:         httpServer,
	}, nil
}

func (a App) Run() error {
	a.logger.Infof("running http server on %d port", a.config.Server.Port)

	return a.server.Run()
}

func (a App) Shutdown(
	ctx context.Context,
) error {

	a.logger.Info("http server shutdown")

	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}

	a.logger.Info("repository shutdown")

	if err := a.repository.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
