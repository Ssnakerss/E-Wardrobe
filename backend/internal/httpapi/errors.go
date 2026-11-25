package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ErrorResponse represents a unified API error.
type ErrorResponse struct {
	Error struct {
		Code    string                 `json:"code"`
		Message string                 `json:"message"`
		Details map[string]interface{} `json:"details,omitempty"`
	} `json:"error"`
}

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

// NewAppError creates a new AppError with the given code, message and status.
func NewAppError(code string, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// BadRequest returns a 400 bad request error.
func BadRequest(message string) *AppError {
	return NewAppError("bad_request", message, http.StatusBadRequest)
}

// Unauthorized returns a 401 unauthorized error.
func Unauthorized(message string) *AppError {
	return NewAppError("unauthorized", message, http.StatusUnauthorized)
}

// Forbidden returns a 403 forbidden error.
func Forbidden(message string) *AppError {
	return NewAppError("forbidden", message, http.StatusForbidden)
}

// NotFound returns a 404 not found error.
func NotFound(message string) *AppError {
	return NewAppError("not_found", message, http.StatusNotFound)
}

// ValidationError returns a 422 validation error.
func ValidationError(message string) *AppError {
	return NewAppError("validation_failed", message, http.StatusUnprocessableEntity)
}

// Conflict returns a 409 conflict error.
func Conflict(message string) *AppError {
	return NewAppError("conflict", message, http.StatusConflict)
}

// WriteError writes an AppError as a JSON response.
func WriteError(w http.ResponseWriter, err error) {
	var appErr *AppError
	switch e := err.(type) {
	case *AppError:
		appErr = e
	default:
		appErr = &AppError{
			Code:    "internal_error",
			Message: "An unexpected error occurred",
			Status:  http.StatusInternalServerError,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.Status)

	resp := ErrorResponse{
		Error: struct {
			Code    string                 `json:"code"`
			Message string                 `json:"message"`
			Details map[string]interface{} `json:"details,omitempty"`
		}{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
	}

	_ = json.NewEncoder(w).Encode(resp)
}
