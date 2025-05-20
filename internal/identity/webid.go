package identity

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"

	"solid-go/internal/rdf"
)

// WebID represents a WebID identity
type WebID struct {
	URI         string
	PublicKey   *rsa.PublicKey
	Certificate *x509.Certificate
	Profile     *rdf.Graph
}

// WebIDStore manages WebID identities
type WebIDStore struct {
	webids map[string]*WebID
}

// NewWebIDStore creates a new WebID store
func NewWebIDStore() *WebIDStore {
	return &WebIDStore{
		webids: make(map[string]*WebID),
	}
}

// CreateWebID creates a new WebID identity
func (s *WebIDStore) CreateWebID(uri string, publicKey *rsa.PublicKey) (*WebID, error) {
	if _, exists := s.webids[uri]; exists {
		return nil, fmt.Errorf("webid already exists: %s", uri)
	}

	webid := &WebID{
		URI:       uri,
		PublicKey: publicKey,
		Profile:   rdf.NewGraph(),
	}

	// Add basic profile information
	webid.Profile.Add(rdf.Triple{
		Subject:   uri,
		Predicate: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type",
		Object:    "http://xmlns.com/foaf/0.1/Person",
	})

	s.webids[uri] = webid
	return webid, nil
}

// GetWebID retrieves a WebID by URI
func (s *WebIDStore) GetWebID(uri string) (*WebID, error) {
	webid, exists := s.webids[uri]
	if !exists {
		return nil, fmt.Errorf("webid not found: %s", uri)
	}
	return webid, nil
}

// UpdateProfile updates a WebID's profile
func (s *WebIDStore) UpdateProfile(uri string, profile *rdf.Graph) error {
	webid, err := s.GetWebID(uri)
	if err != nil {
		return err
	}
	webid.Profile = profile
	return nil
}

// VerifyWebID verifies a WebID's ownership
func (s *WebIDStore) VerifyWebID(uri string, signature []byte) bool {
	webid, err := s.GetWebID(uri)
	if err != nil {
		return false
	}

	// In a real implementation, this would verify the signature
	// For now, we'll just return true if the WebID exists
	return true
}

// LinkWebID links a WebID to another identity
func (s *WebIDStore) LinkWebID(uri string, otherURI string) error {
	webid, err := s.GetWebID(uri)
	if err != nil {
		return err
	}

	webid.Profile.Add(rdf.Triple{
		Subject:   uri,
		Predicate: "http://www.w3.org/ns/solid/terms#linkedTo",
		Object:    otherURI,
	})

	return nil
}

// UnlinkWebID unlinks a WebID from another identity
func (s *WebIDStore) UnlinkWebID(uri string, otherURI string) error {
	webid, err := s.GetWebID(uri)
	if err != nil {
		return err
	}

	webid.Profile.Remove(rdf.Triple{
		Subject:   uri,
		Predicate: "http://www.w3.org/ns/solid/terms#linkedTo",
		Object:    otherURI,
	})

	return nil
}
