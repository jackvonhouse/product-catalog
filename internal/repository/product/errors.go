package product

import (
	"github.com/jackvonhouse/product-catalog/internal/repository/errors"
	"strings"
)

func (r Repository) errInternalCreateProduct(
	err error,
) error {

	return errors.ErrInternal("creating", "product", err)
}

func (r Repository) errInternalGetProducts(
	err error,
) error {

	return errors.ErrInternal("getting", "products", err)
}

func (r Repository) errInternalGetProduct(
	err error,
) error {

	return errors.ErrInternal("getting", "product", err)
}

func (r Repository) errInternalAttachProductToCategory(
	err error,
) error {

	return errors.ErrInternal("attaching", "product to category", err)
}

func (r Repository) errInternalBuildSql(
	err error,
) error {

	return errors.ErrInternal("building", "sql query", err)
}

func (r Repository) errInternalUpdateProduct(
	err error,
) error {

	return errors.ErrInternal("updating", "product", err)
}

func (r Repository) errInternalUpdateProductInCategory(
	err error,
) error {

	return errors.ErrInternal("updating", "product in category", err)
}

func (r Repository) errInternalDeleteProduct(
	err error,
) error {

	return errors.ErrInternal("deleting", "product", err)
}

func (r Repository) errProductAlreadyExists(
	err error,
) error {

	return errors.ErrAlreadyExists("product", err)
}

func (r Repository) errProductInCategoryAlreadyExists(
	err error,
) error {

	return errors.ErrAlreadyExists("product in category", err)
}

func (r Repository) errNotFound(
	unit string,
	err error,
) error {

	return errors.ErrNotFound(unit, err)
}

func (r Repository) extractTable(
	errMsg string,
) string {

	start := strings.Index(errMsg, `"`)

	if start == -1 {
		return "unknown"
	}

	start++

	end := strings.Index(errMsg[start:], `"`)
	if start == -1 {
		return "unknown"
	}

	return errMsg[start : start+end]
}
