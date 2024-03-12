package errors

import (
	errors "errors"
)

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func Wrap(err error, wrapper *Instance) *Instance {
	wrapper.Err = err

	return wrapper
}

func Unwrap(err error) error {
	switch e := err.(type) {
	case *Instance:
		return e.Err
	default:
		return nil
	}
}

func TypeId(err error) uint32 {
	switch e := err.(type) {
	case *Instance:
		return e.TypeId
	default:
		return 0
	}
}

func TypeIs(err error, t *Type) bool {
	switch e := err.(type) {
	case *Instance:
		return e.TypeId == t.TypeId
	default:
		return false
	}
}

func Has(err error, t *Type) bool {
	for {
		if err == nil {
			return false
		}

		if TypeIs(err, t) {
			return true
		}

		err = Unwrap(err)
	}
}
