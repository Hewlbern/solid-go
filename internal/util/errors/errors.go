package errors

import (
	"fmt"
	"strings"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// Error types
	ValidationError   ErrorType = "ValidationError"
	NotFoundError     ErrorType = "NotFoundError"
	UnauthorizedError ErrorType = "UnauthorizedError"
	ForbiddenError    ErrorType = "ForbiddenError"
	ConflictError     ErrorType = "ConflictError"
	InternalError     ErrorType = "InternalError"
)

// CustomError represents a custom error with type and message
type CustomError struct {
	Type    ErrorType
	Message string
	Err     error
}

// Error implements the error interface
func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the wrapped error
func (e *CustomError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error
func NewValidationError(message string, err error) error {
	return &CustomError{
		Type:    ValidationError,
		Message: message,
		Err:     err,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string, err error) error {
	return &CustomError{
		Type:    NotFoundError,
		Message: message,
		Err:     err,
	}
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string, err error) error {
	return &CustomError{
		Type:    UnauthorizedError,
		Message: message,
		Err:     err,
	}
}

// NewForbiddenError creates a new forbidden error
func NewForbiddenError(message string, err error) error {
	return &CustomError{
		Type:    ForbiddenError,
		Message: message,
		Err:     err,
	}
}

// NewConflictError creates a new conflict error
func NewConflictError(message string, err error) error {
	return &CustomError{
		Type:    ConflictError,
		Message: message,
		Err:     err,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(message string, err error) error {
	return &CustomError{
		Type:    InternalError,
		Message: message,
		Err:     err,
	}
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	return isErrorType(err, ValidationError)
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	return isErrorType(err, NotFoundError)
}

// IsUnauthorizedError checks if an error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	return isErrorType(err, UnauthorizedError)
}

// IsForbiddenError checks if an error is a forbidden error
func IsForbiddenError(err error) bool {
	return isErrorType(err, ForbiddenError)
}

// IsConflictError checks if an error is a conflict error
func IsConflictError(err error) bool {
	return isErrorType(err, ConflictError)
}

// IsInternalError checks if an error is an internal error
func IsInternalError(err error) bool {
	return isErrorType(err, InternalError)
}

// isErrorType checks if an error is of a specific type
func isErrorType(err error, errorType ErrorType) bool {
	if err == nil {
		return false
	}
	if customErr, ok := err.(*CustomError); ok {
		return customErr.Type == errorType
	}
	return false
}

// GetErrorType returns the type of an error
func GetErrorType(err error) ErrorType {
	if customErr, ok := err.(*CustomError); ok {
		return customErr.Type
	}
	return InternalError
}

// GetErrorMessage returns the message of an error
func GetErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	if customErr, ok := err.(*CustomError); ok {
		return customErr.Message
	}
	return err.Error()
}

// GetErrorStack returns the error stack as a string
func GetErrorStack(err error) string {
	if err == nil {
		return ""
	}
	var stack []string
	for err != nil {
		stack = append(stack, err.Error())
		if customErr, ok := err.(*CustomError); ok {
			err = customErr.Err
		} else {
			break
		}
	}
	return strings.Join(stack, " -> ")
}
