package storage

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileStorage implements the Storage interface using the file system.
type FileStorage struct {
	// basePath is the base directory for file storage
	basePath string
}

// NewFileStorage creates a new file storage with the specified base path.
func NewFileStorage(basePath string) (Storage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, &StorageError{
			Operation: "NewFileStorage",
			Path:      basePath,
			Err:       err,
		}
	}

	// Create a metadata directory for storing content types and ACLs
	metadataDir := filepath.Join(basePath, ".metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return nil, &StorageError{
			Operation: "NewFileStorage",
			Path:      metadataDir,
			Err:       err,
		}
	}

	return &FileStorage{
		basePath: basePath,
	}, nil
}

// StoreResource implements Storage.StoreResource.
func (f *FileStorage) StoreResource(path string, data []byte, contentType string) error {
	// Normalize path
	path = normalizePath(path)

	// Check if parent container exists
	parent := getParentPath(path)
	if parent != "/" {
		exists, isContainer, err := f.checkResourceExists(parent)
		if err != nil {
			return &StorageError{
				Operation: "StoreResource",
				Path:      path,
				Err:       err,
			}
		}
		if !exists {
			return &StorageError{
				Operation: "StoreResource",
				Path:      path,
				Err:       ErrResourceNotFound,
			}
		}
		if !isContainer {
			return &StorageError{
				Operation: "StoreResource",
				Path:      path,
				Err:       ErrNotContainer,
			}
		}
	}

	// Get file path
	filePath := f.getFilePath(path)

	// Check if resource exists and is a container
	if stat, err := os.Stat(filePath); err == nil && stat.IsDir() {
		return &StorageError{
			Operation: "StoreResource",
			Path:      path,
			Err:       ErrIsContainer,
		}
	}

	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return &StorageError{
			Operation: "StoreResource",
			Path:      path,
			Err:       err,
		}
	}

	// Write data to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return &StorageError{
			Operation: "StoreResource",
			Path:      path,
			Err:       err,
		}
	}

	// Store content type
	if err := f.storeContentType(path, contentType); err != nil {
		return &StorageError{
			Operation: "StoreResource",
			Path:      path,
			Err:       err,
		}
	}

	return nil
}

// GetResource implements Storage.GetResource.
func (f *FileStorage) GetResource(path string) ([]byte, string, error) {
	// Normalize path
	path = normalizePath(path)

	// Get file path
	filePath := f.getFilePath(path)

	// Check if resource exists
	stat, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", &StorageError{
				Operation: "GetResource",
				Path:      path,
				Err:       ErrResourceNotFound,
			}
		}
		return nil, "", &StorageError{
			Operation: "GetResource",
			Path:      path,
			Err:       err,
		}
	}

	// Check if resource is a container
	if stat.IsDir() {
		return nil, "", &StorageError{
			Operation: "GetResource",
			Path:      path,
			Err:       ErrIsContainer,
		}
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", &StorageError{
			Operation: "GetResource",
			Path:      path,
			Err:       err,
		}
	}

	// Get content type
	contentType, err := f.getContentType(path)
	if err != nil {
		return nil, "", &StorageError{
			Operation: "GetResource",
			Path:      path,
			Err:       err,
		}
	}

	return data, contentType, nil
}

// ResourceExists implements Storage.ResourceExists.
func (f *FileStorage) ResourceExists(path string) (bool, error) {
	// Normalize path
	path = normalizePath(path)

	// Get file path
	filePath := f.getFilePath(path)

	// Check if file exists
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, &StorageError{
			Operation: "ResourceExists",
			Path:      path,
			Err:       err,
		}
	}

	return true, nil
}

// DeleteResource implements Storage.DeleteResource.
func (f *FileStorage) DeleteResource(path string) error {
	// Normalize path
	path = normalizePath(path)

	// Get file path
	filePath := f.getFilePath(path)

	// Check if resource exists
	stat, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &StorageError{
				Operation: "DeleteResource",
				Path:      path,
				Err:       ErrResourceNotFound,
			}
		}
		return &StorageError{
			Operation: "DeleteResource",
			Path:      path,
			Err:       err,
		}
	}

	// If it's a container, check if it's empty
	if stat.IsDir() {
		files, err := os.ReadDir(filePath)
		if err != nil {
			return &StorageError{
				Operation: "DeleteResource",
				Path:      path,
				Err:       err,
			}
		}
		if len(files) > 0 {
			return &StorageError{
				Operation: "DeleteResource",
				Path:      path,
				Err:       errors.New("container is not empty"),
			}
		}
	}

	// Delete resource
	if err := os.RemoveAll(filePath); err != nil {
		return &StorageError{
			Operation: "DeleteResource",
			Path:      path,
			Err:       err,
		}
	}

	// Delete metadata
	f.deleteMetadata(path)

	return nil
}

