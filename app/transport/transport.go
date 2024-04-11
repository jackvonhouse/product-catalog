package transport

import (
	"github.com/gorilla/mux"
	"github.com/jackvonhouse/product-catalog/app/usecase"
	"github.com/jackvonhouse/product-catalog/internal/transport/auth"
	"github.com/jackvonhouse/product-catalog/internal/transport/category"
	"github.com/jackvonhouse/product-catalog/internal/transport/product"
	"github.com/jackvonhouse/product-catalog/internal/transport/router"
	"github.com/jackvonhouse/product-catalog/pkg/log"
)

type Transport struct {
	router router.Router
}

func New(
	useCase usecase.UseCase,
	logger log.Logger,
) Transport {

	transportLogger := logger.WithField("layer", "transport")

	r := router.New("/api/v1")

	r.Handle(map[string]router.Handlify{
		"/product":  product.New(useCase.Product, useCase.AccessToken, transportLogger),
		"/category": category.New(useCase.Category, useCase.AccessToken, transportLogger),
		"/user":     auth.New(useCase.Auth, transportLogger),
	})

	return Transport{
		router: r,
	}
}

func (t Transport) Router() *mux.Router { return t.router.Router() }
