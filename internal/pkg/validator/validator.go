package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps go-playground/validator with custom error formatting.
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator.
func New() *Validator {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
	return &Validator{validate: v}
}

// ValidationError represents a single field validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Validate validates a struct and returns formatted errors.
func (v *Validator) Validate(s any) []ValidationError {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	var errors []ValidationError
	for _, e := range err.(validator.ValidationErrors) {
		ns := e.Namespace()
		if dot := strings.Index(ns, "."); dot >= 0 {
			ns = ns[dot+1:]
		}
		errors = append(errors, ValidationError{
			Field:   ns,
			Message: formatError(e),
		})
	}
	return errors
}

func formatError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", e.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", e.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", e.Param())
	case "uuid":
		return "must be a valid UUID"
	case "url":
		return "must be a valid URL"
	default:
		return fmt.Sprintf("failed validation: %s", e.Tag())
	}
}