// GetResourceMetadata implements Storage.GetResourceMetadata.
func (f *FileStorage) GetResourceMetadata(path string) (*ResourceMetadata, error) {
	// Normalize path
	path = normalizePath(path)

	// Get file path
	filePath := f.getFilePath(path)

	// Check if resource exists
	stat, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &StorageError{
				Operation: "GetResourceMetadata",
				Path:      path,
				Err:       ErrResourceNotFound,
			}
		}
		return nil, &StorageError{
			Operation: "GetResourceMetadata",
			Path:      path,
			Err:       err,
		}
	}

	// Get content type
	contentType, err := f.getContentType(path)
	if err != nil {
		return nil, &StorageError{
			Operation: "GetResourceMetadata",
			Path:      path,
			Err:       err,
		}
	}

	return &ResourceMetadata{
		Path:         path,
		ContentType:  contentType,
		Size:         stat.Size(),
		LastModified: stat.ModTime().Format(time.RFC3339),
		IsContainer:  stat.IsDir(),
	}, nil
}

// CreateContainer implements Storage.CreateContainer.
func (f *FileStorage) CreateContainer(path string) error {
	// Normalize path
	path = normalizePath(path)
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	// Get file path
	filePath := f.getFilePath(path)

	// Check if container already exists
	if _, err := os.Stat(filePath); err == nil {
		return &StorageError{
			Operation: "CreateContainer",
			Path:      path,
			Err:       ErrResourceExists,
		}
	}

	// Check if parent container exists
	parent := getParentPath(path)
	if parent != "/" {
		exists, isContainer, err := f.checkResourceExists(parent)
		if err != nil {
			return &StorageError{
				Operation: "CreateContainer",
				Path:      path,
				Err:       err,
			}
		}
		if !exists {
			return &StorageError{
				Operation: "CreateContainer",
				Path:      path,
				Err:       ErrResourceNotFound,
			}
		}
		if !isContainer {
			return &StorageError{
				Operation: "CreateContainer",
				Path:      path,
				Err:       ErrNotContainer,
			}
		}
	}

	// Create container
	if err := os.MkdirAll(filePath, 0755); err != nil {
		return &StorageError{
			Operation: "CreateContainer",
			Path:      path,
			Err:       err,
		}
	}

	// Store content type
	if err := f.storeContentType(path, "text/turtle"); err != nil {
		return &StorageError{
			Operation: "CreateContainer",
			Path:      path,
			Err:       err,
		}
	}

	return nil
}

// ListContainer implements Storage.ListContainer.
func (f *FileStorage) ListContainer(path string) ([]ResourceMetadata, error) {
	// Normalize path
	path = normalizePath(path)
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	// Get file path
	filePath := f.getFilePath(path)

	// Check if container exists
	stat, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &StorageError{
				Operation: "ListContainer",
				Path:      path,
				Err:       ErrResourceNotFound,
			}
		}
		return nil, &StorageError{
			Operation: "ListContainer",
			Path:      path,
			Err:       err,
		}
	}

	// Check if it's a container
	if !stat.IsDir() {
		return nil, &StorageError{
			Operation: "ListContainer",
			Path:      path,
			Err:       ErrNotContainer,
		}
	}

	// Read directory
	files, err := os.ReadDir(filePath)
	if err != nil {
		return nil, &StorageError{
			Operation: "ListContainer",
			Path:      path,
			Err:       err,
		}
	}

	// Create result list
	var results []ResourceMetadata
	for _, file := range files {
		name := file.Name()
		if strings.HasPrefix(name, ".") {
			// Skip hidden files
			continue
		}

		resourcePath := path + name
		if file.IsDir() {
			resourcePath += "/"
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		contentType, err := f.getContentType(resourcePath)
		if err != nil {
			contentType = "application/octet-stream"
		}

		results = append(results, ResourceMetadata{
			Path:         resourcePath,
			ContentType:  contentType,
			Size:         info.Size(),
			LastModified: info.ModTime().Format(time.RFC3339),
			IsContainer:  file.IsDir(),
		})
	}

	return results, nil
}

