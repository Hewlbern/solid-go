package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestWebIDStrategy_Authenticate(t *testing.T) {
	// Create a test certificate
	cert, err := createTestCertificate("https://example.com/card#me")
	if err != nil {
		t.Fatalf("createTestCertificate failed: %v", err)
	}

	// Create a test server that returns a WebID profile
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/turtle")
		w.Write([]byte(`
			@prefix foaf: <http://xmlns.com/foaf/0.1/> .
			@prefix cert: <http://www.w3.org/ns/auth/cert#> .
			@prefix rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#> .

			<https://example.com/card#me>
				a foaf:Person ;
				foaf:name "Test User" ;
				foaf:mbox <mailto:test@example.com> ;
				cert:key """` + string(pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		})) + `""" .
		`))
	}))
	defer server.Close()

	// Create a test request with the certificate
	req := httptest.NewRequest("GET", "/", nil)
	req.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{cert},
	}

	// Create the WebID strategy
	strategy, err := NewWebIDStrategy(nil)
	if err != nil {
		t.Fatalf("NewWebIDStrategy failed: %v", err)
	}

	// Authenticate the request
	agent, err := strategy.Authenticate(req)
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}
	if agent == nil {
		t.Fatal("agent is nil")
	}
	if agent.ID != "https://example.com/card#me" {
		t.Errorf("ID = %v, want %v", agent.ID, "https://example.com/card#me")
	}
	if agent.Name != "Test User" {
		t.Errorf("Name = %v, want %v", agent.Name, "Test User")
	}
	if agent.Email != "test@example.com" {
		t.Errorf("Email = %v, want %v", agent.Email, "test@example.com")
	}
	if !agent.IsAuthenticated {
		t.Error("IsAuthenticated = false, want true")
	}
	if agent.Type != TypeUser {
		t.Errorf("Type = %v, want %v", agent.Type, TypeUser)
	}
}

func TestWebIDStrategy_Authenticate_NoCertificate(t *testing.T) {
	// Create a test request without a certificate
	req := httptest.NewRequest("GET", "/", nil)

	// Create the WebID strategy
	strategy, err := NewWebIDStrategy(nil)
	if err != nil {
		t.Fatalf("NewWebIDStrategy failed: %v", err)
	}

	// Authenticate the request
	agent, err := strategy.Authenticate(req)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if agent != nil {
		t.Errorf("agent = %v, want nil", agent)
	}
	if !strings.Contains(err.Error(), "no client certificate provided") {
		t.Errorf("error = %v, want to contain 'no client certificate provided'", err)
	}
}

func TestWebIDStrategy_Authenticate_InvalidCertificate(t *testing.T) {
	// Create a test certificate without a WebID
	cert, err := createTestCertificate("")
	if err != nil {
		t.Fatalf("createTestCertificate failed: %v", err)
	}

	// Create a test request with the certificate
	req := httptest.NewRequest("GET", "/", nil)
	req.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{cert},
	}

	// Create the WebID strategy
	strategy, err := NewWebIDStrategy(nil)
	if err != nil {
		t.Fatalf("NewWebIDStrategy failed: %v", err)
	}

	// Authenticate the request
	agent, err := strategy.Authenticate(req)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if agent != nil {
		t.Errorf("agent = %v, want nil", agent)
	}
	if !strings.Contains(err.Error(), "no valid WebID found in certificate") {
		t.Errorf("error = %v, want to contain 'no valid WebID found in certificate'", err)
	}
}

func TestWebIDStrategy_Authenticate_InvalidProfile(t *testing.T) {
	// Create a test certificate
	cert, err := createTestCertificate("https://example.com/card#me")
	if err != nil {
		t.Fatalf("createTestCertificate failed: %v", err)
	}

	// Create a test server that returns an invalid WebID profile
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Create a test request with the certificate
	req := httptest.NewRequest("GET", "/", nil)
	req.TLS = &tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{cert},
	}

	// Create the WebID strategy
	strategy, err := NewWebIDStrategy(nil)
	if err != nil {
		t.Fatalf("NewWebIDStrategy failed: %v", err)
	}

	// Authenticate the request
	agent, err := strategy.Authenticate(req)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if agent != nil {
		t.Errorf("agent = %v, want nil", agent)
	}
	if !strings.Contains(err.Error(), "failed to fetch WebID profile") {
		t.Errorf("error = %v, want to contain 'failed to fetch WebID profile'", err)
	}
}

// Helper function to create a test certificate
func createTestCertificate(webID string) (*x509.Certificate, error) {
	// Generate a private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	// Create a certificate template
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   webID,
			Organization: []string{"Test Organization"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// Add WebID to Subject Alternative Names if provided
	if webID != "" {
		template.URIs = []*url.URL{
			{
				Scheme:   "https",
				Host:     "example.com",
				Path:     "/card",
				Fragment: "me",
			},
		}
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, err
	}

	return cert, nil
}
