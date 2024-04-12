package refresh

import (
	"context"
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	pgerr "github.com/jackc/pgerrcode"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	errpkg "github.com/jackvonhouse/product-catalog/pkg/errors"
	"github.com/jackvonhouse/product-catalog/pkg/log"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"time"
)

type Repository struct {
	logger log.Logger

	db *sqlx.DB
}

func New(
	db *sqlx.DB,
	logger log.Logger,
) Repository {

	return Repository{
		logger: logger.WithField("unit", "refresh-token"),
		db:     db,
	}
}

func (r Repository) Create(
	ctx context.Context,
	user dto.User,
	refresh dto.RefreshToken,
) (int, error) {

	if err := r.deleteExpired(ctx); err != nil {
		r.logger.Warnf("can't delete expired tokens: %s", err)
	}

	expireAt := time.Now().Add(
		time.Duration(refresh.ExpireDuration) * time.Minute,
	).Unix()

	query, args, err := sq.
		Insert("refresh").
		Columns("token", "user_id", "expire_at").
		Values(refresh.Token, user.ID, expireAt).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"user": map[string]any{
				"id":       user.ID,
				"username": user.Username,
			},
			"refresh": map[string]any{
				"id":    refresh.ID,
				"token": refresh.Token,
			},
		},
	})

	if err != nil {
		logger.Warnf("unknown error on building sql query: %s", err)

		return 0, r.errInternalBuildSql(err)
	}

	var refreshTokenId int

	if err := r.db.GetContext(ctx, &refreshTokenId, query, args...); err != nil {
		if e, ok := err.(*pq.Error); ok {
			switch e.Code {

			case pgerr.UniqueViolation:
				logger.Warnf("refresh token already exists: %s", err)

				return 0, r.errRefreshAlreadyExists(err)

			case pgerr.ForeignKeyViolation:
				logger.Warnf("user not found: %s", err)

				return 0, r.errNotFound("user", err)

			default:
				logger.Warnf("unknown error on creating refresh token: %s", err)

				return 0, r.errInternalCreateRefresh(err)
			}
		}

		logger.Warnf("unknown error on creating refresh token: %s", err)

		return 0, r.errInternalCreateRefresh(err)
	}

	return refreshTokenId, nil
}

func (r Repository) GetById(
	ctx context.Context,
	id int,
) (dto.RefreshToken, error) {

	if err := r.deleteExpired(ctx); err != nil {
		r.logger.Warnf("can't delete expired tokens: %s", err)
	}

	query, args, err := sq.
		Select("*").
		From("refresh").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"token": map[string]any{
				"id": id,
			},
		},
	})

	if err != nil {
		logger.Warnf("unknown error on building sql query: %s", err)

		return dto.RefreshToken{}, r.errInternalBuildSql(err)
	}

	var refreshToken dto.RefreshToken

	if err := r.db.GetContext(ctx, &refreshToken, query, args...); err != nil {
		if !errpkg.Is(err, sql.ErrNoRows) {
			logger.Warnf("unknown error on getting refresh token: %s", err)

			return dto.RefreshToken{}, r.errInternalGetRefresh(err)
		}

		logger.Warnf("refresh token not found: %s", err)

		return dto.RefreshToken{}, r.errNotFound("refresh token", err)
	}

	if refreshToken.ExpireAt < time.Now().Unix() {
		logger.Warn("refresh token expired")

		return dto.RefreshToken{}, r.errExpired("refresh token", nil)
	}

	return refreshToken, nil
}

func (r Repository) DeleteByUserId(
	ctx context.Context,
	id int,
) error {

	if err := r.deleteExpired(ctx); err != nil {
		r.logger.Warnf("can't delete expired tokens: %s", err)
	}

	query, args, err := sq.
		Delete("refresh").
		Where(sq.Eq{"user_id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"user": map[string]any{
				"id": id,
			},
		},
	})

	if err != nil {
		logger.Warnf("unknown error on building sql query: %s", err)

		return r.errInternalBuildSql(err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		logger.Warnf("unknown error on deleting refresh token: %s", err)

		return r.errInternalDeleteRefresh(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Warnf("unknown error on deleting refresh token: %s", err)

		return r.errInternalDeleteRefresh(err)
	}

	if rowsAffected == 0 {
		logger.Warnf("refresh token not found")

		return r.errNotFound("refresh token", nil)
	}

	return nil
}

func (r Repository) deleteExpired(
	ctx context.Context,
) error {

	now := time.Now().Unix()

	query, args, err := sq.
		Delete("refresh").
		Where(sq.LtOrEq{"expire_at": now}).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"now": now,
		},
	})

	if err != nil {
		logger.Warnf("unknown error on building sql query: %s", err)

		return r.errInternalBuildSql(err)
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		logger.Warnf("unknown error on deleting refresh token: %s", err)

		return r.errInternalDeleteRefresh(err)
	}

	rowsAffected, _ := result.RowsAffected()

	r.logger.Infof("deleted %d expired refresh tokens", rowsAffected)

	return nil
}