// StoreACL implements Storage.StoreACL.
func (f *FileStorage) StoreACL(path string, data []byte) error {
	// Normalize path
	path = normalizePath(path)

	// Check if resource exists
	exists, _, err := f.checkResourceExists(path)
	if err != nil {
		return &StorageError{
			Operation: "StoreACL",
			Path:      path,
			Err:       err,
		}
	}
	if !exists {
		return &StorageError{
			Operation: "StoreACL",
			Path:      path,
			Err:       ErrResourceNotFound,
		}
	}

	// Get ACL file path
	aclPath := f.getAclPath(path)

	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(aclPath), 0755); err != nil {
		return &StorageError{
			Operation: "StoreACL",
			Path:      path,
			Err:       err,
		}
	}

	// Write ACL to file
	if err := os.WriteFile(aclPath, data, 0644); err != nil {
		return &StorageError{
			Operation: "StoreACL",
			Path:      path,
			Err:       err,
		}
	}

	return nil
}

// GetACL implements Storage.GetACL.
func (f *FileStorage) GetACL(path string) ([]byte, error) {
	// Normalize path
	path = normalizePath(path)

	// Check if resource exists
	exists, _, err := f.checkResourceExists(path)
	if err != nil {
		return nil, &StorageError{
			Operation: "GetACL",
			Path:      path,
			Err:       err,
		}
	}
	if !exists {
		return nil, &StorageError{
			Operation: "GetACL",
			Path:      path,
			Err:       ErrResourceNotFound,
		}
	}

	// Get ACL file path
	aclPath := f.getAclPath(path)

	// Check if ACL exists
	if _, err := os.Stat(aclPath); err != nil {
		if os.IsNotExist(err) {
			// Try to find an inherited ACL
			currentPath := path
			for {
				currentPath = getParentPath(currentPath)
				if currentPath == "/" {
					break
				}

				inheritedAclPath := f.getAclPath(currentPath)
				if _, err := os.Stat(inheritedAclPath); err == nil {
					// Found an inherited ACL
					return os.ReadFile(inheritedAclPath)
				}
			}

			// Try root ACL
			rootAclPath := f.getAclPath("/")
			if _, err := os.Stat(rootAclPath); err == nil {
				return os.ReadFile(rootAclPath)
			}

			// Return empty ACL
			return []byte{}, nil
		}
		return nil, &StorageError{
			Operation: "GetACL",
			Path:      path,
			Err:       err,
		}
	}

	// Read ACL from file
	aclData, err := os.ReadFile(aclPath)
	if err != nil {
		return nil, &StorageError{
			Operation: "GetACL",
			Path:      path,
			Err:       err,
		}
	}

	return aclData, nil
}

// Helper methods

// getFilePath converts a resource path to a file system path.
func (f *FileStorage) getFilePath(path string) string {
	// Remove leading slash from path
	path = strings.TrimPrefix(path, "/")
	return filepath.Join(f.basePath, path)
}

// getMetadataPath returns the path for metadata storage.
func (f *FileStorage) getMetadataPath(path string) string {
	// Remove leading slash from path
	path = strings.TrimPrefix(path, "/")
	return filepath.Join(f.basePath, ".metadata", path)
}

// getContentTypePath returns the path for storing content type.
func (f *FileStorage) getContentTypePath(path string) string {
	return f.getMetadataPath(path) + ".type"
}

// getAclPath returns the path for storing ACL.
func (f *FileStorage) getAclPath(path string) string {
	return f.getMetadataPath(path) + ".acl"
}

// storeContentType stores the content type for a resource.
func (f *FileStorage) storeContentType(path string, contentType string) error {
	contentTypePath := f.getContentTypePath(path)

	// Create parent directories if they don't exist
	if err := os.MkdirAll(filepath.Dir(contentTypePath), 0755); err != nil {
		return err
	}

	// Write content type to file
	return os.WriteFile(contentTypePath, []byte(contentType), 0644)
}

// getContentType retrieves the content type for a resource.
func (f *FileStorage) getContentType(path string) (string, error) {
	contentTypePath := f.getContentTypePath(path)

	// Check if content type file exists
	if _, err := os.Stat(contentTypePath); err != nil {
		if os.IsNotExist(err) {
			// Default content type
			if strings.HasSuffix(path, "/") {
				return "text/turtle", nil
			}
			return "application/octet-stream", nil
		}
		return "", err
	}

	// Read content type from file
	contentTypeBytes, err := os.ReadFile(contentTypePath)
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(contentTypeBytes)), nil
}

// deleteMetadata deletes all metadata for a resource.
func (f *FileStorage) deleteMetadata(path string) {
	metadataBasePath := f.getMetadataPath(path)
	// Delete content type file
	os.Remove(metadataBasePath + ".type")
	// Delete ACL file
	os.Remove(metadataBasePath + ".acl")
}

// checkResourceExists checks if a resource exists and if it's a container.
func (f *FileStorage) checkResourceExists(path string) (bool, bool, error) {
	filePath := f.getFilePath(path)
	stat, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, stat.IsDir(), nil
}
