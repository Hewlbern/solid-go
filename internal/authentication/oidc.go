package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/yourusername/solid-go/internal/storage"
)

// OIDCStrategy implements the OpenID Connect authentication strategy.
// It authenticates users based on OpenID Connect tokens and session cookies.
// The strategy supports both Bearer tokens and session cookies for authentication.
type OIDCStrategy struct {
	// storage is the storage for OIDC data
	storage storage.Storage
	// enabled indicates whether the strategy is enabled
	enabled bool
	// sessions maps session tokens to agent IDs
	sessions map[string]sessionInfo
}

// sessionInfo stores information about an active session.
// It contains the agent ID and expiration time for a session.
type sessionInfo struct {
	// agentID is the ID of the authenticated agent
	agentID string
	// expires is the time when the session expires
	expires time.Time
}

// NewOIDCStrategy creates a new OpenID Connect authentication strategy.
// It takes a storage instance for storing OIDC data and returns a new strategy.
//
// Example:
//
//	storage := storage.NewMemoryStorage()
//	strategy, err := NewOIDCStrategy(storage)
//	if err != nil {
//	    // Handle error
//	}
func NewOIDCStrategy(storage storage.Storage) (Strategy, error) {
	return &OIDCStrategy{
		storage:  storage,
		enabled:  true,
		sessions: make(map[string]sessionInfo),
	}, nil
}

// Name returns the name of the strategy.
// It returns "oidc" to identify this as an OpenID Connect strategy.
func (s *OIDCStrategy) Name() string {
	return "oidc"
}

// IsEnabled returns whether the strategy is enabled.
// It returns true if the strategy is enabled, false otherwise.
func (s *OIDCStrategy) IsEnabled() bool {
	return s.enabled
}

// Authenticate authenticates a request using OpenID Connect.
// It first checks for an Authorization header with a Bearer token.
// If not found, it checks for a session cookie.
// If successful, it returns an authenticated agent. If not, it returns an error.
//
// Example:
//
//	agent, err := strategy.Authenticate(r)
//	if err != nil {
//	    // Handle authentication error
//	}
func (s *OIDCStrategy) Authenticate(r *http.Request) (*Agent, error) {
	// Get the authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// Check for a session cookie
		cookie, err := r.Cookie("solid_session")
		if err != nil || cookie.Value == "" {
			return nil, &AuthError{
				Strategy: s.Name(),
				Message:  "no authorization header or session cookie provided",
				Err:      ErrAuthenticationRequired,
			}
		}
		return s.validateSession(cookie.Value)
	}

	// Parse the authorization header
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, &AuthError{
			Strategy: s.Name(),
			Message:  "invalid authorization header format",
			Err:      ErrAuthenticationRequired,
		}
	}

	// Get the token
	token := parts[1]

	// Validate the token
	return s.validateToken(token)
}

// validateToken validates an OpenID Connect token.
// It checks if the token is valid and not expired.
// If successful, it returns an authenticated agent. If not, it returns an error.
func (s *OIDCStrategy) validateToken(token string) (*Agent, error) {
	// In a real implementation, we would:
	// 1. Verify the token signature
	// 2. Check if the token is expired
	// 3. Verify the issuer
	// 4. Extract claims from the token

	// For this simplified implementation, we'll just check if the token is in our session map
	// In a real implementation, we would decode and verify the JWT token

	// TODO: Implement proper OIDC token validation

	// Check if the token is in our session map
	session, exists := s.sessions[token]
	if !exists {
		return nil, &AuthError{
			Strategy: s.Name(),
			Message:  "invalid token",
			Err:      ErrAuthenticationFailed,
		}
	}

	// Check if the session has expired
	if time.Now().After(session.expires) {
		delete(s.sessions, token)
		return nil, &AuthError{
			Strategy: s.Name(),
			Message:  "token expired",
			Err:      ErrAuthenticationFailed,
		}
	}

	// Get the agent information
	return s.getAgentForID(session.agentID)
}

// validateSession validates a session cookie.
// It checks if the session is valid and not expired.
// If successful, it returns an authenticated agent. If not, it returns an error.
func (s *OIDCStrategy) validateSession(sessionID string) (*Agent, error) {
	// This is similar to validateToken, but for session cookies
	return s.validateToken(sessionID)
}

// getAgentForID gets agent information for a given agent ID.
// It creates an agent from the ID and returns it.
// In a real implementation, it would fetch the agent information from storage.
func (s *OIDCStrategy) getAgentForID(agentID string) (*Agent, error) {
	// In a real implementation, we would fetch the agent information from storage
	// For this simplified implementation, we'll just create an agent from the ID

	// TODO: Implement proper agent information retrieval

	// Extract name from agent ID
	name := extractNameFromWebID(agentID)

	return &Agent{
		ID:              agentID,
		Name:            name,
		Email:           "", // Would extract from profile
		IsAuthenticated: true,
		Type:            TypeUser,
		Metadata:        make(map[string]interface{}),
	}, nil
}

