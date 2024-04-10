package usecase

import (
	"github.com/jackvonhouse/product-catalog/app/service"
	"github.com/jackvonhouse/product-catalog/internal/usecase/auth"
	"github.com/jackvonhouse/product-catalog/internal/usecase/category"
	"github.com/jackvonhouse/product-catalog/internal/usecase/jwt/access"
	"github.com/jackvonhouse/product-catalog/internal/usecase/product"
	"github.com/jackvonhouse/product-catalog/pkg/log"
)

type UseCase struct {
	Product     product.UseCase
	Category    category.UseCase
	AccessToken access.UseCase
	Auth        auth.UseCase
}

func New(
	service service.Service,
	logger log.Logger,
) UseCase {

	useCaseLogger := logger.WithField("layer", "usecase")

	return UseCase{
		Product:     product.New(service.Product, service.Category, useCaseLogger),
		Category:    category.New(service.Category, useCaseLogger),
		AccessToken: access.New(service.AccessToken, useCaseLogger),
		Auth:        auth.New(service.AccessToken, service.RefreshToken, service.User, useCaseLogger),
	}
}
