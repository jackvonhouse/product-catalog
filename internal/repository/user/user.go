package user

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"github.com/patrickmn/go-cache"
	"io"
	"strings"
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
		logger: logger.WithField("unit", "user"),
		db:     db,
	}
}

func (r Repository) Create(
	_ context.Context,
	credentials dto.Credentials,
) (string, error) {

	r.db.DeleteExpired()

	logger := r.logger.WithFields(map[string]any{
		"username": credentials.Username,
		"password": "***",
	})

	userId := r.usernameToHash(credentials.Username)

	value := map[string]string{
		"username": credentials.Username,
		"password": credentials.Password,
	}

	if err := r.db.Add(userId, value, cache.NoExpiration); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			logger.Warnf("user already exists: %s", err)

			return "", r.errUserAlreadyExists(err)
		}
	}

	return userId, nil
}

func (r Repository) GetById(
	_ context.Context,
	id string,
) (dto.User, error) {

	logger := r.logger.WithFields(map[string]any{
		"user": map[string]string{
			"id": id,
		},
	})

	value, ok := r.db.Get(id)
	if !ok {
		logger.Warn("user not found")

		return dto.User{}, r.errNotFound("user", nil)
	}

	userMap := value.(map[string]string)

	user := dto.User{
		ID:       id,
		Username: userMap["username"],
		Password: userMap["password"],
	}

	return user, nil
}

func (r Repository) GetByUsername(
	ctx context.Context,
	username string,
) (dto.User, error) {

	userId := r.usernameToHash(username)

	return r.GetById(ctx, userId)
}

func (r Repository) usernameToHash(
	username string,
) string {

	h := md5.New()
	_, _ = io.WriteString(h, username)
	hashSum := h.Sum(nil)

	return fmt.Sprintf("%x", hashSum)
}
