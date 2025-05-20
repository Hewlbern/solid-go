package util

import (
	"path"
	"path/filepath"
	"strings"
)

// PathUtil provides utility functions for path manipulation
type PathUtil struct{}

// NewPathUtil creates a new PathUtil
func NewPathUtil() *PathUtil {
	return &PathUtil{}
}

// Join joins path elements
func (p *PathUtil) Join(elem ...string) string {
	return path.Join(elem...)
}

// Clean cleans a path
func (p *PathUtil) Clean(path string) string {
	return filepath.Clean(path)
}

// Base returns the last element of a path
func (p *PathUtil) Base(path string) string {
	return filepath.Base(path)
}

// Dir returns all but the last element of a path
func (p *PathUtil) Dir(path string) string {
	return filepath.Dir(path)
}

// Ext returns the file extension
func (p *PathUtil) Ext(path string) string {
	return filepath.Ext(path)
}

// IsAbs checks if a path is absolute
func (p *PathUtil) IsAbs(path string) bool {
	return filepath.IsAbs(path)
}

// Rel returns a relative path
func (p *PathUtil) Rel(basepath, targpath string) (string, error) {
	return filepath.Rel(basepath, targpath)
}

// Split splits a path into directory and file components
func (p *PathUtil) Split(path string) (dir, file string) {
	return filepath.Split(path)
}

// ToSlash converts path separators to forward slashes
func (p *PathUtil) ToSlash(path string) string {
	return filepath.ToSlash(path)
}

// FromSlash converts forward slashes to path separators
func (p *PathUtil) FromSlash(path string) string {
	return filepath.FromSlash(path)
}

// HasPrefix checks if a path has a prefix
func (p *PathUtil) HasPrefix(path, prefix string) bool {
	return strings.HasPrefix(path, prefix)
}

// HasSuffix checks if a path has a suffix
func (p *PathUtil) HasSuffix(path, suffix string) bool {
	return strings.HasSuffix(path, suffix)
}

// IsRoot checks if a path is a root path
func (p *PathUtil) IsRoot(path string) bool {
	return path == "/" || path == "\\"
}

// IsSubPath checks if a path is a subpath of another path
func (p *PathUtil) IsSubPath(parent, child string) bool {
	rel, err := p.Rel(parent, child)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..") && rel != "."
}

// Normalize normalizes a path
func (p *PathUtil) Normalize(path string) string {
	return filepath.Clean(filepath.ToSlash(path))
}

// GetParentPath returns the parent path
func (p *PathUtil) GetParentPath(path string) string {
	return filepath.Dir(path)
}

// GetFileName returns the file name without extension
func (p *PathUtil) GetFileName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}

// GetFileExtension returns the file extension
func (p *PathUtil) GetFileExtension(path string) string {
	return filepath.Ext(path)
}

// EnsureTrailingSlash ensures a path ends with a slash
func (p *PathUtil) EnsureTrailingSlash(path string) string {
	if !strings.HasSuffix(path, "/") {
		return path + "/"
	}
	return path
}

// RemoveTrailingSlash removes trailing slash from a path
func (p *PathUtil) RemoveTrailingSlash(path string) string {
	return strings.TrimSuffix(path, "/")
}
