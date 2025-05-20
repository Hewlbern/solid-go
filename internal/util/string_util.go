package util

import (
	"strings"
	"unicode"
)

// StringUtil provides utility functions for string manipulation
type StringUtil struct{}

// NewStringUtil creates a new StringUtil
func NewStringUtil() *StringUtil {
	return &StringUtil{}
}

// IsEmpty checks if a string is empty or contains only whitespace
func (s *StringUtil) IsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

// IsNotEmpty checks if a string is not empty and contains non-whitespace characters
func (s *StringUtil) IsNotEmpty(str string) bool {
	return !s.IsEmpty(str)
}

// IsBlank checks if a string is blank (empty or contains only whitespace)
func (s *StringUtil) IsBlank(str string) bool {
	return s.IsEmpty(str)
}

// IsNotBlank checks if a string is not blank
func (s *StringUtil) IsNotBlank(str string) bool {
	return s.IsNotEmpty(str)
}

// Trim trims whitespace from both ends of a string
func (s *StringUtil) Trim(str string) string {
	return strings.TrimSpace(str)
}

// ToLower converts a string to lowercase
func (s *StringUtil) ToLower(str string) string {
	return strings.ToLower(str)
}

// ToUpper converts a string to uppercase
func (s *StringUtil) ToUpper(str string) string {
	return strings.ToUpper(str)
}

// Contains checks if a string contains another string
func (s *StringUtil) Contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

// StartsWith checks if a string starts with a prefix
func (s *StringUtil) StartsWith(str, prefix string) bool {
	return strings.HasPrefix(str, prefix)
}

// EndsWith checks if a string ends with a suffix
func (s *StringUtil) EndsWith(str, suffix string) bool {
	return strings.HasSuffix(str, suffix)
}

// Split splits a string by a separator
func (s *StringUtil) Split(str, sep string) []string {
	return strings.Split(str, sep)
}

// Join joins strings with a separator
func (s *StringUtil) Join(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

// Replace replaces all occurrences of old with new in str
func (s *StringUtil) Replace(str, old, new string) string {
	return strings.ReplaceAll(str, old, new)
}

// IsAlpha checks if a string contains only alphabetic characters
func (s *StringUtil) IsAlpha(str string) bool {
	for _, r := range str {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// IsNumeric checks if a string contains only numeric characters
func (s *StringUtil) IsNumeric(str string) bool {
	for _, r := range str {
		if !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

// IsAlphaNumeric checks if a string contains only alphanumeric characters
func (s *StringUtil) IsAlphaNumeric(str string) bool {
	for _, r := range str {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}
