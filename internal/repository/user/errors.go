package user

import (
	"github.com/jackvonhouse/product-catalog/internal/repository/errors"
	"strings"
)

func (r Repository) errInternalCreateUser(
	err error,
) error {

	return errors.ErrInternal("creating", "user", err)
}

func (r Repository) errInternalGetUser(
	err error,
) error {

	return errors.ErrInternal("getting", "user", err)
}

func (r Repository) errInternalBuildSql(
	err error,
) error {

	return errors.ErrInternal("building", "sql query", err)
}

func (r Repository) errInternalDeleteProduct(
	err error,
) error {

	return errors.ErrInternal("deleting", "product", err)
}

func (r Repository) errUserAlreadyExists(
	err error,
) error {

	return errors.ErrAlreadyExists("user", err)
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
