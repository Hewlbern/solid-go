package server

import (
	"context"
)

// Authorizer defines the interface for authorization
type Authorizer interface {
	// Authorize checks if an operation is authorized
	Authorize(ctx context.Context, op Operation) error
}

// Operation represents an HTTP operation
type Operation struct {
	Method      string
	Target      string
	ContentType string
	Headers     map[string][]string
}

// NewOperation creates a new Operation
func NewOperation(method, target, contentType string, headers map[string][]string) Operation {
	return Operation{
		Method:      method,
		Target:      target,
		ContentType: contentType,
		Headers:     headers,
	}
}
