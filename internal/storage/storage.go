package storage

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Storage defines the interface for storage operations
type Storage interface {
	// Get retrieves data from storage
	Get(ctx context.Context, path string) ([]byte, error)

	// Put stores data in storage
	Put(ctx context.Context, path string, data []byte) error

	// Delete removes data from storage
	Delete(ctx context.Context, path string) error

	// List returns a list of resources in a container
	List(ctx context.Context, path string) ([]string, error)

	// Exists checks if a resource exists
	Exists(ctx context.Context, path string) (bool, error)
}

// FileStorage implements Storage using the local filesystem
type FileStorage struct {
	rootPath string
}

// NewFileStorage creates a new FileStorage instance
func NewFileStorage(rootPath string) (*FileStorage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(rootPath, 0755); err != nil {
		return nil, err
	}

	return &FileStorage{
		rootPath: rootPath,
	}, nil
}

// Get implements Storage.Get
func (s *FileStorage) Get(ctx context.Context, path string) ([]byte, error) {
	fullPath := filepath.Join(s.rootPath, path)
	return ioutil.ReadFile(fullPath)
}

// Put implements Storage.Put
func (s *FileStorage) Put(ctx context.Context, path string, data []byte) error {
	fullPath := filepath.Join(s.rootPath, path)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(fullPath, data, 0644)
}

// Delete implements Storage.Delete
func (s *FileStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.rootPath, path)
	return os.Remove(fullPath)
}

// List implements Storage.List
func (s *FileStorage) List(ctx context.Context, path string) ([]string, error) {
	fullPath := filepath.Join(s.rootPath, path)

	// Read directory contents
	entries, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var resources []string
	for _, entry := range entries {
		// Convert to relative path
		relPath := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			relPath += "/"
		}
		resources = append(resources, relPath)
	}

	return resources, nil
}

// Exists implements Storage.Exists
func (s *FileStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(s.rootPath, path)
	_, err := os.Stat(fullPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
