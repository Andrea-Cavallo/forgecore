package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v = validator.New(validator.WithRequiredStructFields())

// FieldError represents a single validation error on a struct field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FieldErrors is a slice of FieldError that implements the error interface.
type FieldErrors []FieldError

func (fe FieldErrors) Error() string {
	msgs := make([]string, len(fe))
	for i, e := range fe {
		msgs[i] = fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return strings.Join(msgs, "; ")
}

// Validate runs struct validation and returns FieldErrors or nil.
func Validate(input any) error {
	err := v.Struct(input)
	if err == nil {
		return nil
	}
	var fieldErrors FieldErrors
	for _, e := range err.(validator.ValidationErrors) {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   strings.ToLower(e.Field()),
			Message: humanMessage(e),
		})
	}
	return fieldErrors
}

func humanMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s", e.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", e.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", e.Param())
	default:
		return fmt.Sprintf("failed validation: %s", e.Tag())
	}
}
