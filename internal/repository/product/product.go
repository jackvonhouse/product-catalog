package product

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
		logger: logger.WithField("unit", "product"),
		db:     db,
	}
}

func (r Repository) Create(
	ctx context.Context,
	data dto.CreateProduct,
	category dto.Category,
) (int, error) {

	rollback := func(tx *sqlx.Tx) error {
		if err := tx.Rollback(); err != nil {
			r.logger.Warnf("unknown error on rollback: %s", err)

			return r.errInternalCreateProduct(err)
		}

		return nil
	}

	commit := func(tx *sqlx.Tx) error {
		if err := tx.Commit(); err != nil {
			r.logger.Warnf("unknown error on commit: %s", err)

			return r.errInternalCreateProduct(err)
		}

		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		r.logger.Warnf("unknown error on starting transaction: %s", err)

		return 0, r.errInternalCreateProduct(err)
	}

	productId, err := r.createProduct(ctx, tx, data)
	if err != nil {
		if rErr := rollback(tx); rErr != nil {
			return 0, rErr
		}

		return 0, err
	}

	if err := r.attachProductToCategory(ctx, tx, productId, category); err != nil {
		if rErr := rollback(tx); rErr != nil {
			return 0, rErr
		}

		return 0, err
	}

	return productId, commit(tx)
}

func (r Repository) createProduct(
	ctx context.Context,
	tx *sqlx.Tx,
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
		logger.Warnf("unknown error on building sql query: %s", err)

		return 0, r.errInternalBuildSql(err)
	}

	var productId int

	if err := tx.GetContext(ctx, &productId, query, args...); err != nil {
		if e, ok := err.(*pq.Error); ok {
			switch e.Code {

			case pgerr.UniqueViolation:
				logger.Warnf("product already exists: %s", err)

				return 0, r.errProductAlreadyExists(err)

			default:
				logger.Warnf("unknown error on creating product: %s", err)

				return 0, r.errInternalCreateProduct(err)
			}
		}

		logger.Warnf("unknown error on creating product: %s", err)

		return 0, r.errInternalCreateProduct(err)
	}

	return productId, nil
}

func (r Repository) attachProductToCategory(
	ctx context.Context,
	tx *sqlx.Tx,
	productId int,
	category dto.Category,
) error {

	query, args, err := sq.
		Insert("product_of_category").
		Columns("product_id", "category_id").
		Values(productId, category.ID).
		Suffix("RETURNING product_id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"product_id":  productId,
			"category_id": category.ID,
		},
	})

	if err != nil {
		logger.Warnf("unknown error on building sql query: %s", err)

		return r.errInternalBuildSql(err)
	}

	var productOfCategoryId int

	if err := tx.GetContext(ctx, &productOfCategoryId, query, args...); err != nil {
		if e, ok := err.(*pq.Error); ok {
			switch e.Code {

			case pgerr.UniqueViolation:
				logger.Warnf("product in category already exists: %s", err)

				return r.errProductInCategoryAlreadyExists(err)

			case pgerr.ForeignKeyViolation:
				table := r.extractTable(e.Detail)

				logger.Warnf("%s not found: %s", table, err)

				return r.errNotFound(table, err)

			default:
				logger.Warnf("unknown error on attaching product to category: %s", err)

				return r.errInternalAttachProductToCategory(err)
			}
		}

		logger.Warnf("unknown error on attaching product to category: %s", err)

		return r.errInternalAttachProductToCategory(err)
	}

	return nil
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

		return []dto.Product{}, r.errInternalBuildSql(err)
	}

	products := make([]dto.Product, 0)

	if err := r.db.SelectContext(ctx, &products, query, args...); err != nil {
		if !errpkg.Is(err, sql.ErrNoRows) {
			logger.Warnf("unknown error on getting products: %s", err)

			return []dto.Product{}, r.errInternalGetProducts(err)
		}

		logger.Warnf("no products: %s", err)

		return []dto.Product{}, r.errNotFound("products", err)
	}

	return products, nil
}

func (r Repository) GetById(
	ctx context.Context,
	id int,
) (dto.Product, error) {

	query, args, err := sq.
		Select("*").
		From("product").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"product_id": id,
		},
	})

	if err != nil {
		logger.Warnf("unknown error on building sql query: %s", err)

		return dto.Product{}, r.errInternalBuildSql(err)
	}

	product := dto.Product{}

	if err := r.db.GetContext(ctx, &product, query, args...); err != nil {
		if !errpkg.Is(err, sql.ErrNoRows) {
			logger.Warnf("unknown error on getting product: %s", err)

			return dto.Product{}, r.errInternalGetProduct(err)
		}

		logger.Warnf("product not found: %s", err)

		return dto.Product{}, r.errNotFound("product", err)
	}

	return product, nil
}

func (r Repository) GetByCategoryId(
	ctx context.Context,
	data dto.GetProduct,
	category dto.Category,
) ([]dto.Product, error) {

	var (
		offset = uint64(data.Offset)
		limit  = uint64(data.Limit)
	)

	query, args, err := sq.
		Select("p.*").
		LeftJoin("product_of_category pc ON pc.product_id = p.id").
		LeftJoin("category c ON c.id = pc.category_id").
		From("product p").
		Where(sq.Eq{"pc.category_id": category.ID}).
		OrderBy("id ASC").
		Offset(offset).
		Limit(limit).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"limit":       data.Limit,
			"offset":      data.Offset,
			"category_id": category.ID,
		},
	})

	if err != nil {
		logger.Warnf("unknown error on building sql query: %s", err)

		return []dto.Product{}, r.errInternalBuildSql(err)
	}

	products := make([]dto.Product, 0)

	if err := r.db.SelectContext(ctx, &products, query, args...); err != nil {
		if !errpkg.Is(err, sql.ErrNoRows) {
			logger.Warnf("unknown error on getting products: %s", err)

			return []dto.Product{}, r.errInternalGetProducts(err)
		}

		logger.Warnf("no products: %s", err)

		return []dto.Product{}, r.errNotFound("products", err)
	}

	return products, nil
}

