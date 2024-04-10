package validator

import (
	"fmt"
	"github.com/jackvonhouse/product-catalog/internal/errors"
	"unicode"
)

const (
	minUserNameLen = 4
	maxUserNameLen = 31
)

type ValidatorFunc func(rune) bool

func IsValidUsername(
	userName string,
) error {

	userNameLen := len(userName)

	if userNameLen > maxUserNameLen || userNameLen < minUserNameLen {
		return errors.ErrInvalid.New(
			fmt.Sprintf("username length must be between %d and %d (actual %d)",
				minUserNameLen, maxUserNameLen, userNameLen,
			),
		)
	}

	firstSymbol := rune(userName[0])

	if !unicode.IsLetter(firstSymbol) {
		return errors.ErrInvalid.New("username first character must be letter")
	}

	if !unicode.Is(unicode.Latin, firstSymbol) {
		return errors.ErrInvalid.New("username must contains only latin chars")
	}

	for _, symbol := range userName {
		isLatin := unicode.Is(unicode.Latin, symbol)
		isLetter := unicode.IsLetter(symbol)
		isNumber := unicode.IsNumber(symbol)
		isUnderscore := isUnderScore(symbol)

		if !isLatin && !isLetter && !isNumber && !isUnderscore {
			return errors.ErrInvalid.New(
				"username must contains only latin chars, numbers and underscores",
			)
		}
	}

	return nil
}

func IsValidCredentials(username, _ string) error {
	return IsValidUsername(username)
}

func isUnderScore(r rune) bool {
	return r == '_'
}
