package server

import (
	"context"
)

// Authorizer defines the interface for authorization
type Authorizer interface {
	// Authorize checks if an operation is authorized
	Authorize(ctx context.Context, op Operation) error
}
