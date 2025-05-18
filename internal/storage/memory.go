package storage

import (
	"bytes"
	"errors"
	"path"
	"strings"
	"sync"
	"time"
)

// Resource represents a resource stored in memory.
type Resource struct {
	// Data is the content of the resource
	Data []byte
	// ContentType is the MIME type of the resource
	ContentType string
	// LastModified is the time the resource was last modified
	LastModified time.Time
	// IsContainer indicates whether the resource is a container
	IsContainer bool
}

// MemoryStorage implements the Storage interface using in-memory storage.
type MemoryStorage struct {
	// resources is a map of path to resource
	resources map[string]*Resource
	// acls is a map of path to ACL data
	acls map[string][]byte
	// mutex protects the resources and acls maps
	mutex sync.RWMutex
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		resources: make(map[string]*Resource),
		acls:      make(map[string][]byte),
	}
}

// StoreResource implements Storage.StoreResource.
func (m *MemoryStorage) StoreResource(path string, data []byte, contentType string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Normalize path
	path = normalizePath(path)

	// Check if parent container exists
	parent := getParentPath(path)
	if parent != "/" {
		parentResource, exists := m.resources[parent]
		if !exists {
			return &StorageError{
				Operation: "StoreResource",
				Path:      path,
				Err:       ErrResourceNotFound,
			}
		}
		if !parentResource.IsContainer {
			return &StorageError{
				Operation: "StoreResource",
				Path:      path,
				Err:       ErrNotContainer,
			}
		}
	}

	// If resource exists, update it
	if resource, exists := m.resources[path]; exists {
		if resource.IsContainer {
			return &StorageError{
				Operation: "StoreResource",
				Path:      path,
				Err:       ErrIsContainer,
			}
		}
		resource.Data = data
		resource.ContentType = contentType
		resource.LastModified = time.Now()
	} else {
		// Create a new resource
		m.resources[path] = &Resource{
			Data:         data,
			ContentType:  contentType,
			LastModified: time.Now(),
			IsContainer:  false,
		}
	}

	return nil
}

// GetResource implements Storage.GetResource.
func (m *MemoryStorage) GetResource(path string) ([]byte, string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Normalize path
	path = normalizePath(path)

	// Get resource
	resource, exists := m.resources[path]
	if !exists {
		return nil, "", &StorageError{
			Operation: "GetResource",
			Path:      path,
			Err:       ErrResourceNotFound,
		}
	}

	if resource.IsContainer {
		return nil, "", &StorageError{
			Operation: "GetResource",
			Path:      path,
			Err:       ErrIsContainer,
		}
	}

	return resource.Data, resource.ContentType, nil
}

// ResourceExists implements Storage.ResourceExists.
func (m *MemoryStorage) ResourceExists(path string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Normalize path
	path = normalizePath(path)

	_, exists := m.resources[path]
	return exists, nil
}

// DeleteResource implements Storage.DeleteResource.
func (m *MemoryStorage) DeleteResource(path string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Normalize path
	path = normalizePath(path)

	// Check if resource exists
	resource, exists := m.resources[path]
	if !exists {
		return &StorageError{
			Operation: "DeleteResource",
			Path:      path,
			Err:       ErrResourceNotFound,
		}
	}

	// If it's a container, check if it's empty
	if resource.IsContainer {
		for storedPath := range m.resources {
			if storedPath != path && strings.HasPrefix(storedPath, path) {
				return &StorageError{
					Operation: "DeleteResource",
					Path:      path,
					Err:       errors.New("container is not empty"),
				}
			}
		}
	}

	// Delete the resource
	delete(m.resources, path)

	// Delete ACL if it exists
	delete(m.acls, path)

	return nil
}

// GetResourceMetadata implements Storage.GetResourceMetadata.
func (m *MemoryStorage) GetResourceMetadata(path string) (*ResourceMetadata, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Normalize path
	path = normalizePath(path)

	// Get resource
	resource, exists := m.resources[path]
	if !exists {
		return nil, &StorageError{
			Operation: "GetResourceMetadata",
			Path:      path,
			Err:       ErrResourceNotFound,
		}
	}

	return &ResourceMetadata{
		Path:         path,
		ContentType:  resource.ContentType,
		Size:         int64(len(resource.Data)),
		LastModified: resource.LastModified.Format(time.RFC3339),
		IsContainer:  resource.IsContainer,
	}, nil
}

