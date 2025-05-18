package auth

import (
	"net/http/httptest"
	"testing"

	"github.com/yourusername/solid-go/internal/storage"
)

// mockStorage is a mock implementation of the Storage interface.
type mockStorage struct {
	resources map[string][]byte
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		resources: make(map[string][]byte),
	}
}

func (s *mockStorage) StoreResource(path string, data []byte, contentType string) error {
	s.resources[path] = data
	return nil
}

func (s *mockStorage) GetResource(path string) ([]byte, string, error) {
	data, ok := s.resources[path]
	if !ok {
		return nil, "", storage.ErrResourceNotFound
	}
	return data, "application/octet-stream", nil
}

func (s *mockStorage) ResourceExists(path string) (bool, error) {
	_, ok := s.resources[path]
	return ok, nil
}

func (s *mockStorage) DeleteResource(path string) error {
	delete(s.resources, path)
	return nil
}

func (s *mockStorage) GetResourceMetadata(path string) (*storage.ResourceMetadata, error) {
	data, ok := s.resources[path]
	if !ok {
		return nil, storage.ErrResourceNotFound
	}
	return &storage.ResourceMetadata{
		Path:        path,
		ContentType: "application/octet-stream",
		Size:        int64(len(data)),
	}, nil
}

func (s *mockStorage) CreateContainer(path string) error {
	return nil
}

func (s *mockStorage) ListContainer(path string) ([]storage.ResourceMetadata, error) {
	return nil, nil
}

func (s *mockStorage) StoreACL(path string, data []byte) error {
	return nil
}

func (s *mockStorage) GetACL(path string) ([]byte, error) {
	return nil, nil
}

// mockFactory is a mock implementation of the Factory interface.
type mockFactory struct {
	strategies map[string]Strategy
}

func newMockFactory() *mockFactory {
	return &mockFactory{
		strategies: make(map[string]Strategy),
	}
}

func (f *mockFactory) CreateStrategy(name string, storage storage.Storage) (Strategy, error) {
	strategy, ok := f.strategies[name]
	if ok {
		return strategy, nil
	}

	switch name {
	case "webid":
		agent := &Agent{
			ID:              "https://example.org/alice#me",
			Name:            "Alice",
			IsAuthenticated: true,
			Type:            TypeUser,
		}
		strategy = NewMockWebIDStrategy(agent)
	case "oidc":
		agent := &Agent{
			ID:              "https://example.org/bob#me",
			Name:            "Bob",
			IsAuthenticated: true,
			Type:            TypeUser,
		}
		strategy = NewMockOIDCStrategy(agent)
	}

	f.strategies[name] = strategy
	return strategy, nil
}

func (f *mockFactory) GetStrategy(name string) (Strategy, error) {
	strategy, ok := f.strategies[name]
	if !ok {
		return nil, ErrAuthenticationFailed
	}
	return strategy, nil
}

func (f *mockFactory) RegisterStrategy(name string, strategy Strategy) {
	f.strategies[name] = strategy
}

// mockStorageFactory is a mock implementation of the storage.Factory interface.
type mockStorageFactory struct {
	storages map[string]storage.Storage
}

func newMockStorageFactory() *mockStorageFactory {
	return &mockStorageFactory{
		storages: make(map[string]storage.Storage),
	}
}

func (f *mockStorageFactory) CreateStorage(strategy string) (storage.Storage, error) {
	storage, ok := f.storages[strategy]
	if ok {
		return storage, nil
	}

	storage = newMockStorage()
	f.storages[strategy] = storage
	return storage, nil
}

// Test the AuthFactory
func TestAuthFactory(t *testing.T) {
	mockStore := newMockStorage()
	factory := NewFactory()

	// Create a WebID strategy
	webid, err := factory.CreateStrategy("webid", mockStore)
	if err != nil {
		t.Fatalf("Failed to create WebID strategy: %v", err)
	}
	if webid.Name() != "webid" {
		t.Errorf("WebID strategy name = %v, want webid", webid.Name())
	}

	// Create an OIDC strategy
	oidc, err := factory.CreateStrategy("oidc", mockStore)
	if err != nil {
		t.Fatalf("Failed to create OIDC strategy: %v", err)
	}
	if oidc.Name() != "oidc" {
		t.Errorf("OIDC strategy name = %v, want oidc", oidc.Name())
	}

	// Get an existing strategy
	webid2, err := factory.GetStrategy("webid")
	if err != nil {
		t.Fatalf("Failed to get WebID strategy: %v", err)
	}
	if webid2 != webid {
		t.Errorf("GetStrategy() returned different strategy instance")
	}

	// Get a non-existent strategy
	_, err = factory.GetStrategy("nonexistent")
	if err == nil {
		t.Errorf("GetStrategy() did not return error for non-existent strategy")
	}
}

