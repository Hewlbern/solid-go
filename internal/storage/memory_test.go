package storage

import (
	"bytes"
	"testing"
)

func TestMemoryStorage_StoreGetResource(t *testing.T) {
	storage := NewMemoryStorage()

	// Create root container implicitly
	storage.resources["/"] = &Resource{
		Data:        []byte{},
		ContentType: "text/turtle",
		IsContainer: true,
	}

	// Test cases
	tests := []struct {
		name        string
		path        string
		data        []byte
		contentType string
		wantErr     bool
	}{
		{
			name:        "Store in root",
			path:        "/test.txt",
			data:        []byte("test data"),
			contentType: "text/plain",
			wantErr:     false,
		},
		{
			name:        "Store with non-normalized path",
			path:        "test2.txt",
			data:        []byte("test data 2"),
			contentType: "text/plain",
			wantErr:     false,
		},
		{
			name:        "Store in non-existent container",
			path:        "/nonexistent/test.txt",
			data:        []byte("test data"),
			contentType: "text/plain",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.StoreResource(tt.path, tt.data, tt.contentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryStorage.StoreResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Normalize path for checking
			path := normalizePath(tt.path)

			// Verify resource was stored
			data, contentType, err := storage.GetResource(path)
			if err != nil {
				t.Errorf("MemoryStorage.GetResource() error = %v", err)
				return
			}

			if !bytes.Equal(data, tt.data) {
				t.Errorf("MemoryStorage.GetResource() data = %v, want %v", string(data), string(tt.data))
			}

			if contentType != tt.contentType {
				t.Errorf("MemoryStorage.GetResource() contentType = %v, want %v", contentType, tt.contentType)
			}
		})
	}
}

func TestMemoryStorage_CreateContainer(t *testing.T) {
	storage := NewMemoryStorage()

	// Create root container implicitly
	storage.resources["/"] = &Resource{
		Data:        []byte{},
		ContentType: "text/turtle",
		IsContainer: true,
	}

	// Create a test container
	err := storage.CreateContainer("/container")
	if err != nil {
		t.Fatalf("MemoryStorage.CreateContainer() error = %v", err)
	}

	// Verify container was created
	exists, err := storage.ResourceExists("/container/")
	if err != nil {
		t.Errorf("MemoryStorage.ResourceExists() error = %v", err)
		return
	}
	if !exists {
		t.Errorf("MemoryStorage.CreateContainer() container not created")
	}

	// Check container metadata
	metadata, err := storage.GetResourceMetadata("/container/")
	if err != nil {
		t.Errorf("MemoryStorage.GetResourceMetadata() error = %v", err)
		return
	}
	if !metadata.IsContainer {
		t.Errorf("MemoryStorage.GetResourceMetadata() IsContainer = %v, want true", metadata.IsContainer)
	}
	if metadata.ContentType != "text/turtle" {
		t.Errorf("MemoryStorage.GetResourceMetadata() ContentType = %v, want %v", metadata.ContentType, "text/turtle")
	}

	// Create a nested container
	err = storage.CreateContainer("/container/nested")
	if err != nil {
		t.Fatalf("MemoryStorage.CreateContainer() error = %v", err)
	}

	// Verify nested container was created
	exists, err = storage.ResourceExists("/container/nested/")
	if err != nil {
		t.Errorf("MemoryStorage.ResourceExists() error = %v", err)
		return
	}
	if !exists {
		t.Errorf("MemoryStorage.CreateContainer() nested container not created")
	}

	// Try to create a container in a non-existent parent
	err = storage.CreateContainer("/nonexistent/container")
	if err == nil {
		t.Errorf("MemoryStorage.CreateContainer() expected error for non-existent parent")
	}
}

