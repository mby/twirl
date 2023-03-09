package shared

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	passwordvalidator "github.com/wagslane/go-password-validator"
)

type Username string

func ValidateUsername(input string) (Username, *Error) {
	input = strings.TrimSpace(input)

	if len(input) < 2 {
		return "", NewError(http.StatusBadRequest, 0, "username must be at least 3 characters")
	}

	regex := "^(\\w){1,15}$"
	matched, e := regexp.MatchString(regex, input)
	if e != nil {
		return "", NewError(http.StatusBadRequest, 0, e.Error())
	}
	if !matched {
		return "", NewError(http.StatusBadRequest, 0, "invalid username, try a twitter name maybe?")
	}

	return Username(input), nil
}

type Password string

func ValidatePassword(input string) (Password, *Error) {
	input = strings.TrimSpace(input)

	if len(input) < 8 {
		return "", NewError(http.StatusBadRequest, 0, "password must be at least 8 characters")
	}

	// check entropy
	const minEntropy = 60
	e := passwordvalidator.Validate(input, minEntropy)
	if e != nil {
		msg := fmt.Sprintf("password is not complex enough: %v", e)
		return "", NewError(http.StatusBadRequest, 0, msg)
	}

	return Password(input), nil
}
