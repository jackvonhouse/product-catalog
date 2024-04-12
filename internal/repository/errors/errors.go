package errors

import (
	"fmt"
	"github.com/jackvonhouse/product-catalog/internal/errors"
)

func ErrInternal(
	action string,
	unit string,
	err error,
) error {

	return errors.
		ErrInternal.
		New(fmt.Sprintf("unknown error on %s %s", action, unit)).
		Wrap(err)
}

func ErrNotFound(
	unit string,
	err error,
) error {

	return errors.
		ErrNotFound.
		New(fmt.Sprintf("%s not found", unit)).
		Wrap(err)
}

func ErrAlreadyExists(
	unit string,
	err error,
) error {

	return errors.
		ErrAlreadyExists.
		New(fmt.Sprintf("%s already exists", unit)).
		Wrap(err)
}

func ErrExpired(
	unit string,
	err error,
) error {

	return errors.
		ErrExpired.
		New(fmt.Sprintf("%s expired", unit)).
		Wrap(err)
}
