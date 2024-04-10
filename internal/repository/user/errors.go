package user

import "github.com/jackvonhouse/product-catalog/internal/repository/errors"

func (r Repository) errUserAlreadyExists(
	err error,
) error {

	return errors.ErrAlreadyExists("user", err)
}

func (r Repository) errNotFound(
	unit string,
	err error,
) error {

	return errors.ErrNotFound(unit, err)
}