func (r Repository) Update(
	ctx context.Context,
	data dto.UpdateProduct,
	product dto.Product,
	category dto.Category,
) (int, error) {

	rollback := func(tx *sqlx.Tx) error {
		if err := tx.Rollback(); err != nil {
			r.logger.Warnf("unknown error on rollback: %s", err)

			return r.errInternalUpdateProduct(err)
		}

		return nil
	}

	commit := func(tx *sqlx.Tx) error {
		if err := tx.Commit(); err != nil {
			r.logger.Warnf("unknown error on commit: %s", err)

			return r.errInternalUpdateProduct(err)
		}

		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		r.logger.Warnf("unknown error on starting transaction: %s", err)

		return 0, r.errInternalUpdateProduct(err)
	}

	productId, err := r.updateProduct(ctx, tx, data, product)
	if err != nil {
		if rErr := rollback(tx); rErr != nil {
			return 0, rErr
		}

		return productId, err
	}

	if data.OldCategoryId == data.NewCategoryId {
		return productId, commit(tx)
	}

	if err := r.updateProductCategory(ctx, tx, data, category); err != nil {
		if rErr := rollback(tx); rErr != nil {
			return 0, rErr
		}

		return 0, err
	}

	return productId, commit(tx)
}

func (r Repository) updateProduct(
	ctx context.Context,
	tx *sqlx.Tx,
	data dto.UpdateProduct,
	product dto.Product,
) (int, error) {

	query, args, err := sq.
		Update("product").
		SetMap(map[string]any{
			"name": data.Name,
		}).
		Where(sq.Eq{"id": product.ID}).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"id": product.ID,
			"name": map[string]any{
				"before": product.Name,
				"after":  data.Name,
			},
		},
	})

	if err != nil {
		logger.Warnf("unknown error on building sql query: %s", err)

		return 0, r.errInternalBuildSql(err)
	}

	var productId int

	if err := tx.GetContext(ctx, &productId, query, args...); err != nil {
		if errpkg.Is(err, sql.ErrNoRows) {
			logger.Warnf("product not found: %s", err)

			return 0, r.errNotFound("product", err)
		}

		if e, ok := err.(*pq.Error); ok {
			switch e.Code {

			case pgerr.UniqueViolation:
				logger.Warnf("product already exists: %s", err)

				return 0, r.errProductAlreadyExists(err)

			case pgerr.ForeignKeyViolation:
				table := r.extractTable(e.Detail)

				logger.Warnf("%s not found: %s", table, err)

				return 0, r.errNotFound(table, err)

			default:
				logger.Warnf("unknown error on updating product: %s", err)

				return productId, r.errInternalUpdateProduct(err)
			}
		}

		logger.Warnf("unknown error on updating product: %s", err)

		return productId, r.errInternalUpdateProduct(err)
	}

	return productId, nil
}

func (r Repository) updateProductCategory(
	ctx context.Context,
	tx *sqlx.Tx,
	data dto.UpdateProduct,
	category dto.Category,
) error {

	query, args, err := sq.
		Update("product_of_category").
		SetMap(map[string]any{
			"category_id": category.ID,
		}).
		Where(sq.And{
			sq.Eq{"category_id": data.OldCategoryId},
			sq.Eq{"product_id": data.ID},
		}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"id": data.ID,
			"category": map[string]any{
				"before": data.OldCategoryId,
				"after":  category.ID,
			},
		},
	})

	if err != nil {
		logger.Warnf("unknown error on building sql query: %s", err)

		return r.errInternalBuildSql(err)
	}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		if errpkg.Is(err, sql.ErrNoRows) {
			logger.Warnf("product in category not found: %s", err)

			return r.errNotFound("product in category", err)
		}

		if e, ok := err.(*pq.Error); ok {
			switch e.Code {

			case pgerr.UniqueViolation:
				logger.Warnf("product in category already exists: %s", err)

				return r.errProductInCategoryAlreadyExists(err)

			case pgerr.ForeignKeyViolation:
				table := r.extractTable(e.Detail)

				logger.Warnf("%s not found: %s", table, err)

				return r.errNotFound(table, err)

			default:
				logger.Warnf("unknown error on updating product in category: %s", err)

				return r.errInternalUpdateProductInCategory(err)
			}
		}

		logger.Warnf("unknown error on updating product in category: %s", err)

		return r.errInternalUpdateProductInCategory(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		logger.Warnf("category not found")

		return r.errNotFound("category", err)
	}

	return nil
}

func (r Repository) Delete(
	ctx context.Context,
	product dto.Product,
) (int, error) {

	query, args, err := sq.
		Delete("product").
		Where(sq.Eq{"id": product.ID}).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	logger := r.logger.WithFields(map[string]any{
		"query": query,
		"args": map[string]any{
			"product_id": product.ID,
		},
	})

	if err != nil {
		logger.Warnf("unknown error on building sql query: %s", err)

		return 0, r.errInternalBuildSql(err)
	}

	var productId int

	if err := r.db.GetContext(ctx, &productId, query, args...); err != nil {
		if !errpkg.Is(err, sql.ErrNoRows) {
			logger.Warnf("unknown error on deleting product: %s", err)

			return 0, r.errInternalDeleteProduct(err)
		}

		logger.Warnf("product not found: %s", err)

		return 0, r.errNotFound("product", err)
	}

	return productId, nil
}
