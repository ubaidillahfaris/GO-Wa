package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType represents the type of error
type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "VALIDATION_ERROR"
	ErrorTypeNotFound       ErrorType = "NOT_FOUND"
	ErrorTypeUnauthorized   ErrorType = "UNAUTHORIZED"
	ErrorTypeForbidden      ErrorType = "FORBIDDEN"
	ErrorTypeConflict       ErrorType = "CONFLICT"
	ErrorTypeInternal       ErrorType = "INTERNAL_ERROR"
	ErrorTypeWhatsApp       ErrorType = "WHATSAPP_ERROR"
	ErrorTypeDatabase       ErrorType = "DATABASE_ERROR"
	ErrorTypeConnection     ErrorType = "CONNECTION_ERROR"

	// Aliases for backward compatibility
	ErrTypeValidation   = ErrorTypeValidation
	ErrTypeNotFound     = ErrorTypeNotFound
	ErrTypeUnauthorized = ErrorTypeUnauthorized
	ErrTypeForbidden    = ErrorTypeForbidden
	ErrTypeConflict     = ErrorTypeConflict
	ErrTypeInternal     = ErrorTypeInternal
	ErrTypeWhatsApp     = ErrorTypeWhatsApp
	ErrTypeDatabase     = ErrorTypeDatabase
	ErrTypeConnection   = ErrorTypeConnection
)

// AppError represents a custom application error
type AppError struct {
	Type       ErrorType              `json:"type"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"-"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Err        error                  `json:"-"`
}

// CustomError is an alias for AppError for backward compatibility
type CustomError = AppError

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// New creates a new AppError
func New(errType ErrorType, message string) *AppError {
	return &AppError{
		Type:       errType,
		Message:    message,
		StatusCode: getStatusCode(errType),
	}
}

// Wrap wraps an existing error with AppError
func Wrap(err error, errType ErrorType, message string) *AppError {
	return &AppError{
		Type:       errType,
		Message:    message,
		StatusCode: getStatusCode(errType),
		Err:        err,
	}
}

// getStatusCode maps error types to HTTP status codes
func getStatusCode(errType ErrorType) int {
	switch errType {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case ErrorTypeForbidden:
		return http.StatusForbidden
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeWhatsApp, ErrorTypeConnection:
		return http.StatusServiceUnavailable
	case ErrorTypeDatabase, ErrorTypeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from error
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	// If not an AppError, wrap it as internal error
	return Wrap(err, ErrorTypeInternal, "Internal server error")
}

// Common error constructors
func NewValidationError(message string) *AppError {
	return New(ErrorTypeValidation, message)
}

func NewNotFoundError(resource string) *AppError {
	return New(ErrorTypeNotFound, fmt.Sprintf("%s not found", resource))
}

func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "Unauthorized access"
	}
	return New(ErrorTypeUnauthorized, message)
}

func NewWhatsAppError(message string, err error) *AppError {
	return Wrap(err, ErrorTypeWhatsApp, message)
}

func NewDatabaseError(message string, err error) *AppError {
	return Wrap(err, ErrorTypeDatabase, message)
}

func NewConnectionError(message string, err error) *AppError {
	return Wrap(err, ErrorTypeConnection, message)
}

func NewInternalError(message string, err error) *AppError {
	return Wrap(err, ErrorTypeInternal, message)
}

// IsNotFound checks if error is a not found error
func IsNotFound(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeNotFound
	}
	return false
}
