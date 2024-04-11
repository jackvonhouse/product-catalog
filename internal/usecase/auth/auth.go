package auth

import (
	"context"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
)

type serviceAccessToken interface {
	Create(context.Context, dto.AccessToken) (string, error)

	Parse(string) (dto.AccessToken, error)

	Verify(string) error
}

type serviceRefreshToken interface {
	Create(context.Context, dto.User) (int, string, error)

	GetById(context.Context, int) (dto.RefreshToken, error)

	DeleteByUserId(context.Context, int) error

	Verify(string, string) error
}

type serviceUser interface {
	Create(context.Context, dto.Credentials) (int, error)

	GetByUsername(context.Context, string) (dto.User, error)

	Verify(context.Context, dto.Credentials) error
}

type UseCase struct {
	accessToken  serviceAccessToken
	refreshToken serviceRefreshToken
	user         serviceUser

	logger log.Logger
}

func New(
	accessToken serviceAccessToken,
	refreshToken serviceRefreshToken,
	user serviceUser,
	logger log.Logger,
) UseCase {

	return UseCase{
		accessToken:  accessToken,
		refreshToken: refreshToken,
		user:         user,
		logger:       logger.WithField("unit", "user"),
	}
}

func (u UseCase) SignUp(
	ctx context.Context,
	credentials dto.Credentials,
) (dto.TokenPair, error) {

	id, err := u.user.Create(ctx, credentials)
	if err != nil {
		u.logger.Warnf("can't create user: %s", err)

		return dto.TokenPair{}, err
	}

	incompleteUser := dto.User{
		ID:       id,
		Username: credentials.Username,
	}

	return u.createTokenPair(ctx, incompleteUser)
}

func (u UseCase) SignIn(
	ctx context.Context,
	credentials dto.Credentials,
) (dto.TokenPair, error) {

	if err := u.user.Verify(ctx, credentials); err != nil {
		u.logger.Warnf("password verify failed: %s", err)

		return dto.TokenPair{}, err
	}

	return u.updateTokenPair(ctx, credentials.Username)
}

func (u UseCase) Refresh(
	ctx context.Context,
	data dto.TokenPair,
) (dto.TokenPair, error) {

	accessToken, err := u.accessToken.Parse(data.AccessToken)
	if err != nil {
		return dto.TokenPair{}, err
	}

	refreshToken, err := u.refreshToken.GetById(ctx, accessToken.RefreshTokenId)
	if err != nil {
		return dto.TokenPair{}, err
	}

	if err := u.refreshToken.Verify(data.RefreshToken, refreshToken.Token); err != nil {
		return dto.TokenPair{}, err
	}

	return u.updateTokenPair(ctx, accessToken.Username)
}

func (u UseCase) createTokenPair(
	ctx context.Context,
	user dto.User,
) (dto.TokenPair, error) {

	refreshTokenId, refreshToken, err := u.refreshToken.Create(ctx, user)
	if err != nil {
		u.logger.Warnf("can't create refresh token: %s", err)

		return dto.TokenPair{}, err
	}

	access := dto.AccessToken{
		Username:       user.Username,
		RefreshTokenId: refreshTokenId,
	}

	accessToken, err := u.accessToken.Create(ctx, access)
	if err != nil {
		u.logger.Warnf("can't create access token: %s", err)

		return dto.TokenPair{}, err
	}

	return dto.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (u UseCase) updateTokenPair(
	ctx context.Context,
	username string,
) (dto.TokenPair, error) {

	user, err := u.user.GetByUsername(ctx, username)
	if err != nil {
		u.logger.Warnf("can't get user: %s", err)

		return dto.TokenPair{}, err
	}

	if err := u.refreshToken.DeleteByUserId(ctx, user.ID); err != nil {
		u.logger.Warnf("can't delete old refresh token: %s", err)

		return dto.TokenPair{}, err
	}

	return u.createTokenPair(ctx, user)
}
