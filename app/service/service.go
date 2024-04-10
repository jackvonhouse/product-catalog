package service

import (
	"github.com/jackvonhouse/product-catalog/app/repository"
	"github.com/jackvonhouse/product-catalog/config"
	"github.com/jackvonhouse/product-catalog/internal/service/category"
	"github.com/jackvonhouse/product-catalog/internal/service/jwt/access"
	"github.com/jackvonhouse/product-catalog/internal/service/jwt/refresh"
	"github.com/jackvonhouse/product-catalog/internal/service/product"
	"github.com/jackvonhouse/product-catalog/internal/service/user"
	"github.com/jackvonhouse/product-catalog/pkg/log"
)

type Service struct {
	Product      product.Service
	Category     category.Service
	AccessToken  access.Service
	RefreshToken refresh.Service
	User         user.Service
}

func New(
	repository repository.Repository,
	config config.JWT,
	logger log.Logger,
) Service {

	serviceLogger := logger.WithField("layer", "service")

	return Service{
		Product:      product.New(repository.Product, serviceLogger),
		Category:     category.New(repository.Category, serviceLogger),
		AccessToken:  access.New(config, serviceLogger),
		RefreshToken: refresh.New(repository.RefreshToken, config, serviceLogger),
		User:         user.New(repository.User, serviceLogger),
	}
}
