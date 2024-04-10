package refresh

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"github.com/jackvonhouse/product-catalog/config"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/internal/errors"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"golang.org/x/crypto/bcrypt"
)

const (
	refreshTokenSize = 32
)

type repository interface {
	Create(context.Context, dto.User, dto.RefreshToken) (string, error)

	GetById(context.Context, string) (dto.RefreshToken, error)

	Delete(context.Context, string) error
}

type Service struct {
	repository repository

	logger         log.Logger
	secretKey      string
	expireDuration int
}

func New(
	repository repository,
	config config.JWT,
	logger log.Logger,
) Service {

	return Service{
		repository:     repository,
		logger:         logger.WithField("unit", "refresh_token"),
		expireDuration: config.RefreshToken.Exp,
	}
}

func (s Service) Create(
	ctx context.Context,
	user dto.User,
) (string, string, error) {

	token, hashedToken, err := s.generateRefresh()
	if err != nil {
		return "", "", err
	}

	refreshToken := dto.RefreshToken{
		Token:          hashedToken,
		ExpireDuration: s.expireDuration,
	}

	refreshTokenId, err := s.repository.Create(ctx, user, refreshToken)

	return refreshTokenId, token, err
}

func (s Service) GetById(
	ctx context.Context,
	id string,
) (dto.RefreshToken, error) {

	return s.repository.GetById(ctx, id)
}

func (s Service) Delete(
	ctx context.Context,
	id string,
) error {

	return s.repository.Delete(ctx, id)
}

func (s Service) Verify(
	token, hashedToken string,
) error {

	if err := bcrypt.CompareHashAndPassword([]byte(hashedToken), []byte(token)); err != nil {
		s.logger.Warnf("tokens not equals: %s", err)

		return errors.
			ErrInvalidToken.
			New("refresh token has been modified or corrupted").
			Wrap(err)
	}

	return nil
}

func (s Service) generateRefresh() (string, string, error) {
	random := make([]byte, refreshTokenSize)

	_, err := rand.Read(random)
	if err != nil {
		s.logger.Warnf("can't generate refresh token: %s", err)

		return "", "", errors.
			ErrInternal.
			New("can't generate refresh token").
			Wrap(err)
	}

	token := base64.RawStdEncoding.EncodeToString(random)

	hashedToken, err := s.hashRefresh(token)
	if err != nil {
		return "", "", err
	}

	return token, hashedToken, nil
}

func (s Service) hashRefresh(
	token string,
) (string, error) {

	hashedToken, err := bcrypt.GenerateFromPassword(
		[]byte(token),
		bcrypt.DefaultCost,
	)

	if err != nil {
		s.logger.Warnf("can't hash refresh token: %s", err)

		return "", errors.
			ErrInternal.
			New("can't hash refresh token").
			Wrap(err)
	}

	return string(hashedToken), nil
}