// RegisterSession registers a new session for an agent.
// It generates a session token and stores it with the agent ID and expiration time.
// It returns the session token and any error that occurred.
//
// Example:
//
//	token, err := strategy.RegisterSession(agentID, 24*time.Hour)
//	if err != nil {
//	    // Handle error
//	}
func (s *OIDCStrategy) RegisterSession(agentID string, duration time.Duration) (string, error) {
	if agentID == "" {
		return "", errors.New("agent ID cannot be empty")
	}

	// Generate a session token
	token := generateToken()

	// Register the session
	s.sessions[token] = sessionInfo{
		agentID: agentID,
		expires: time.Now().Add(duration),
	}

	return token, nil
}

// generateToken generates a random token.
// In a real implementation, it would generate a secure random token.
// For this simplified implementation, it just uses a timestamp.
func generateToken() string {
	// In a real implementation, we would generate a secure random token
	// For this simplified implementation, we'll just use a timestamp

	// TODO: Implement proper secure token generation
	return fmt.Sprintf("token_%d", time.Now().UnixNano())
}

// SolidOIDCConfig represents the configuration for Solid OIDC.
// It contains the OIDC issuer URI, client ID, client secret, and redirect URI.
type SolidOIDCConfig struct {
	// Issuer is the OIDC issuer URI
	Issuer string `json:"issuer"`
	// ClientID is the OIDC client ID
	ClientID string `json:"client_id"`
	// ClientSecret is the OIDC client secret
	ClientSecret string `json:"client_secret"`
	// RedirectURI is the OIDC redirect URI
	RedirectURI string `json:"redirect_uri"`
}

// SaveOIDCConfig saves the OIDC configuration to storage.
// It serializes the config to JSON and stores it in the storage.
// It returns any error that occurred.
//
// Example:
//
//	config := &SolidOIDCConfig{
//	    Issuer: "https://example.org",
//	    ClientID: "client_id",
//	    ClientSecret: "client_secret",
//	    RedirectURI: "https://example.org/callback",
//	}
//	err := strategy.SaveOIDCConfig(config)
//	if err != nil {
//	    // Handle error
//	}
func (s *OIDCStrategy) SaveOIDCConfig(config *SolidOIDCConfig) error {
	// Serialize the config to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to serialize OIDC config: %w", err)
	}

	// Store the config
	return s.storage.StoreResource("/.well-known/solid-oidc-config.json", configJSON, "application/json")
}

// LoadOIDCConfig loads the OIDC configuration from storage.
// It loads the config from storage and deserializes it from JSON.
// It returns the config and any error that occurred.
//
// Example:
//
//	config, err := strategy.LoadOIDCConfig()
//	if err != nil {
//	    // Handle error
//	}
func (s *OIDCStrategy) LoadOIDCConfig() (*SolidOIDCConfig, error) {
	// Load the config
	configJSON, _, err := s.storage.GetResource("/.well-known/solid-oidc-config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load OIDC config: %w", err)
	}

	// Deserialize the config
	var config SolidOIDCConfig
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return nil, fmt.Errorf("failed to deserialize OIDC config: %w", err)
	}

	return &config, nil
}

// MockOIDCStrategy is a mock implementation of the OIDC strategy for testing.
// It always returns a predefined agent for all authentication requests.
type MockOIDCStrategy struct {
	// agent is the agent to return for all authentication requests
	agent *Agent
	// enabled indicates whether the strategy is enabled
	enabled bool
}

// NewMockOIDCStrategy creates a new mock OIDC authentication strategy.
// It takes an agent to return for all authentication requests and returns a new strategy.
//
// Example:
//
//	agent := &Agent{ID: "https://example.org/alice#me", IsAuthenticated: true}
//	strategy := NewMockOIDCStrategy(agent)
func NewMockOIDCStrategy(agent *Agent) Strategy {
	return &MockOIDCStrategy{
		agent:   agent,
		enabled: true,
	}
}

// Name returns the name of the strategy.
// It returns "oidc" to identify this as an OpenID Connect strategy.
func (s *MockOIDCStrategy) Name() string {
	return "oidc"
}

// IsEnabled returns whether the strategy is enabled.
// It returns true if the strategy is enabled, false otherwise.
func (s *MockOIDCStrategy) IsEnabled() bool {
	return s.enabled
}

// Authenticate returns the predefined agent.
// It returns the agent if the strategy is enabled, or an error if it is disabled.
//
// Example:
//
//	agent, err := strategy.Authenticate(r)
//	if err != nil {
//	    // Handle authentication error
//	}
func (s *MockOIDCStrategy) Authenticate(r *http.Request) (*Agent, error) {
	if !s.enabled {
		return nil, fmt.Errorf("OIDC authentication is disabled")
	}
	return s.agent, nil
}
