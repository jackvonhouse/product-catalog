package refresh

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"github.com/patrickmn/go-cache"
	"strings"
	"time"
)

type Repository struct {
	logger log.Logger

	db *cache.Cache
}

func New(
	db *cache.Cache,
	logger log.Logger,
) Repository {

	return Repository{
		logger: logger.WithField("unit", "jwt"),
		db:     db,
	}
}

func (r Repository) Create(
	_ context.Context,
	user dto.User,
	refresh dto.RefreshToken,
) (string, error) {

	r.db.DeleteExpired()

	expiredAt := time.Now().Add(
		time.Duration(refresh.ExpireDuration) * time.Minute,
	)

	logger := r.logger.WithFields(map[string]any{
		"user": map[string]any{
			"id":   user.ID,
			"name": user.Username,
		},
		"refresh": map[string]any{
			"token":      refresh.Token,
			"expired_at": expiredAt,
		},
	})

	refreshTokenId := uuid.New().String()
	expireDuration := time.Duration(refresh.ExpireDuration) * time.Minute

	if err := r.db.Add(refreshTokenId, refresh.Token, expireDuration); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			logger.Warnf("refresh token already exists: %s", err)

			return "", r.errRefreshAlreadyExists(err)
		}
	}

	return refreshTokenId, nil
}

func (r Repository) GetById(
	_ context.Context,
	id string,
) (dto.RefreshToken, error) {

	logger := r.logger.WithFields(map[string]any{
		"token": map[string]any{
			"id": id,
		},
	})

	value, expiredAt, ok := r.db.GetWithExpiration(id)
	if !ok {
		logger.Warn("refresh token not found")

		return dto.RefreshToken{}, r.errNotFound("refresh token", nil)
	}

	if expiredAt.Before(time.Now()) {
		logger.Warn("refresh token expired")

		return dto.RefreshToken{}, r.errExpired("refresh token", nil)
	}

	refreshToken := dto.RefreshToken{
		ID:    id,
		Token: value.(string),
	}

	return refreshToken, nil
}

func (r Repository) Delete(
	_ context.Context,
	id string,
) error {

	r.db.Delete(id)

	return nil
}
