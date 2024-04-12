package access

import (
	"context"
	"github.com/jackvonhouse/product-catalog/pkg/log"
)

type serviceAccessToken interface {
	Verify(string) error
}

type UseCase struct {
	accessToken serviceAccessToken
	logger      log.Logger
}

func New(
	accessToken serviceAccessToken,
	logger log.Logger,
) UseCase {

	return UseCase{
		accessToken: accessToken,
		logger:      logger.WithField("unit", "access_token"),
	}
}

func (u UseCase) Verify(
	_ context.Context,
	token string,
) error {

	return u.accessToken.Verify(token)
}
