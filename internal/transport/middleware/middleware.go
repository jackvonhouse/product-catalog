package middleware

import (
	"context"
	"github.com/jackvonhouse/product-catalog/internal/transport"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"net/http"
	"strings"
)

const (
	authorizationHeader = "Authorization"
)

type useCaseAccessToken interface {
	Verify(context.Context, string) error
}

type Middleware struct {
	accessToken useCaseAccessToken
	logger      log.Logger
}

func New(
	jwt useCaseAccessToken,
	logger log.Logger,
) Middleware {

	return Middleware{
		accessToken: jwt,
		logger:      logger.WithField("unit", "middleware"),
	}
}

func (m Middleware) AuthorizedOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get(authorizationHeader)

		if authHeader == "" {
			transport.Error(w,
				http.StatusUnauthorized,
				http.StatusText(http.StatusUnauthorized),
			)

			return
		}

		accessToken := strings.Trim(
			strings.Replace(authHeader, "Bearer", "", 1),
			" ",
		)

		if err := m.accessToken.Verify(r.Context(), accessToken); err != nil {
			m.logger.
				WithField("token", accessToken).
				Warnf("access token verification failed: %s", err)

			code, msg := transport.ErrorToHttpResponse(err)

			transport.Error(w, code, msg)

			return
		}

		next.ServeHTTP(w, r)
	})
}