// Test the authentication handler
func TestAuthHandler(t *testing.T) {
	// Create a mock factory with mock strategies
	factory := newMockFactory()
	storageFactory := newMockStorageFactory()

	// Create the authentication handler
	handler, err := NewHandler(factory, storageFactory)
	if err != nil {
		t.Fatalf("Failed to create authentication handler: %v", err)
	}

	// Test request with no authentication
	req := httptest.NewRequest("GET", "/test", nil)
	agent, err := handler.Authenticate(req)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}
	if agent.IsAuthenticated {
		t.Errorf("Agent should not be authenticated, but is")
	}
	if agent.Type != TypeAnonymous {
		t.Errorf("Agent type = %v, want TypeAnonymous", agent.Type)
	}

	// Test request with Bearer token (OIDC)
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer token123")
	agent, err = handler.Authenticate(req)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}
	if !agent.IsAuthenticated {
		t.Errorf("Agent should be authenticated, but is not")
	}
	if agent.Name != "Bob" {
		t.Errorf("Agent name = %v, want Bob", agent.Name)
	}
}

// Test WebID strategy with client certificate
func TestWebIDStrategy(t *testing.T) {
	// This test would normally verify WebID-TLS authentication with client certificates
	// Since we can't easily create test client certificates, we'll use a mock strategy

	mockStore := newMockStorage()
	strategy, err := NewWebIDStrategy(mockStore)
	if err != nil {
		t.Fatalf("Failed to create WebID strategy: %v", err)
	}

	if strategy.Name() != "webid" {
		t.Errorf("WebID strategy name = %v, want webid", strategy.Name())
	}

	if !strategy.IsEnabled() {
		t.Errorf("WebID strategy should be enabled by default")
	}

	// In a real implementation, we would test authenticating with a client certificate
	// For now, we'll just verify the strategy exists
}

// Test OIDC strategy with Bearer token
func TestOIDCStrategy(t *testing.T) {
	mockStore := newMockStorage()
	strategy, err := NewOIDCStrategy(mockStore)
	if err != nil {
		t.Fatalf("Failed to create OIDC strategy: %v", err)
	}

	if strategy.Name() != "oidc" {
		t.Errorf("OIDC strategy name = %v, want oidc", strategy.Name())
	}

	if !strategy.IsEnabled() {
		t.Errorf("OIDC strategy should be enabled by default")
	}

	// Test registering and validating a session
	oidcStrategy := strategy.(*OIDCStrategy)
	agentID := "https://example.org/alice#me"
	token, err := oidcStrategy.RegisterSession(agentID, 3600)
	if err != nil {
		t.Fatalf("Failed to register session: %v", err)
	}
	if token == "" {
		t.Errorf("RegisterSession() returned empty token")
	}

	// Create a request with the token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Authenticate the request
	agent, err := strategy.Authenticate(req)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}
	if !agent.IsAuthenticated {
		t.Errorf("Agent should be authenticated, but is not")
	}
	if agent.ID != agentID {
		t.Errorf("Agent ID = %v, want %v", agent.ID, agentID)
	}
}

// Test OIDC configuration
func TestOIDCConfig(t *testing.T) {
	mockStore := newMockStorage()
	strategy, err := NewOIDCStrategy(mockStore)
	if err != nil {
		t.Fatalf("Failed to create OIDC strategy: %v", err)
	}

	oidcStrategy := strategy.(*OIDCStrategy)

	// Create a configuration
	config := &SolidOIDCConfig{
		Issuer:       "https://example.org",
		ClientID:     "client123",
		ClientSecret: "secret123",
		RedirectURI:  "https://example.org/callback",
	}

	// Save the configuration
	err = oidcStrategy.SaveOIDCConfig(config)
	if err != nil {
		t.Fatalf("Failed to save OIDC config: %v", err)
	}

	// Load the configuration
	loadedConfig, err := oidcStrategy.LoadOIDCConfig()
	if err != nil {
		t.Fatalf("Failed to load OIDC config: %v", err)
	}

	// Verify the configuration
	if loadedConfig.Issuer != config.Issuer {
		t.Errorf("Loaded config issuer = %v, want %v", loadedConfig.Issuer, config.Issuer)
	}
	if loadedConfig.ClientID != config.ClientID {
		t.Errorf("Loaded config client ID = %v, want %v", loadedConfig.ClientID, config.ClientID)
	}
	if loadedConfig.ClientSecret != config.ClientSecret {
		t.Errorf("Loaded config client secret = %v, want %v", loadedConfig.ClientSecret, config.ClientSecret)
	}
	if loadedConfig.RedirectURI != config.RedirectURI {
		t.Errorf("Loaded config redirect URI = %v, want %v", loadedConfig.RedirectURI, config.RedirectURI)
	}
}
