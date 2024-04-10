package refresh

import (
	"github.com/jackvonhouse/product-catalog/internal/repository/errors"
)

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
