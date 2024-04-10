package access

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackvonhouse/product-catalog/config"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/internal/errors"
	claim "github.com/jackvonhouse/product-catalog/internal/service/jwt"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"time"
)

type Service struct {
	logger    log.Logger
	secretKey string
	config    config.JWT
}

func New(
	config config.JWT,
	logger log.Logger,
) Service {

	return Service{
		logger:    logger.WithField("unit", "jwt"),
		secretKey: config.SecretKey,
		config:    config,
	}
}

func (s Service) Create(
	_ context.Context,
	data dto.AccessToken,
) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claim.AccessTokenClaim{
		Username:       data.Username,
		RefreshTokenId: data.RefreshTokenId,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(
					time.Duration(s.config.AccessToken.Exp) * time.Minute,
				),
			),
		},
	})

	signedToken, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		s.logger.Warnf("can't sign access jwt: %s", err)

		return "", errors.
			ErrInternal.
			New("can't sign access jwt").
			Wrap(err)
	}

	return signedToken, nil
}

func (s Service) Parse(
	token string,
) (dto.AccessToken, error) {

	accessTokenClaim, err := s.parseAccessToken(token)
	if err != nil {
		return dto.AccessToken{}, err
	}

	accessToken := dto.AccessToken{
		Username:       accessTokenClaim.Username,
		RefreshTokenId: accessTokenClaim.RefreshTokenId,
	}

	return accessToken, nil
}

func (s Service) Verify(
	token string,
) error {

	_, err := s.parseAccessToken(token)
	if err != nil {
		return err
	}

	return nil
}

func (s Service) getKey(_ *jwt.Token) (any, error) {
	return []byte(s.config.SecretKey), nil
}

func (s Service) parseAccessToken(
	token string,
) (claim.AccessTokenClaim, error) {

	accessTokenClaim := claim.AccessTokenClaim{}

	t, err := jwt.ParseWithClaims(token, &accessTokenClaim, s.getKey)
	if err != nil || !t.Valid {
		s.logger.Warnf("can't parse access token: %s", err)

		return claim.AccessTokenClaim{}, errors.
			ErrInvalidToken.
			New("access token has been modified or corrupted").
			Wrap(err)
	}

	if accessTokenClaim.ExpiresAt.Before(time.Now()) {
		s.logger.Warnf("access token has been expired: %s", err)

		return claim.AccessTokenClaim{}, errors.
			ErrInternal.
			New("access token has been expired").
			Wrap(err)
	}

	return accessTokenClaim, nil
}
