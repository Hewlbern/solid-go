package storage

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
)

// FileStorage implements the Storage interface using the filesystem
type FileStorage struct {
	basePath string
}

// NewFileStorage creates a new file-based storage
func NewFileStorage(basePath string) (*FileStorage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}

	return &FileStorage{
		basePath: basePath,
	}, nil
}

// Get retrieves data from a file
func (s *FileStorage) Get(ctx context.Context, path string) ([]byte, error) {
	fullPath := filepath.Join(s.basePath, path)
	return ioutil.ReadFile(fullPath)
}

// Put stores data in a file
func (s *FileStorage) Put(ctx context.Context, path string, data []byte) error {
	fullPath := filepath.Join(s.basePath, path)

	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(fullPath, data, 0644)
}

// Delete removes a file
func (s *FileStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.basePath, path)
	return os.Remove(fullPath)
}

// List returns a list of resources in a directory
func (s *FileStorage) List(ctx context.Context, path string) ([]string, error) {
	fullPath := filepath.Join(s.basePath, path)

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
