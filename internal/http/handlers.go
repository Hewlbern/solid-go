package http

import (
	"context"
	"solid-go/internal/server"
)

// GetOperationHandler handles GET operations
type GetOperationHandler struct {
	storage Storage
}

// NewGetOperationHandler creates a new GET operation handler
func NewGetOperationHandler(storage Storage) *GetOperationHandler {
	return &GetOperationHandler{
		storage: storage,
	}
}

// Handle implements OperationHandler
func (h *GetOperationHandler) Handle(ctx context.Context, op server.Operation) (*server.Representation, error) {
	data, err := h.storage.Get(ctx, op.Target)
	if err != nil {
		return nil, err
	}

	return &server.Representation{
		Data: data,
		Metadata: map[string]string{
			"Content-Type": "text/turtle", // Default to Turtle format
		},
	}, nil
}

// PutOperationHandler handles PUT operations
type PutOperationHandler struct {
	storage Storage
}

// NewPutOperationHandler creates a new PUT operation handler
func NewPutOperationHandler(storage Storage) *PutOperationHandler {
	return &PutOperationHandler{
		storage: storage,
	}
}

// Handle implements OperationHandler
func (h *PutOperationHandler) Handle(ctx context.Context, op server.Operation) (*server.Representation, error) {
	err := h.storage.Put(ctx, op.Target, op.Body)
	if err != nil {
		return nil, err
	}

	return &server.Representation{
		Metadata: map[string]string{
			"Content-Type": "text/turtle",
		},
	}, nil
}

// DeleteOperationHandler handles DELETE operations
type DeleteOperationHandler struct {
	storage Storage
}

// NewDeleteOperationHandler creates a new DELETE operation handler
func NewDeleteOperationHandler(storage Storage) *DeleteOperationHandler {
	return &DeleteOperationHandler{
		storage: storage,
	}
}

// Handle implements OperationHandler
func (h *DeleteOperationHandler) Handle(ctx context.Context, op server.Operation) (*server.Representation, error) {
	err := h.storage.Delete(ctx, op.Target)
	if err != nil {
		return nil, err
	}

	return &server.Representation{}, nil
}

// Storage interface for resource storage
type Storage interface {
	Get(ctx context.Context, path string) ([]byte, error)
	Put(ctx context.Context, path string, data []byte) error
	Delete(ctx context.Context, path string) error
}
