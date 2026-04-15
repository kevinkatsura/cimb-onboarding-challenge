package validator

import (
	"payment-initiation-acquiring-service/pkg/apperror"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateStruct validates a struct's fields based on `validate` tags.
// Returns an AppError with ErrInvalidFieldFormat if validation fails.
func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var msgs []string
			for _, fe := range validationErrors {
				msgs = append(msgs, fmt.Sprintf("%s: failed on '%s'", fe.Field(), fe.Tag()))
			}
			return apperror.New(apperror.ErrInvalidFieldFormat, strings.Join(msgs, "; "))
		}
		return apperror.Wrap(apperror.ErrInvalidFieldFormat, "validation error", err)
	}
	return nil
}
