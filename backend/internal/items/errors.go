package items

import (
	"fmt"
	"net/http"
)

// AppError is an application-level error with a code and message.
type AppError struct {
	Code    string
	Message string
	Status  int
	Details map[string]interface{}
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ValidationError returns a 422 validation error.
func ValidationError(message string) *AppError {
	return &AppError{
		Code:    "validation_failed",
		Message: message,
		Status:  http.StatusUnprocessableEntity,
	}
}

// ConflictError returns a 409 conflict error.
func ConflictError(message string) *AppError {
	return &AppError{
		Code:    "conflict",
		Message: message,
		Status:  http.StatusConflict,
	}
}
