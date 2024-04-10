package category

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
		logger: logger.WithField("unit", "category"),
		db:     db,
	}
}

func (r Repository) Create(
	ctx context.Context,
	data dto.CreateCategory,
) (int, error) {

	query, args, err := sq.
		Insert("category").
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
		logger.Warnf("unknown error on building sql query: %s", err)

		return 0, r.errInternalBuildSql(err)
	}

	var categoryId int

	if err := r.db.GetContext(ctx, &categoryId, query, args...); err != nil {
		if e, ok := err.(*pq.Error); ok {
			switch e.Code {

			case pgerr.UniqueViolation:
				logger.Warnf("category already exists: %s", err)

				return 0, r.errCategoryAlreadyExists(err)

			default:
				logger.Warnf("unknown error on creating category: %s", err)

				return 0, r.errInternalCreateCategory(err)
			}
		}

		logger.Warnf("unknown error on creating category: %s", err)

		return 0, r.errInternalCreateCategory(err)
	}

	return categoryId, nil
}

func (r Repository) Get(
	ctx context.Context,
	data dto.GetCategory,
) ([]dto.Category, error) {

	var (
		offset = uint64(data.Offset)
		limit  = uint64(data.Limit)
	)

	query, args, err := sq.
		Select("*").
		From("category").
		OrderBy("id ASC").
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
		logger.Warnf("unknown error on building sql query: %s", err)

		return []dto.Category{}, r.errInternalBuildSql(err)
	}

	categories := make([]dto.Category, 0)

	if err := r.db.SelectContext(ctx, &categories, query, args...); err != nil {
		if !errpkg.Is(err, sql.ErrNoRows) {
			logger.Warnf("unknown error on getting categories: %s", err)

			return []dto.Category{}, r.errInternalGetCategories(err)
		}

		logger.Warnf("no categories: %s", err)

		return []dto.Category{}, r.errNotFound("categories", err)
	}

	return categories, nil
}

func (r Repository) GetById(
	ctx context.Context,
	id int,
) (dto.Category, error) {

	query, args, err := sq.
		Select("*").
		From("category").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"category_id": id,
		},
	})

	if err != nil {
		logger.Warnf("unknown error on building sql query: %s", err)

		return dto.Category{}, r.errInternalBuildSql(err)
	}

	category := dto.Category{}

	if err := r.db.GetContext(ctx, &category, query, args...); err != nil {
		if !errpkg.Is(err, sql.ErrNoRows) {
			logger.Warnf("unknown error on getting category: %s", err)

			return dto.Category{}, r.errInternalGetCategory(err)
		}

		logger.Warnf("category not found: %s", err)

		return dto.Category{}, r.errNotFound("category", err)
	}

	return category, nil
}

func (r Repository) Update(
	ctx context.Context,
	data dto.UpdateCategory,
) (int, error) {

	query, args, err := sq.
		Update("category").
		SetMap(map[string]any{
			"name": data.Name,
		}).
		Where(sq.Eq{"id": data.ID}).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"category_id": data.ID,
			"name":        data.Name,
		},
	})

	if err != nil {
		logger.Warnf("unknown error on building sql query: %s", err)

		return 0, r.errInternalBuildSql(err)
	}

	var categoryId int

	if err := r.db.GetContext(ctx, &categoryId, query, args...); err != nil {
		if errpkg.Is(err, sql.ErrNoRows) {
			logger.Warnf("category not found: %s", err)

			return 0, r.errNotFound("category", err)
		}

		if e, ok := err.(*pq.Error); ok {
			switch e.Code {

			case pgerr.UniqueViolation:
				logger.Warnf("category already exists: %s", err)

				return 0, r.errCategoryAlreadyExists(err)

			default:
				logger.Warnf("unknown error on updating category: %s", err)

				return 0, r.errInternalUpdateCategory(err)
			}
		}

		logger.Warnf("unknown error on updating category: %s", err)

		return 0, r.errInternalUpdateCategory(err)
	}

	return categoryId, nil
}

func (r Repository) Delete(
	ctx context.Context,
	category dto.Category,
) (int, error) {

	query, args, err := sq.
		Delete("category").
		Where(sq.Eq{"id": category.ID}).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"category_id": category.ID,
		},
	})

	if err != nil {
		logger.Warnf("unknown error on building sql query: %s", err)

		return 0, r.errInternalBuildSql(err)
	}

	var categoryId int

	if err := r.db.GetContext(ctx, &categoryId, query, args...); err != nil {
		if !errpkg.Is(err, sql.ErrNoRows) {
			logger.Warnf("unknown error on deleting category: %s", err)

			return 0, r.errInternalDeleteCategory(err)
		}

		logger.Warnf("category not found: %s", err)

		return 0, r.errNotFound("category", err)
	}

	return categoryId, nil
}
