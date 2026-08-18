package validator

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

// TranslateError returns the translated message of the validation error.
func TranslateError(errs validator.ValidationErrors) string {
	messages := make([]string, len(errs))
	for i, err := range errs {
		messages[i] = err.Translate(tran)
	}

	return strings.Join(messages, "; ")
}
