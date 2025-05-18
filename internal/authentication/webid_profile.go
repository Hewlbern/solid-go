package auth

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// WebIDProfile represents a parsed WebID profile document.
// It contains the WebID URI, name, email addresses, and public keys of an agent.
type WebIDProfile struct {
	// WebID is the URI of the WebID
	WebID string
	// Name is the name of the agent
	Name string
	// Emails are the email addresses associated with the agent
	Emails []string
	// PublicKeys are the public keys associated with the agent
	PublicKeys []string
}

// WebIDProfileError represents an error in WebID profile handling.
// It contains a message describing the error and the underlying error.
type WebIDProfileError struct {
	// Message describes the error
	Message string
	// Err is the underlying error
	Err error
}

// Error returns a string representation of the error.
// It includes both the message and the underlying error if present.
func (e *WebIDProfileError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("webid profile error: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("webid profile error: %s", e.Message)
}

// Unwrap returns the underlying error.
// This allows errors to be unwrapped using errors.Unwrap.
func (e *WebIDProfileError) Unwrap() error {
	return e.Err
}

// FetchWebIDProfile fetches and parses a WebID profile document.
// It makes an HTTP request to fetch the profile and parses it into a WebIDProfile.
// It returns the parsed profile and any error that occurred.
//
// Example:
//
//	profile, err := FetchWebIDProfile("https://example.org/alice#me")
//	if err != nil {
//	    // Handle error
//	}
func FetchWebIDProfile(webID string) (*WebIDProfile, error) {
	// Make HTTP request to fetch the profile
	resp, err := http.Get(webID)
	if err != nil {
		return nil, &WebIDProfileError{
			Message: "failed to fetch WebID profile",
			Err:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &WebIDProfileError{
			Message: fmt.Sprintf("failed to fetch WebID profile: status code %d", resp.StatusCode),
		}
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &WebIDProfileError{
			Message: "failed to read WebID profile",
			Err:     err,
		}
	}

	// Parse the profile document
	profile, err := parseWebIDProfile(webID, string(body))
	if err != nil {
		return nil, err
	}

	return profile, nil
}

// parseWebIDProfile parses a WebID profile document.
// It extracts the name, email addresses, and public keys from the content.
// It returns the parsed profile and any error that occurred.
func parseWebIDProfile(webID, content string) (*WebIDProfile, error) {
	profile := &WebIDProfile{
		WebID: webID,
	}

	// Extract name from content
	if name := extractNameFromContent(content); name != "" {
		profile.Name = name
	}

	// Extract emails from content
	profile.Emails = extractEmailsFromContent(content)

	// Extract public keys from content
	profile.PublicKeys = extractPublicKeysFromContent(content)

	// Validate that we found at least one public key
	if len(profile.PublicKeys) == 0 {
		return nil, &WebIDProfileError{
			Message: "no public keys found in profile",
		}
	}

	return profile, nil
}

// extractNameFromContent extracts the name from the profile content.
// It looks for common name patterns in the content and returns the first match.
// If no name is found, it returns an empty string.
func extractNameFromContent(content string) string {
	// Look for common name patterns
	patterns := []string{
		`<foaf:name>(.*?)</foaf:name>`,
		`"name":\s*"(.*?)"`,
		`foaf:name\s+"(.*?)"`,
	}

	for _, pattern := range patterns {
		if matches := regexp.MustCompile(pattern).FindStringSubmatch(content); len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// extractEmailsFromContent extracts email addresses from the profile content.
// It looks for common email patterns in the content and returns all matches.
// If no emails are found, it returns an empty slice.
func extractEmailsFromContent(content string) []string {
	var emails []string

	// Look for common email patterns
	patterns := []string{
		`<foaf:mbox>(.*?)</foaf:mbox>`,
		`"email":\s*"(.*?)"`,
		`foaf:mbox\s+<mailto:(.*?)>`,
	}

	for _, pattern := range patterns {
		matches := regexp.MustCompile(pattern).FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				emails = append(emails, match[1])
			}
		}
	}

	return emails
}

// extractPublicKeysFromContent extracts public keys from the profile content.
// It looks for common public key patterns in the content and returns all matches.
// If no public keys are found, it returns an empty slice.
func extractPublicKeysFromContent(content string) []string {
	var keys []string

	// Look for common public key patterns
	patterns := []string{
		`<cert:key>(.*?)</cert:key>`,
		`"publicKey":\s*"(.*?)"`,
		`cert:key\s+"""([^"]*)"""`,
	}

	for _, pattern := range patterns {
		matches := regexp.MustCompile(pattern).FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				keys = append(keys, match[1])
			}
		}
	}

	return keys
}

// VerifyCertificate verifies a certificate against a WebID profile.
// It checks if the certificate's public key matches any in the profile.
// It returns nil if the certificate is verified, or an error if not.
//
// Example:
//
//	err := VerifyCertificate(cert, profile)
//	if err != nil {
//	    // Handle error
//	}
func VerifyCertificate(cert *x509.Certificate, profile *WebIDProfile) error {
	if cert == nil {
		return &WebIDProfileError{
			Message: "certificate is nil",
		}
	}

	if profile == nil {
		return &WebIDProfileError{
			Message: "profile is nil",
		}
	}

	// Convert certificate to PEM format
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	// Check if the certificate's public key matches any in the profile
	for _, key := range profile.PublicKeys {
		if strings.Contains(key, string(certPEM)) {
			return nil
		}
	}

	return &WebIDProfileError{
		Message: "certificate not found in WebID profile",
	}
}
