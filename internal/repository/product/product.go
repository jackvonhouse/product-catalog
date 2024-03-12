package product

import (
	"context"
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	pgerr "github.com/jackc/pgerrcode"
	"github.com/jackvonhouse/product-catalog/internal/dto"
	"github.com/jackvonhouse/product-catalog/internal/errors"
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
		logger: logger.WithField("unit", "product"),
		db:     db,
	}
}

func (r Repository) Create(
	ctx context.Context,
	data dto.CreateProduct,
) (int, error) {

	query, args, err := sq.
		Insert("product").
		Columns("name").
		Values(data.Name).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"name": data.Name,
		},
	})

	if err != nil {
		logger.Warnf("error on building sql query: %s", err)

		return 0, errors.
			ErrInternal.
			New("error on building sql query").
			Wrap(err)
	}

	var productId int

	if err := r.db.GetContext(ctx, &productId, query, args...); err != nil {
		if e, ok := err.(*pq.Error); ok {
			switch e.Code {

			case pgerr.UniqueViolation:
				logger.Warnf("product already exists: %s", err)

				return 0, errors.
					ErrAlreadyExists.
					New("product already exists").
					Wrap(err)

			default:
				logger.Warnf("unknown error on creating product: %s", err)

				return 0, errors.
					ErrInternal.
					New("unknown error on creating product").
					Wrap(err)
			}
		}

		logger.Warnf("unknown error on creating product: %s", err)

		return 0, errors.
			ErrInternal.
			New("unknown error on creating product").
			Wrap(err)
	}

	return productId, nil
}

func (r Repository) Get(
	ctx context.Context,
	data dto.GetProduct,
) ([]dto.Product, error) {

	var (
		offset = uint64(data.Offset)
		limit  = uint64(data.Limit)
	)

	query, args, err := sq.
		Select("*").
		From("product").
		OrderBy("id DESC").
		Offset(offset).
		Limit(limit).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"limit":  data.Limit,
			"offset": data.Offset,
		},
	})

	if err != nil {
		logger.Warnf("error on building sql query: %s", err)

		return nil, errors.
			ErrInternal.
			New("error on building sql query").
			Wrap(err)
	}

	products := make([]dto.Product, 0)

	if err := r.db.SelectContext(ctx, &products, query, args...); err != nil {
		if !errpkg.Is(err, sql.ErrNoRows) {
			logger.Warnf("unknown error on getting products: %s", err)

			return []dto.Product{}, errors.
				ErrInternal.
				New("unknown error on getting products").
				Wrap(err)
		}

		logger.Warnf("no products: %s", err)

		return []dto.Product{}, errors.
			ErrNotFound.
			New("no products").
			Wrap(err)
	}

	return products, nil
}

func (r Repository) GetByCategoryId(
	ctx context.Context,
	data dto.GetProduct,
	categoryId int,
) ([]dto.Product, error) {

	var (
		offset = uint64(data.Offset)
		limit  = uint64(data.Limit)
	)

	query, args, err := sq.
		Select("*").
		From("product").
		Where(sq.Eq{"category_id": categoryId}).
		OrderBy("id DESC").
		Offset(offset).
		Limit(limit).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"limit":       data.Limit,
			"offset":      data.Offset,
			"category_id": categoryId,
		},
	})

	if err != nil {
		logger.Warnf("error on building sql query: %s", err)

		return nil, errors.
			ErrInternal.
			New("error on building sql query").
			Wrap(err)
	}

	products := make([]dto.Product, 0)

	if err := r.db.SelectContext(ctx, &products, query, args...); err != nil {
		if !errpkg.Is(err, sql.ErrNoRows) {
			logger.Warnf("unknown error on getting products: %s", err)

			return []dto.Product{}, errors.
				ErrInternal.
				New("unknown error on getting products").
				Wrap(err)
		}

		logger.Warnf("no products: %s", err)

		return []dto.Product{}, errors.
			ErrNotFound.
			New("no products").
			Wrap(err)
	}

	return products, nil
}

func (r Repository) Update(
	ctx context.Context,
	data dto.UpdateProduct,
) (int, error) {

	query, args, err := sq.
		Update("product").
		SetMap(map[string]any{
			"id":   data.ID,
			"name": data.Name,
		}).
		Where(sq.Eq{"id": data.ID}).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"product_id": data.ID,
			"name":       data.Name,
		},
	})

	if err != nil {
		logger.Warnf("error on building sql query: %s", err)

		return 0, errors.
			ErrInternal.
			New("error on building sql query").
			Wrap(err)
	}

	var productId int

	if err := r.db.GetContext(ctx, &productId, query, args...); err != nil {
		if errpkg.Is(err, sql.ErrNoRows) {
			logger.Warnf("product not found: %s", err)

			return 0, errors.
				ErrNotFound.
				New("product not found").
				Wrap(err)
		}

		if e, ok := err.(*pq.Error); ok {
			switch e.Code {

			case pgerr.UniqueViolation:
				logger.Warnf("product already exists: %s", err)

				return 0, errors.
					ErrAlreadyExists.
					New("product already exists").
					Wrap(err)

			default:
				logger.Warnf("unknown error on updating product: %s", err)

				return 0, errors.
					ErrInternal.
					New("unknown error on updating product").
					Wrap(err)
			}
		}

		logger.Warnf("unknown error on updating product: %s", err)

		return 0, errors.
			ErrInternal.
			New("unknown error on updating product").
			Wrap(err)
	}

	return productId, nil
}

func (r Repository) Delete(
	ctx context.Context,
	id int,
) (int, error) {

	query, args, err := sq.
		Delete("product").
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"product_id": id,
		},
	})

	if err != nil {
		logger.Warnf("error on building sql query: %s", err)

		return 0, errors.
			ErrInternal.
			New("error on building sql query").
			Wrap(err)
	}

	var productId int

	if err := r.db.GetContext(ctx, &productId, query, args...); err != nil {
		if !errpkg.Is(err, sql.ErrNoRows) {
			logger.Warnf("unknown error on deleting product: %s", err)

			return 0, errors.
				ErrInternal.
				New("unknown error on deleting product").
				Wrap(err)
		}

		logger.Warnf("product not found: %s", err)

		return 0, errors.
			ErrNotFound.
			New("product not found").
			Wrap(err)
	}

	return productId, nil
}
