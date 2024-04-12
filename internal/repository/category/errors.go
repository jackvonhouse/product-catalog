package category

import (
	"github.com/jackvonhouse/product-catalog/internal/repository/errors"
	"strings"
)

func (r Repository) errInternalCreateCategory(
	err error,
) error {

	return errors.ErrInternal("creating", "category", err)
}

func (r Repository) errInternalGetCategories(
	err error,
) error {

	return errors.ErrInternal("getting", "categories", err)
}

func (r Repository) errInternalGetCategory(
	err error,
) error {

	return errors.ErrInternal("getting", "category", err)
}

func (r Repository) errInternalAttachCategoryToCategory(
	err error,
) error {

	return errors.ErrInternal("attaching", "product to category", err)
}

func (r Repository) errInternalBuildSql(
	err error,
) error {

	return errors.ErrInternal("building", "sql query", err)
}

func (r Repository) errInternalUpdateCategory(
	err error,
) error {

	return errors.ErrInternal("updating", "category", err)
}

func (r Repository) errInternalUpdateCategoryInCategory(
	err error,
) error {

	return errors.ErrInternal("updating", "category in category", err)
}

func (r Repository) errInternalDeleteCategory(
	err error,
) error {

	return errors.ErrInternal("deleting", "category", err)
}

func (r Repository) errCategoryAlreadyExists(
	err error,
) error {

	return errors.ErrAlreadyExists("category", err)
}

func (r Repository) errCategoryInCategoryAlreadyExists(
	err error,
) error {

	return errors.ErrAlreadyExists("category in category", err)
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
