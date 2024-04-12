package refresh

import (
	"github.com/jackvonhouse/product-catalog/internal/repository/errors"
)

func (r Repository) errInternalBuildSql(
	err error,
) error {

	return errors.ErrInternal("building", "sql query", err)
}

func (r Repository) errInternalCreateRefresh(
	err error,
) error {

	return errors.ErrInternal("creating", "refresh token", err)
}

func (r Repository) errInternalGetRefresh(
	err error,
) error {

	return errors.ErrInternal("getting", "refresh token", err)
}

func (r Repository) errInternalDeleteRefresh(
	err error,
) error {

	return errors.ErrInternal("creating", "refresh token", err)
}

func (r Repository) errRefreshAlreadyExists(
	err error,
) error {

	return errors.ErrAlreadyExists("refresh token", err)
}

func (r Repository) errNotFound(
	unit string,
	err error,
) error {

	return errors.ErrNotFound(unit, err)
}

func (r Repository) errExpired(
	unit string,
	err error,
) error {

	return errors.ErrExpired(unit, err)
}
