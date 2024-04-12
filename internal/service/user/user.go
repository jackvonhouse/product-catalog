package user

import (
	"context"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/internal/errors"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"golang.org/x/crypto/bcrypt"
)

type repository interface {
	Create(context.Context, dto.Credentials) (int, error)

	GetByUsername(context.Context, string) (dto.User, error)
}

type Service struct {
	repository repository

	logger log.Logger
}

func New(
	repository repository,
	logger log.Logger,
) Service {

	return Service{
		repository: repository,
		logger:     logger.WithField("unit", "user"),
	}
}

func (s Service) Create(
	ctx context.Context,
	credentials dto.Credentials,
) (int, error) {

	password, err := bcrypt.GenerateFromPassword(
		[]byte(credentials.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		s.logger.Warnf("can't hash password: %s", err)

		return 0, errors.
			ErrInternal.
			New("can't hash password").
			Wrap(err)
	}

	credentials.Password = string(password)

	return s.repository.Create(ctx, credentials)
}

func (s Service) GetByUsername(
	ctx context.Context,
	userName string,
) (dto.User, error) {

	return s.repository.GetByUsername(ctx, userName)
}

func (s Service) Verify(
	ctx context.Context,
	credentials dto.Credentials,
) error {

	user, err := s.GetByUsername(ctx, credentials.Username)
	if err != nil {
		s.logger.Warnf("user not found: %s", err)

		return err
	}

	hashedPassword := []byte(user.Password)
	password := []byte(credentials.Password)

	if err := bcrypt.CompareHashAndPassword(hashedPassword, password); err != nil {
		s.logger.Warnf("can't compare passwords: %s", err)

		return errors.
			ErrInternal.
			New("can't compare passwords").
			Wrap(err)
	}

	return nil
}