// CreateContainer implements Storage.CreateContainer.
func (m *MemoryStorage) CreateContainer(path string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Normalize path
	path = normalizePath(path)

	// Check if parent container exists
	parent := getParentPath(path)
	if parent != "/" {
		parentResource, exists := m.resources[parent]
		if !exists {
			// Create parent container recursively
			if err := m.CreateContainer(parent); err != nil {
				return err
			}
		} else if !parentResource.IsContainer {
			return &StorageError{
				Operation: "CreateContainer",
				Path:      path,
				Err:       ErrNotContainer,
			}
		}
	}

	// If resource exists, check if it's a container
	if resource, exists := m.resources[path]; exists {
		if !resource.IsContainer {
			return &StorageError{
				Operation: "CreateContainer",
				Path:      path,
				Err:       ErrIsContainer,
			}
		}
		return nil
	}

	// Create a new container
	m.resources[path] = &Resource{
		Data:         []byte{},
		ContentType:  "text/turtle",
		LastModified: time.Now(),
		IsContainer:  true,
	}

	return nil
}

// ListContainer implements Storage.ListContainer.
func (m *MemoryStorage) ListContainer(path string) ([]ResourceMetadata, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Normalize path
	path = normalizePath(path)
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	// Check if container exists
	resource, exists := m.resources[path]
	if !exists {
		return nil, &StorageError{
			Operation: "ListContainer",
			Path:      path,
			Err:       ErrResourceNotFound,
		}
	}

	if !resource.IsContainer {
		return nil, &StorageError{
			Operation: "ListContainer",
			Path:      path,
			Err:       ErrNotContainer,
		}
	}

	// List resources in container
	var results []ResourceMetadata
	for storedPath, resource := range m.resources {
		// Check if the resource is directly in this container
		if storedPath != path && strings.HasPrefix(storedPath, path) {
			// Skip resources that are in subdirectories
			relPath := strings.TrimPrefix(storedPath, path)
			if strings.Contains(relPath, "/") && !resource.IsContainer {
				continue
			}

			results = append(results, ResourceMetadata{
				Path:         storedPath,
				ContentType:  resource.ContentType,
				Size:         int64(len(resource.Data)),
				LastModified: resource.LastModified.Format(time.RFC3339),
				IsContainer:  resource.IsContainer,
			})
		}
	}

	return results, nil
}

// StoreACL implements Storage.StoreACL.
func (m *MemoryStorage) StoreACL(path string, data []byte) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Normalize path
	path = normalizePath(path)

	// Check if resource exists
	if _, exists := m.resources[path]; !exists {
		return &StorageError{
			Operation: "StoreACL",
			Path:      path,
			Err:       ErrResourceNotFound,
		}
	}

	// Store ACL
	m.acls[path] = bytes.Clone(data)

	return nil
}

// GetACL implements Storage.GetACL.
func (m *MemoryStorage) GetACL(path string) ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Normalize path
	path = normalizePath(path)

	// Check if resource exists
	if _, exists := m.resources[path]; !exists {
		return nil, &StorageError{
			Operation: "GetACL",
			Path:      path,
			Err:       ErrResourceNotFound,
		}
	}

	// Get ACL
	acl, exists := m.acls[path]
	if !exists {
		// Try to find an inherited ACL
		for acl, acl = nil, nil; path != "/"; path = getParentPath(path) {
			if aclData, exists := m.acls[path]; exists {
				acl = aclData
				break
			}
		}

		// If no ACL was found, use root ACL
		if acl == nil {
			acl, _ = m.acls["/"]
		}

		// If still no ACL, return empty ACL
		if acl == nil {
			return []byte{}, nil
		}
	}

	return bytes.Clone(acl), nil
}

// Helper functions

// normalizePath normalizes a path.
func normalizePath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	return path.Clean(p)
}

// getParentPath returns the parent path.
func getParentPath(p string) string {
	p = normalizePath(p)
	if p == "/" {
		return "/"
	}
	return path.Dir(p)
}
