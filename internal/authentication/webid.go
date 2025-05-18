package auth

import (
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/yourusername/solid-go/internal/storage"
)

// WebIDStrategy implements the WebID-TLS authentication strategy.
// It authenticates users based on their WebID URI and TLS client certificate.
// The strategy verifies that the certificate is valid and matches the WebID profile.
type WebIDStrategy struct {
	// storage is the storage for WebID profiles
	storage storage.Storage
	// enabled indicates whether the strategy is enabled
	enabled bool
	// httpClient is the HTTP client to use for fetching WebID profiles
	httpClient *http.Client
}

// NewWebIDStrategy creates a new WebID-TLS authentication strategy.
// It takes a storage instance for storing WebID profiles and returns a new strategy.
//
// Example:
//
//	storage := storage.NewMemoryStorage()
//	strategy, err := NewWebIDStrategy(storage)
//	if err != nil {
//	    // Handle error
//	}
func NewWebIDStrategy(storage storage.Storage) (Strategy, error) {
	return &WebIDStrategy{
		storage:    storage,
		enabled:    true,
		httpClient: &http.Client{},
	}, nil
}

// Name returns the name of the strategy.
// It returns "webid" to identify this as a WebID-TLS strategy.
func (s *WebIDStrategy) Name() string {
	return "webid"
}

// IsEnabled returns whether the strategy is enabled.
// It returns true if the strategy is enabled, false otherwise.
func (s *WebIDStrategy) IsEnabled() bool {
	return s.enabled
}

// Authenticate authenticates a request using WebID-TLS.
// It checks for a client certificate, extracts the WebID URI, and verifies it against the profile.
// If successful, it returns an authenticated agent. If not, it returns an error.
//
// Example:
//
//	agent, err := strategy.Authenticate(r)
//	if err != nil {
//	    // Handle authentication error
//	}
func (s *WebIDStrategy) Authenticate(r *http.Request) (*Agent, error) {
	// Check if the client provided a certificate
	clientCertificates := r.TLS.PeerCertificates
	if len(clientCertificates) == 0 {
		return nil, &AuthError{
			Strategy: s.Name(),
			Message:  "no client certificate provided",
			Err:      ErrAuthenticationRequired,
		}
	}

	// Get the client certificate
	clientCert := clientCertificates[0]

	// Extract WebID from the certificate
	webID, err := extractWebIDFromCert(clientCert)
	if err != nil {
		return nil, &AuthError{
			Strategy: s.Name(),
			Message:  "failed to extract WebID from certificate",
			Err:      err,
		}
	}

	// Verify WebID
	agent, err := s.verifyWebID(webID, clientCert)
	if err != nil {
		return nil, &AuthError{
			Strategy: s.Name(),
			Message:  "WebID verification failed",
			Err:      err,
		}
	}

	return agent, nil
}

// extractWebIDFromCert extracts the WebID URI from a client certificate.
// It first checks the Subject Alternative Names for a URI, then falls back to the Common Name.
// It returns the WebID URI if found, or an error if no valid WebID is found.
func extractWebIDFromCert(cert *x509.Certificate) (string, error) {
	// Check Subject Alternative Names for a URI
	for _, uri := range cert.URIs {
		if isValidWebID(uri.String()) {
			return uri.String(), nil
		}
	}

	// Check Common Name as a fallback
	if isValidWebID(cert.Subject.CommonName) {
		return cert.Subject.CommonName, nil
	}

	return "", errors.New("no valid WebID found in certificate")
}

// isValidWebID checks if a URI is a valid WebID.
// A valid WebID is an HTTP(S) URI.
func isValidWebID(uri string) bool {
	// A valid WebID is an HTTP(S) URI
	return strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://")
}

// verifyWebID verifies a WebID by fetching the profile document and checking the certificate.
// It fetches the WebID profile, verifies the certificate against it, and creates an agent.
// If successful, it returns an authenticated agent. If not, it returns an error.
func (s *WebIDStrategy) verifyWebID(webID string, cert *x509.Certificate) (*Agent, error) {
	// Fetch the WebID profile
	profile, err := FetchWebIDProfile(webID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch WebID profile: %w", err)
	}

	// Verify the certificate against the profile
	if err := VerifyCertificate(cert, profile); err != nil {
		return nil, fmt.Errorf("certificate verification failed: %w", err)
	}

	// Create an agent from the profile
	agent := &Agent{
		ID:              webID,
		Name:            profile.Name,
		Email:           "", // Would extract from profile
		IsAuthenticated: true,
		Type:            TypeUser,
		Metadata:        make(map[string]interface{}),
	}

	// Add email if available
	if len(profile.Emails) > 0 {
		agent.Email = profile.Emails[0]
	}

	// Add additional metadata
	agent.Metadata["webid"] = webID
	agent.Metadata["emails"] = profile.Emails

	return agent, nil
}

// extractNameFromWebID extracts a name from a WebID URI.
// It tries to extract a name from the fragment identifier, path component, or host component.
// If no name can be extracted, it returns "Unknown".
func extractNameFromWebID(webID string) string {
	// Extract the fragment identifier
	if idx := strings.IndexByte(webID, '#'); idx >= 0 {
		fragment := webID[idx+1:]
		if fragment != "" {
			return fragment
		}
	}

	// Extract the path component
	if match := regexp.MustCompile(`/([^/]+)/?$`).FindStringSubmatch(webID); len(match) > 1 {
		return match[1]
	}

	// Extract the host component
	if match := regexp.MustCompile(`^https?://([^/]+)`).FindStringSubmatch(webID); len(match) > 1 {
		return match[1]
	}

	return "Unknown"
}

// MockWebIDStrategy is a mock implementation of the WebID-TLS strategy for testing.
// It always returns a predefined agent for all authentication requests.
type MockWebIDStrategy struct {
	// agent is the agent to return for all authentication requests
	agent *Agent
	// enabled indicates whether the strategy is enabled
	enabled bool
}

// NewMockWebIDStrategy creates a new mock WebID-TLS authentication strategy.
// It takes an agent to return for all authentication requests and returns a new strategy.
//
// Example:
//
//	agent := &Agent{ID: "https://example.org/alice#me", IsAuthenticated: true}
//	strategy := NewMockWebIDStrategy(agent)
func NewMockWebIDStrategy(agent *Agent) Strategy {
	return &MockWebIDStrategy{
		agent:   agent,
		enabled: true,
	}
}

// Name returns the name of the strategy.
// It returns "webid" to identify this as a WebID-TLS strategy.
func (s *MockWebIDStrategy) Name() string {
	return "webid"
}

// IsEnabled returns whether the strategy is enabled.
// It returns true if the strategy is enabled, false otherwise.
func (s *MockWebIDStrategy) IsEnabled() bool {
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
func (s *MockWebIDStrategy) Authenticate(r *http.Request) (*Agent, error) {
	if !s.enabled {
		return nil, fmt.Errorf("WebID authentication is disabled")
	}
	return s.agent, nil
}
