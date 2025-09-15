package validator

import (
	"reflect"
	"strings"

	"github.com/engramiq/engramiq-backend/pkg/errors"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Use JSON tag names in validation errors
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidateStruct validates a struct and returns formatted errors
func ValidateStruct(s interface{}) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrors []errors.ValidationError
	
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			validationErrors = append(validationErrors, errors.ValidationError{
				Field:   e.Field(),
				Message: getErrorMessage(e),
			})
		}
	}

	return errors.NewValidationError(validationErrors)
}

func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Must be at least " + e.Param() + " characters long"
	case "max":
		return "Must be at most " + e.Param() + " characters long"
	case "uuid":
		return "Must be a valid UUID"
	default:
		return "Invalid value"
	}
}