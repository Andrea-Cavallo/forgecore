package configschema

import (
	"fmt"
	"strconv"
)

type Kind string

const (
	String Kind = "string"
	Bool   Kind = "bool"
	Int    Kind = "int"
)

type Field struct {
	Key      string
	Default  string
	Kind     Kind
	Required bool
	Secret   bool
}

type Schema []Field

type Secret string

func (s Secret) String() string {
	if s == "" {
		return ""
	}
	return "[redatto]"
}

func (s Secret) Value() string {
	return string(s)
}

func (s Schema) Keys() []string {
	keys := make([]string, 0, len(s))
	for _, field := range s {
		keys = append(keys, field.Key)
	}
	return keys
}

func (s Schema) Defaults() map[string]string {
	defaults := make(map[string]string, len(s))
	for _, field := range s {
		if field.Default == "" {
			continue
		}
		defaults[field.Key] = field.Default
	}
	return defaults
}

func (s Schema) Validate(values map[string]string) error {
	for _, field := range s {
		if err := validateField(field, values[field.Key]); err != nil {
			return err
		}
	}
	return nil
}

func validateField(field Field, value string) error {
	if field.Required && value == "" {
		return fmt.Errorf("configurazione %s mancante", field.Key)
	}
	if value == "" {
		return nil
	}
	switch field.Kind {
	case Bool:
		return validateBool(field.Key, value)
	case Int:
		return validateInt(field.Key, value)
	default:
		return nil
	}
}

func validateBool(key, value string) error {
	if _, err := strconv.ParseBool(value); err != nil {
		return fmt.Errorf("configurazione %s non booleana: %w", key, err)
	}
	return nil
}

func validateInt(key, value string) error {
	if _, err := strconv.Atoi(value); err != nil {
		return fmt.Errorf("configurazione %s non intera: %w", key, err)
	}
	return nil
}