func TestMemoryStorage_ListContainer(t *testing.T) {
	storage := NewMemoryStorage()

	// Create root container implicitly
	storage.resources["/"] = &Resource{
		Data:        []byte{},
		ContentType: "text/turtle",
		IsContainer: true,
	}

	// Create a test container
	err := storage.CreateContainer("/container")
	if err != nil {
		t.Fatalf("MemoryStorage.CreateContainer() error = %v", err)
	}

	// Create a nested container
	err = storage.CreateContainer("/container/nested")
	if err != nil {
		t.Fatalf("MemoryStorage.CreateContainer() error = %v", err)
	}

	// Create resources in the container
	err = storage.StoreResource("/container/test1.txt", []byte("test data 1"), "text/plain")
	if err != nil {
		t.Fatalf("MemoryStorage.StoreResource() error = %v", err)
	}
	err = storage.StoreResource("/container/test2.txt", []byte("test data 2"), "text/plain")
	if err != nil {
		t.Fatalf("MemoryStorage.StoreResource() error = %v", err)
	}

	// List the container
	resources, err := storage.ListContainer("/container")
	if err != nil {
		t.Fatalf("MemoryStorage.ListContainer() error = %v", err)
	}

	// Check the number of resources
	if len(resources) != 3 {
		t.Errorf("MemoryStorage.ListContainer() len = %v, want 3", len(resources))
	}

	// Check that all expected resources are in the list
	paths := make(map[string]bool)
	for _, res := range resources {
		paths[res.Path] = true
	}

	expectedPaths := []string{
		"/container/test1.txt",
		"/container/test2.txt",
		"/container/nested/",
	}

	for _, path := range expectedPaths {
		if !paths[path] {
			t.Errorf("MemoryStorage.ListContainer() missing path %v", path)
		}
	}
}

func TestMemoryStorage_DeleteResource(t *testing.T) {
	storage := NewMemoryStorage()

	// Create root container implicitly
	storage.resources["/"] = &Resource{
		Data:        []byte{},
		ContentType: "text/turtle",
		IsContainer: true,
	}

	// Create a test container
	err := storage.CreateContainer("/container")
	if err != nil {
		t.Fatalf("MemoryStorage.CreateContainer() error = %v", err)
	}

	// Store a resource
	err = storage.StoreResource("/container/test.txt", []byte("test data"), "text/plain")
	if err != nil {
		t.Fatalf("MemoryStorage.StoreResource() error = %v", err)
	}

	// Delete the resource
	err = storage.DeleteResource("/container/test.txt")
	if err != nil {
		t.Errorf("MemoryStorage.DeleteResource() error = %v", err)
	}

	// Verify the resource was deleted
	exists, err := storage.ResourceExists("/container/test.txt")
	if err != nil {
		t.Errorf("MemoryStorage.ResourceExists() error = %v", err)
		return
	}
	if exists {
		t.Error("MemoryStorage.DeleteResource() resource still exists")
	}
}

func TestMemoryStorage_ACL(t *testing.T) {
	storage := NewMemoryStorage()

	// Create root container implicitly
	storage.resources["/"] = &Resource{
		Data:        []byte{},
		ContentType: "text/turtle",
		IsContainer: true,
	}

	// Create a test container
	err := storage.CreateContainer("/container")
	if err != nil {
		t.Fatalf("MemoryStorage.CreateContainer() error = %v", err)
	}

	// Store an ACL
	aclData := []byte("@prefix acl: <http://www.w3.org/ns/auth/acl#>.\n" +
		"<#owner> a acl:Authorization;\n" +
		"  acl:agent <http://example.org/user1>;\n" +
		"  acl:mode acl:Read, acl:Write.")

	err = storage.StoreACL("/container", aclData)
	if err != nil {
		t.Errorf("MemoryStorage.StoreACL() error = %v", err)
	}

	// Get the ACL
	storedACL, err := storage.GetACL("/container")
	if err != nil {
		t.Errorf("MemoryStorage.GetACL() error = %v", err)
		return
	}

	// Verify the ACL
	if !bytes.Equal(storedACL, aclData) {
		t.Errorf("MemoryStorage.GetACL() = %v, want %v", string(storedACL), string(aclData))
	}
}
