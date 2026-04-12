package validation

import (
	"errors"
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
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return fmt.Errorf("errore di validazione interno: %w", err)
	}
	var fieldErrors FieldErrors
	for _, e := range ve {
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
		return "è obbligatorio"
	case "email":
		return "deve essere un indirizzo email valido"
	case "min":
		return fmt.Sprintf("deve essere almeno %s", e.Param())
	case "max":
		return fmt.Sprintf("deve essere al massimo %s", e.Param())
	case "oneof":
		return fmt.Sprintf("deve essere uno tra: %s", e.Param())
	default:
		return fmt.Sprintf("validazione fallita: %s", e.Tag())
	}
}
