package user

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
		logger: logger.WithField("unit", "user"),
		db:     db,
	}
}

func (r Repository) Create(
	ctx context.Context,
	credentials dto.Credentials,
) (int, error) {

	query, args, err := sq.
		Insert(`"user"`).
		Columns("username", "password").
		Values(credentials.Username, credentials.Password).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"username": credentials.Username,
			"password": "***",
		},
	})

	if err != nil {
		logger.Warnf("unknown error on building sql query: %s", err)

		return 0, r.errInternalBuildSql(err)
	}

	var userId int

	if err := r.db.GetContext(ctx, &userId, query, args...); err != nil {
		if e, ok := err.(*pq.Error); ok {
			switch e.Code {

			case pgerr.UniqueViolation:
				logger.Warnf("user already exists: %s", err)

				return 0, r.errUserAlreadyExists(err)

			default:
				logger.Warnf("unknown error on creating user: %s", err)

				return 0, r.errInternalCreateUser(err)
			}
		}

		logger.Warnf("unknown error on creating user: %s", err)

		return 0, r.errInternalCreateUser(err)
	}

	return userId, nil
}

func (r Repository) GetByUsername(
	ctx context.Context,
	username string,
) (dto.User, error) {

	query, args, err := sq.
		Select("*").
		From(`"user"`).
		Where(sq.Eq{"username": username}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"username": username,
		},
	})

	if err != nil {
		logger.Warnf("unknown error on building sql query: %s", err)

		return dto.User{}, r.errInternalBuildSql(err)
	}

	user := dto.User{}

	if err := r.db.GetContext(ctx, &user, query, args...); err != nil {
		if !errpkg.Is(err, sql.ErrNoRows) {
			logger.Warnf("unknown error on getting user: %s", err)

			return dto.User{}, r.errInternalGetUser(err)
		}

		logger.Warnf("user not found: %s", err)

		return dto.User{}, r.errNotFound("user", err)
	}

	return user, nil
}
