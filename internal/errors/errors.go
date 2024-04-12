package errors

import "github.com/jackvonhouse/product-catalog/pkg/errors"

var (
	ErrInternal      = errors.NewType("internal error")
	ErrNotFound      = errors.NewType("not found")
	ErrAlreadyExists = errors.NewType("already exists")
	ErrInvalid       = errors.NewType("invalid data")
	ErrExpired       = errors.NewType("expired")
	ErrInvalidToken  = errors.NewType("invalid token")
)
