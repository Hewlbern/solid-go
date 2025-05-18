package storage

import (
	"errors"
	"fmt"
)

// ResourceMetadata holds metadata about a resource.
type ResourceMetadata struct {
	// Path is the path to the resource
	Path string
	// ContentType is the MIME type of the resource
	ContentType string
	// Size is the size of the resource in bytes
	Size int64
	// LastModified is the time the resource was last modified
	LastModified string
	// IsContainer indicates whether the resource is a container
	IsContainer bool
}

// Storage defines the interface for storage strategies.
// It implements the Strategy pattern for different storage backends.
type Storage interface {
	// StoreResource saves a resource to storage
	StoreResource(path string, data []byte, contentType string) error

	// GetResource retrieves a resource from storage
	GetResource(path string) ([]byte, string, error)

	// ResourceExists checks if a resource exists
	ResourceExists(path string) (bool, error)

	// DeleteResource removes a resource from storage
	DeleteResource(path string) error

	// GetResourceMetadata gets metadata for a resource
	GetResourceMetadata(path string) (*ResourceMetadata, error)

	// CreateContainer creates a new container
	CreateContainer(path string) error

	// ListContainer lists the contents of a container
	ListContainer(path string) ([]ResourceMetadata, error)

	// StoreACL stores access control rules for a resource
	StoreACL(path string, data []byte) error

	// GetACL retrieves access control rules for a resource
	GetACL(path string) ([]byte, error)
}

// Factory is a factory for creating Storage instances.
// It implements the Factory Method pattern.
type Factory interface {
	// CreateStorage creates a new storage instance
	CreateStorage(strategy string) (Storage, error)
}

// StorageFactory implements the Factory interface.
type StorageFactory struct {
	// BasePath is the base directory for file storage
	BasePath string
}

// NewFactory creates a new StorageFactory.
func NewFactory(basePath string) (Factory, error) {
	if basePath == "" {
		return nil, errors.New("base path cannot be empty")
	}

	return &StorageFactory{
		BasePath: basePath,
	}, nil
}

// CreateStorage creates a new storage instance with the specified strategy.
func (f *StorageFactory) CreateStorage(strategy string) (Storage, error) {
	switch strategy {
	case "memory":
		return NewMemoryStorage(), nil
	case "file":
		return NewFileStorage(f.BasePath)
	default:
		return nil, fmt.Errorf("unknown storage strategy: %s", strategy)
	}
}

// StorageError represents an error in the storage layer.
type StorageError struct {
	// Operation is the operation that failed
	Operation string
	// Path is the path involved in the operation
	Path string
	// Err is the underlying error
	Err error
}

// Error returns a string representation of the error.
func (e *StorageError) Error() string {
	return fmt.Sprintf("storage error during %s on %s: %v", e.Operation, e.Path, e.Err)
}

// Unwrap returns the underlying error.
func (e *StorageError) Unwrap() error {
	return e.Err
}

// ErrResourceNotFound is returned when a resource is not found.
var ErrResourceNotFound = errors.New("resource not found")

// ErrResourceExists is returned when a resource already exists.
var ErrResourceExists = errors.New("resource already exists")

// ErrNotContainer is returned when a path is expected to be a container but isn't.
var ErrNotContainer = errors.New("path is not a container")

// ErrIsContainer is returned when a path is expected not to be a container but is.
var ErrIsContainer = errors.New("path is a container") 