package identifiers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

// IdentifierUtil provides utility functions for generating identifiers
type IdentifierUtil struct{}

// NewIdentifierUtil creates a new IdentifierUtil
func NewIdentifierUtil() *IdentifierUtil {
	return &IdentifierUtil{}
}

// GenerateUUID generates a UUID v4
func (i *IdentifierUtil) GenerateUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

// GenerateRandomString generates a random string of the specified length
func (i *IdentifierUtil) GenerateRandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}

// GenerateTimestampID generates an ID with a timestamp prefix
func (i *IdentifierUtil) GenerateTimestampID(prefix string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d", prefix, timestamp)
}

// GenerateSequentialID generates a sequential ID with a prefix
func (i *IdentifierUtil) GenerateSequentialID(prefix string, sequence int) string {
	return fmt.Sprintf("%s-%d", prefix, sequence)
}

// GenerateCompositeID generates a composite ID from multiple parts
func (i *IdentifierUtil) GenerateCompositeID(parts ...string) string {
	return strings.Join(parts, "-")
}

// GenerateHashID generates a hash-based ID from a string
func (i *IdentifierUtil) GenerateHashID(input string) string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", b)
}

// GenerateSlug generates a URL-friendly slug from a string
func (i *IdentifierUtil) GenerateSlug(input string) string {
	// Convert to lowercase
	slug := strings.ToLower(input)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove special characters
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)

	// Remove multiple hyphens
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	return slug
}

// GeneratePathID generates a path-based ID
func (i *IdentifierUtil) GeneratePathID(parts ...string) string {
	return strings.Join(parts, "/")
}

// GenerateNamespaceID generates a namespace-based ID
func (i *IdentifierUtil) GenerateNamespaceID(namespace string, id string) string {
	return fmt.Sprintf("%s:%s", namespace, id)
}

// GenerateVersionedID generates a versioned ID
func (i *IdentifierUtil) GenerateVersionedID(id string, version int) string {
	return fmt.Sprintf("%s-v%d", id, version)
}

// GenerateTemporaryID generates a temporary ID
func (i *IdentifierUtil) GenerateTemporaryID(prefix string) string {
	timestamp := time.Now().UnixNano()
	random, _ := i.GenerateRandomString(8)
	return fmt.Sprintf("%s-temp-%d-%s", prefix, timestamp, random)
}

// GenerateReferenceID generates a reference ID
func (i *IdentifierUtil) GenerateReferenceID(source string, target string) string {
	return fmt.Sprintf("ref-%s-to-%s", source, target)
}

// GenerateCompositeReferenceID generates a composite reference ID
func (i *IdentifierUtil) GenerateCompositeReferenceID(source string, target string, relationship string) string {
	return fmt.Sprintf("ref-%s-%s-%s", source, relationship, target)
}
