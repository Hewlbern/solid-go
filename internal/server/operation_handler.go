package server

import (
	"context"
	"io"
	"net/http"
)

// OperationHandler handles HTTP operations
type OperationHandler interface {
	Handle(ctx context.Context, op Operation) (*Representation, error)
}

// OperationHttpHandler handles HTTP operations
type OperationHttpHandler interface {
	HttpHandler
	// HandleOperation processes an operation and returns a representation
	HandleOperation(ctx context.Context, op Operation) (*Representation, error)
}

// BaseOperationHandler is a base implementation of OperationHandler
type BaseOperationHandler struct {
	handler OperationHandler
}

// NewBaseOperationHandler creates a new BaseOperationHandler
func NewBaseOperationHandler(handler OperationHandler) *BaseOperationHandler {
	return &BaseOperationHandler{
		handler: handler,
	}
}

// Handle implements HttpHandler
func (h *BaseOperationHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	// Convert HTTP request to operation
	op := Operation{
		Method:      r.Method,
		Target:      r.URL.Path,
		ContentType: r.Header.Get("Content-Type"),
		Headers:     r.Header,
	}

	// Read request body if present
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}
		op.Body = &Representation{
			Data: body,
		}
	}

	// Handle the operation
	repr, err := h.handler.Handle(r.Context(), op)
	if err != nil {
		return err
	}

	// Write response
	if repr != nil {
		for k, v := range repr.Metadata {
			w.Header().Set(k, v)
		}
		w.Write(repr.Data)
	}

	return nil
}

// HandleOperation handles an operation
func (h *BaseOperationHandler) HandleOperation(ctx context.Context, op Operation) (*Representation, error) {
	return h.handler.Handle(ctx, op)
}
