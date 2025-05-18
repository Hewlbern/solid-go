package auth

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestFetchWebIDProfile tests the FetchWebIDProfile function.
// It creates a test server that returns a WebID profile and verifies that
// the profile is correctly parsed.
func TestFetchWebIDProfile(t *testing.T) {
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
			Bytes: []byte("test certificate"),
		})) + `""" .
		`))
	}))
	defer server.Close()

	// Fetch the WebID profile
	profile, err := FetchWebIDProfile(server.URL)
	if err != nil {
		t.Fatalf("FetchWebIDProfile failed: %v", err)
	}
	if profile == nil {
		t.Fatal("profile is nil")
	}
	if profile.WebID != server.URL {
		t.Errorf("WebID = %v, want %v", profile.WebID, server.URL)
	}
	if profile.Name != "Test User" {
		t.Errorf("Name = %v, want %v", profile.Name, "Test User")
	}
	if len(profile.Emails) != 1 || profile.Emails[0] != "test@example.com" {
		t.Errorf("Emails = %v, want %v", profile.Emails, []string{"test@example.com"})
	}
	if len(profile.PublicKeys) == 0 {
		t.Error("PublicKeys is empty")
	}
}

// TestFetchWebIDProfile_NotFound tests the FetchWebIDProfile function when the profile is not found.
// It creates a test server that returns a 404 and verifies that an error is returned.
func TestFetchWebIDProfile_NotFound(t *testing.T) {
	// Create a test server that returns a 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Fetch the WebID profile
	profile, err := FetchWebIDProfile(server.URL)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if profile != nil {
		t.Errorf("profile = %v, want nil", profile)
	}
	if !strings.Contains(err.Error(), "failed to fetch WebID profile") {
		t.Errorf("error = %v, want to contain 'failed to fetch WebID profile'", err)
	}
}

// TestFetchWebIDProfile_InvalidContent tests the FetchWebIDProfile function with invalid content.
// It creates a test server that returns invalid content and verifies that
// the profile is created with empty fields.
func TestFetchWebIDProfile_InvalidContent(t *testing.T) {
	// Create a test server that returns invalid content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid content"))
	}))
	defer server.Close()

	// Fetch the WebID profile
	profile, err := FetchWebIDProfile(server.URL)
	if err != nil {
		t.Fatalf("FetchWebIDProfile failed: %v", err)
	}
	if profile == nil {
		t.Fatal("profile is nil")
	}
	if profile.WebID != server.URL {
		t.Errorf("WebID = %v, want %v", profile.WebID, server.URL)
	}
	if profile.Name != "" {
		t.Errorf("Name = %v, want empty string", profile.Name)
	}
	if len(profile.Emails) != 0 {
		t.Errorf("Emails = %v, want empty slice", profile.Emails)
	}
	if len(profile.PublicKeys) != 0 {
		t.Errorf("PublicKeys = %v, want empty slice", profile.PublicKeys)
	}
}

// TestVerifyCertificate tests the VerifyCertificate function.
// It creates a test certificate and profile with a matching certificate
// and verifies that the certificate is accepted.
func TestVerifyCertificate(t *testing.T) {
	// Create a test certificate
	cert := &x509.Certificate{
		Raw: []byte("test certificate"),
	}

	// Create a test profile with a matching certificate
	profile := &WebIDProfile{
		WebID: "https://example.com/card#me",
		PublicKeys: []string{
			string(pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: []byte("test certificate"),
			})),
		},
	}

	// Verify the certificate
	err := VerifyCertificate(cert, profile)
	if err != nil {
		t.Errorf("VerifyCertificate failed: %v", err)
	}
}

// TestVerifyCertificate_NoMatch tests the VerifyCertificate function when the certificate does not match.
// It creates a test certificate and profile with a non-matching certificate
// and verifies that an error is returned.
func TestVerifyCertificate_NoMatch(t *testing.T) {
	// Create a test certificate
	cert := &x509.Certificate{
		Raw: []byte("test certificate"),
	}

	// Create a test profile with a non-matching certificate
	profile := &WebIDProfile{
		WebID: "https://example.com/card#me",
		PublicKeys: []string{
			string(pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: []byte("different certificate"),
			})),
		},
	}

	// Verify the certificate
	err := VerifyCertificate(cert, profile)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "certificate not found in WebID profile") {
		t.Errorf("error = %v, want to contain 'certificate not found in WebID profile'", err)
	}
}

// TestExtractNameFromContent tests the extractNameFromContent function.
// It tests various formats of name in the content and verifies that
// the name is correctly extracted.
func TestExtractNameFromContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "FOAF name",
			content: `
				@prefix foaf: <http://xmlns.com/foaf/0.1/> .
				<https://example.com/card#me> foaf:name "Test User" .
			`,
			expected: "Test User",
		},
		{
			name:     "JSON name",
			content:  `{"name": "Test User"}`,
			expected: "Test User",
		},
		{
			name:     "No name",
			content:  "invalid content",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := extractNameFromContent(tt.content)
			if name != tt.expected {
				t.Errorf("extractNameFromContent() = %v, want %v", name, tt.expected)
			}
		})
	}
}

// TestExtractEmailsFromContent tests the extractEmailsFromContent function.
// It tests various formats of email in the content and verifies that
// the emails are correctly extracted.
func TestExtractEmailsFromContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "FOAF mbox",
			content: `
				@prefix foaf: <http://xmlns.com/foaf/0.1/> .
				<https://example.com/card#me> foaf:mbox <mailto:test@example.com> .
			`,
			expected: []string{"test@example.com"},
		},
		{
			name:     "JSON email",
			content:  `{"email": "test@example.com"}`,
			expected: []string{"test@example.com"},
		},
		{
			name:     "No email",
			content:  "invalid content",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emails := extractEmailsFromContent(tt.content)
			if len(emails) != len(tt.expected) {
				t.Errorf("extractEmailsFromContent() length = %v, want %v", len(emails), len(tt.expected))
			}
			for i := range emails {
				if emails[i] != tt.expected[i] {
					t.Errorf("extractEmailsFromContent()[%d] = %v, want %v", i, emails[i], tt.expected[i])
				}
			}
		})
	}
}

// TestExtractPublicKeysFromContent tests the extractPublicKeysFromContent function.
// It tests various formats of public key in the content and verifies that
// the public keys are correctly extracted.
func TestExtractPublicKeysFromContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name: "Certificate key",
			content: `
				@prefix cert: <http://www.w3.org/ns/auth/cert#> .
				<https://example.com/card#me> cert:key """test key""" .
			`,
			expected: []string{"test key"},
		},
		{
			name:     "JSON public key",
			content:  `{"publicKey": "test key"}`,
			expected: []string{"test key"},
		},
		{
			name:     "No public key",
			content:  "invalid content",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := extractPublicKeysFromContent(tt.content)
			if len(keys) != len(tt.expected) {
				t.Errorf("extractPublicKeysFromContent() length = %v, want %v", len(keys), len(tt.expected))
			}
			for i := range keys {
				if keys[i] != tt.expected[i] {
					t.Errorf("extractPublicKeysFromContent()[%d] = %v, want %v", i, keys[i], tt.expected[i])
				}
			}
		})
	}
}
