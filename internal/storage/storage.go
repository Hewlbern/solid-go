package storage

import (
	"context"
)

// Storage defines the interface for data storage
type Storage interface {
	// Get retrieves data from storage
	Get(ctx context.Context, path string) ([]byte, error)

	// Put stores data in storage
	Put(ctx context.Context, path string, data []byte) error

	// Delete removes data from storage
	Delete(ctx context.Context, path string) error

	// List returns a list of resources in a container
	List(ctx context.Context, path string) ([]string, error)
}

// FileStorage implements the Storage interface using the filesystem
type FileStorage struct {
	basePath string
}

// NewFileStorage creates a new file-based storage
func NewFileStorage(basePath string) *FileStorage {
	return &FileStorage{
		basePath: basePath,
	}
}

// Get retrieves data from a file
func (s *FileStorage) Get(ctx context.Context, path string) ([]byte, error) {
	// TODO: Implement file reading
	return nil, nil
}

// Put stores data in a file
func (s *FileStorage) Put(ctx context.Context, path string, data []byte) error {
	// TODO: Implement file writing
	return nil
}

// Delete removes a file
func (s *FileStorage) Delete(ctx context.Context, path string) error {
	// TODO: Implement file deletion
	return nil
}

// List returns a list of files in a directory
func (s *FileStorage) List(ctx context.Context, path string) ([]string, error) {
	// TODO: Implement directory listing
	return nil, nil
}
