package transport

import (
	"net/http"
	"strconv"

	"github.com/jackvonhouse/product-catalog/internal/errors"
	errpkg "github.com/jackvonhouse/product-catalog/pkg/errors"
)

func StringToInt(valueStr string) (int, error) {
	valueInt, err := strconv.Atoi(valueStr)

	if err != nil {
		return 0, err
	}

	return valueInt, nil
}

var defaultErrorHttpCodes = map[uint32]int{
	errors.ErrInternal.TypeId:      http.StatusInternalServerError,
	errors.ErrAlreadyExists.TypeId: http.StatusConflict,
	errors.ErrNotFound.TypeId:      http.StatusNotFound,
}

func ErrorToHttpResponse(
	err error,
) (int, string) {

	if errpkg.Has(err, errors.ErrInternal) {
		return http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError)
	}

	code := defaultErrorHttpCodes[errpkg.TypeId(err)]

	if code == 0 {
		return http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError)
	}

	return code, err.Error()
}
