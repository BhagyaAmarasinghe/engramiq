package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"-"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

func (e AppError) Error() string {
	return e.Message
}

// Common error constructors

func NewBadRequest(message string, details ...map[string]interface{}) AppError {
	return AppError{
		Code:       "BAD_REQUEST",
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Details:    mergeDetails(details...),
	}
}

func NewUnauthorized(message string) AppError {
	return AppError{
		Code:       "UNAUTHORIZED",
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

func NewNotFound(resource string, id string) AppError {
	return AppError{
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s not found", resource),
		StatusCode: http.StatusNotFound,
		Details: map[string]interface{}{
			"resource": resource,
			"id":       id,
		},
	}
}

func NewValidationError(errors []ValidationError) AppError {
	details := make([]map[string]string, len(errors))
	for i, err := range errors {
		details[i] = map[string]string{
			"field":   err.Field,
			"message": err.Message,
		}
	}

	return AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "Invalid input data",
		StatusCode: http.StatusUnprocessableEntity,
		Details: map[string]interface{}{
			"errors": details,
		},
	}
}

func NewInternal(message string) AppError {
	return AppError{
		Code:       "INTERNAL_ERROR",
		Message:    message,
		StatusCode: http.StatusInternalServerError,
	}
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func mergeDetails(details ...map[string]interface{}) map[string]interface{} {
	if len(details) == 0 {
		return nil
	}
	return details[0]
}

func IsAppError(err error) (AppError, bool) {
	appErr, ok := err.(AppError)
	return appErr, ok
}