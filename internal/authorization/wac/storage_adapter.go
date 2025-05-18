package wac

import (
	"github.com/yourusername/solid-go/internal/storage"
)

// StorageAdapter adapts storage.Factory to wac.StorageAccessor
type StorageAdapter struct {
	factory storage.Factory
}

// NewStorageAdapter creates a new StorageAdapter
func NewStorageAdapter(factory storage.Factory) *StorageAdapter {
	return &StorageAdapter{
		factory: factory,
	}
}

// GetResource implements wac.StorageAccessor.GetResource
func (a *StorageAdapter) GetResource(path string) ([]byte, string, error) {
	// Create a storage instance with the default strategy
	store, err := a.factory.CreateStorage("file")
	if err != nil {
		return nil, "", err
	}
	return store.GetResource(path)
}

// ResourceExists implements wac.StorageAccessor.ResourceExists
func (a *StorageAdapter) ResourceExists(path string) (bool, error) {
	// Create a storage instance with the default strategy
	store, err := a.factory.CreateStorage("file")
	if err != nil {
		return false, err
	}
	return store.ResourceExists(path)
}
