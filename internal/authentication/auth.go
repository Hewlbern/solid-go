package auth

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/yourusername/solid-go/internal/storage"
)

// AgentType represents the type of an agent.
type AgentType string

// Agent types
const (
	TypeUser        AgentType = "User"
	TypeApplication AgentType = "Application"
	TypeGroup       AgentType = "Group"
	TypeAnonymous   AgentType = "Anonymous"
)

// Agent represents an authenticated agent.
type Agent struct {
	// ID is the WebID URI of the agent
	ID string
	// Name is the display name of the agent
	Name string
	// Email is the email address of the agent
	Email string
	// IsAuthenticated indicates whether the agent is authenticated
	IsAuthenticated bool
	// Type is the type of the agent
	Type AgentType
	// Metadata is additional metadata about the agent
	Metadata map[string]interface{}
}

// Session represents an authentication session.
type Session struct {
	// ID is the unique identifier for the session
	ID string
	// UserID is the ID of the authenticated user
	UserID string
	// Expiry is the session expiry time in seconds
	Expiry int
}

// NewAnonymousAgent creates a new anonymous agent.
func NewAnonymousAgent() *Agent {
	return &Agent{
		ID:              "",
		Name:            "Anonymous",
		Email:           "",
		IsAuthenticated: false,
		Type:            TypeAnonymous,
		Metadata:        make(map[string]interface{}),
	}
}

// Strategy is the interface for authentication strategies.
// It implements the Strategy pattern.
type Strategy interface {
	// Name returns the name of the strategy
	Name() string
	// Authenticate authenticates a request and returns an agent
	Authenticate(r *http.Request) (*Agent, error)
	// IsEnabled returns whether the strategy is enabled
	IsEnabled() bool
}

// Factory is the interface for creating authentication strategies.
// It implements the Factory Method pattern.
type Factory interface {
	// CreateStrategy creates a new authentication strategy
	CreateStrategy(name string, storage storage.Storage) (Strategy, error)
	// GetStrategy returns an existing authentication strategy
	GetStrategy(name string) (Strategy, error)
}

// AuthFactory implements the Factory interface.
type AuthFactory struct {
	// strategies is a map of strategy name to strategy instance
	strategies map[string]Strategy
}

// NewFactory creates a new AuthFactory.
func NewFactory() Factory {
	return &AuthFactory{
		strategies: make(map[string]Strategy),
	}
}

// CreateStrategy creates a new authentication strategy.
func (f *AuthFactory) CreateStrategy(name string, storage storage.Storage) (Strategy, error) {
	// Check if the strategy already exists
	if strategy, exists := f.strategies[name]; exists {
		return strategy, nil
	}

	// Create the strategy
	var strategy Strategy
	var err error

	switch name {
	case "webid":
		strategy, err = NewWebIDStrategy(storage)
	case "oidc":
		strategy, err = NewOIDCStrategy(storage)
	default:
		return nil, fmt.Errorf("unknown authentication strategy: %s", name)
	}

	if err != nil {
		return nil, err
	}

	// Store the strategy
	f.strategies[name] = strategy

	return strategy, nil
}

// GetStrategy returns an existing authentication strategy.
func (f *AuthFactory) GetStrategy(name string) (Strategy, error) {
	strategy, exists := f.strategies[name]
	if !exists {
		return nil, fmt.Errorf("authentication strategy not found: %s", name)
	}
	return strategy, nil
}

// AuthHandler is the interface for authentication handlers.
type AuthHandler interface {
	// Authenticate authenticates a request and returns a session
	Authenticate(req *http.Request, strategy string, credentials map[string]interface{}) (*Session, error)
	// GetSession gets a session from a request
	GetSession(req *http.Request) *Session
	// CreateSession creates a new session
	CreateSession(w http.ResponseWriter, userID string) (*Session, error)
	// DestroySession destroys a session
	DestroySession(w http.ResponseWriter) error
}

// Handler handles authentication requests.
type Handler struct {
	// factory is the factory for creating authentication strategies
	factory Factory
	// storageFactory is the factory for creating storage instances
	storageFactory storage.Factory
	// defaultStrategy is the default authentication strategy
	defaultStrategy string
}

// NewHandler creates a new authentication handler.
func NewHandler(factory Factory, storageFactory storage.Factory) (*Handler, error) {
	// Create a file storage for authentication
	fileStorage, err := storageFactory.CreateStorage("file")
	if err != nil {
		return nil, fmt.Errorf("failed to create storage for authentication: %w", err)
	}

	// Create the WebID strategy
	_, err = factory.CreateStrategy("webid", fileStorage)
	if err != nil {
		return nil, fmt.Errorf("failed to create WebID strategy: %w", err)
	}

	// Create the OIDC strategy
	_, err = factory.CreateStrategy("oidc", fileStorage)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC strategy: %w", err)
	}

	return &Handler{
		factory:         factory,
		storageFactory:  storageFactory,
		defaultStrategy: "webid",
	}, nil
}

// Authenticate authenticates a request and returns an agent.
func (h *Handler) Authenticate(r *http.Request) (*Agent, error) {
	// Try each authentication strategy in order
	strategies := []string{"oidc", "webid"}

	var lastErr error
	var agent *Agent

	for _, strategyName := range strategies {
		strategy, err := h.factory.GetStrategy(strategyName)
		if err != nil {
			continue
		}

		if !strategy.IsEnabled() {
			continue
		}

		agent, err = strategy.Authenticate(r)
		if err == nil && agent != nil && agent.IsAuthenticated {
			return agent, nil
		}

		lastErr = err
	}

	if lastErr != nil {
		return nil, lastErr
	}

	// If no strategy succeeded, return an anonymous agent
	return NewAnonymousAgent(), nil
}

// AuthError represents an authentication error.
type AuthError struct {
	// Strategy is the authentication strategy that failed
	Strategy string
	// Message is the error message
	Message string
	// Err is the underlying error
	Err error
}

// Error returns a string representation of the error.
func (e *AuthError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("authentication error in %s: %s: %v", e.Strategy, e.Message, e.Err)
	}
	return fmt.Sprintf("authentication error in %s: %s", e.Strategy, e.Message)
}

// Unwrap returns the underlying error.
func (e *AuthError) Unwrap() error {
	return e.Err
}

// ErrAuthenticationFailed is returned when authentication fails.
var ErrAuthenticationFailed = errors.New("authentication failed")

// ErrAuthenticationRequired is returned when authentication is required but not provided.
var ErrAuthenticationRequired = errors.New("authentication required")
