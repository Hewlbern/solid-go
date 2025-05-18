package wac

import (
	"net/http"

	"github.com/yourusername/solid-go/internal/auth"
)

// MockStorage implements StorageAccessor for testing.
type MockStorage struct {
	resources map[string][]byte
	acls      map[string][]byte
}

// NewMockStorage creates a new mock storage.
func NewMockStorage() *MockStorage {
	return &MockStorage{
		resources: make(map[string][]byte),
		acls:      make(map[string][]byte),
	}
}

// GetResource implements StorageAccessor.GetResource.
func (m *MockStorage) GetResource(path string) ([]byte, string, error) {
	if data, ok := m.resources[path]; ok {
		return data, "text/turtle", nil
	}
	return nil, "", nil
}

// ResourceExists implements StorageAccessor.ResourceExists.
func (m *MockStorage) ResourceExists(path string) (bool, error) {
	_, ok := m.resources[path]
	return ok, nil
}

// AddResource adds a resource to the mock storage.
func (m *MockStorage) AddResource(path string, data []byte) {
	m.resources[path] = data
}

// AddACL adds an ACL to the mock storage.
func (m *MockStorage) AddACL(path string, data []byte) {
	m.acls[path] = data
}

// MockAuthHandler implements auth.AuthHandler for testing.
type MockAuthHandler struct {
	sessions map[string]*auth.Session
}

// NewMockAuthHandler creates a new mock auth handler.
func NewMockAuthHandler() *MockAuthHandler {
	return &MockAuthHandler{
		sessions: make(map[string]*auth.Session),
	}
}

// Authenticate implements auth.AuthHandler.Authenticate.
func (m *MockAuthHandler) Authenticate(req *http.Request, strategy string, credentials map[string]interface{}) (*auth.Session, error) {
	return nil, nil
}

// GetSession implements auth.AuthHandler.GetSession.
func (m *MockAuthHandler) GetSession(req *http.Request) *auth.Session {
	return nil
}

// CreateSession implements auth.AuthHandler.CreateSession.
func (m *MockAuthHandler) CreateSession(w http.ResponseWriter, userID string) (*auth.Session, error) {
	return nil, nil
}

// DestroySession implements auth.AuthHandler.DestroySession.
func (m *MockAuthHandler) DestroySession(w http.ResponseWriter) error {
	return nil
}

// GetStrategy implements auth.AuthHandler.GetStrategy.
func (m *MockAuthHandler) GetStrategy(name string) (auth.Strategy, error) {
	return nil, nil
}
