// Package authorization provides implementations for reading auxiliary resource permissions.
package authorization

import (
	"context"
	"path/filepath"

	"solid-go/internal/authorization/permissions"
)

// AuxiliaryReader reads permissions from auxiliary resources.
// It handles reading permissions from resources that are auxiliary to the main resource.
type AuxiliaryReader struct {
	reader PermissionReader
}

// NewAuxiliaryReader creates a new AuxiliaryReader with the given reader.
func NewAuxiliaryReader(reader PermissionReader) *AuxiliaryReader {
	return &AuxiliaryReader{
		reader: reader,
	}
}

// Read implements PermissionReader.
// It reads permissions from auxiliary resources associated with the given resource.
func (r *AuxiliaryReader) Read(ctx context.Context, resource string) (permissions.PermissionSet, error) {
	// Get auxiliary resource paths
	auxPaths := r.getAuxiliaryPaths(resource)

	// Create a new permission set
	perms := permissions.NewACLPermissionSet()

	// Read permissions from each auxiliary resource
	for _, path := range auxPaths {
		auxPerms, err := r.reader.Read(ctx, path)
		if err != nil {
			continue // Skip this auxiliary resource if it returns an error
		}

		// Add all permissions from this auxiliary resource
		for _, mode := range []permissions.Permission{
			permissions.Read,
			permissions.Write,
			permissions.Append,
			permissions.Control,
		} {
			if auxPerms.Has(mode) {
				perms.Add(mode)
			}
		}
	}

	return perms, nil
}

// getAuxiliaryPaths returns the paths of auxiliary resources for the given resource.
func (r *AuxiliaryReader) getAuxiliaryPaths(resource string) []string {
	dir := filepath.Dir(resource)
	base := filepath.Base(resource)
	return []string{
		filepath.Join(dir, "."+base+".meta"),
		filepath.Join(dir, "."+base+".acl"),
	}
}

// GetAuxiliaryPath gets the auxiliary resource path
func (r *AuxiliaryReader) GetAuxiliaryPath(path string) string {
	return filepath.Join(filepath.Dir(path), ".aux")
}

// IsAuxiliary checks if a path is an auxiliary resource
func (r *AuxiliaryReader) IsAuxiliary(path string) bool {
	return filepath.Base(path) == ".aux"
}
