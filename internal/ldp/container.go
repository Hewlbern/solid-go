package ldp

import (
	"context"
	"path/filepath"
	"strings"

	"solid-go/internal/rdf"
	"solid-go/internal/storage"
)

// Container represents an LDP container
type Container struct {
	storage storage.Storage
	path    string
}

// NewContainer creates a new LDP container
func NewContainer(storage storage.Storage, path string) *Container {
	return &Container{
		storage: storage,
		path:    path,
	}
}

// CreateResource creates a new resource in the container
func (c *Container) CreateResource(ctx context.Context, name string, data []byte) error {
	path := filepath.Join(c.path, name)
	return c.storage.Put(ctx, path, data)
}

// GetResource retrieves a resource from the container
func (c *Container) GetResource(ctx context.Context, name string) ([]byte, error) {
	path := filepath.Join(c.path, name)
	return c.storage.Get(ctx, path)
}

// DeleteResource deletes a resource from the container
func (c *Container) DeleteResource(ctx context.Context, name string) error {
	path := filepath.Join(c.path, name)
	return c.storage.Delete(ctx, path)
}

// ListResources lists all resources in the container
func (c *Container) ListResources(ctx context.Context) ([]string, error) {
	// In a real implementation, this would use the storage backend to list resources
	// For now, we'll return an empty list
	return []string{}, nil
}

// IsContainer checks if a path is a container
func (c *Container) IsContainer(path string) bool {
	return strings.HasSuffix(path, "/")
}

// GetContainerMetadata returns the container's metadata as RDF
func (c *Container) GetContainerMetadata() *rdf.Graph {
	graph := rdf.NewGraph()
	graph.Add(rdf.Triple{
		Subject:   c.path,
		Predicate: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type",
		Object:    "http://www.w3.org/ns/ldp#Container",
	})
	return graph
}
