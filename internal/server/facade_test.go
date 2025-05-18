package server

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestNewServerFacade(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "solid-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test cases
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "Valid configuration",
			config: &Config{
				Port:        8080,
				StoragePath: tempDir,
				LogLevel:    "info",
			},
			wantErr: false,
		},
		{
			name: "Valid configuration with explicit strategies",
			config: &Config{
				Port:            8080,
				StoragePath:     tempDir,
				LogLevel:        "info",
				StorageStrategy: "file",
				AuthStrategy:    "webid",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewServerFacade(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewServerFacade() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestServerFacade_Health(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "solid-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a server facade
	facade, err := NewServerFacade(&Config{
		Port:        8080,
		StoragePath: tempDir,
		LogLevel:    "info",
	})
	if err != nil {
		t.Fatalf("Failed to create server facade: %v", err)
	}

	// Set up routes
	facade.setupRoutes()

	// Create a test request for the health endpoint
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	facade.mux.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Health endpoint returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expected := "OK\n"
	if rr.Body.String() != expected {
		t.Errorf("Health endpoint returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

func TestServerFacade_StopStart(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "solid-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a server facade with a random port to avoid conflicts
	facade, err := NewServerFacade(&Config{
		Port:        0, // Use a random port
		StoragePath: tempDir,
		LogLevel:    "info",
	})
	if err != nil {
		t.Fatalf("Failed to create server facade: %v", err)
	}

	// Start the server in a goroutine with a context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Update the server to use the context
	facade.server.BaseContext = func(_ net.Listener) context.Context {
		return ctx
	}

	// Start the server in a goroutine
	go func() {
		if err := facade.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Server error: %v", err)
		}
	}()

	// Allow some time for the server to start
	time.Sleep(100 * time.Millisecond)

	// Stop the server
	if err := facade.Stop(); err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}
}

func TestServerFacade_GetFactories(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "solid-test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a server facade
	facade, err := NewServerFacade(&Config{
		Port:        8080,
		StoragePath: tempDir,
		LogLevel:    "info",
	})
	if err != nil {
		t.Fatalf("Failed to create server facade: %v", err)
	}

	// Test getter methods
	if facade.GetStorageFactory() == nil {
		t.Error("GetStorageFactory() returned nil")
	}

	if facade.GetAuthFactory() == nil {
		t.Error("GetAuthFactory() returned nil")
	}

	if facade.GetEventDispatcher() == nil {
		t.Error("GetEventDispatcher() returned nil")
	}
}
