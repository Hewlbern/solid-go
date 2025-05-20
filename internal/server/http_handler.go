package server

import (
	"net/http"
)

// HttpHandler defines the interface for HTTP request handlers
type HttpHandler interface {
	// Handle processes an HTTP request and writes the response
	Handle(w http.ResponseWriter, r *http.Request) error
}

// HttpRequest represents an HTTP request with additional metadata
type HttpRequest struct {
	*http.Request
	Method      string
	Target      string
	Body        []byte
	ContentType string
	Headers     http.Header
}

// HttpResponse represents an HTTP response with additional metadata
type HttpResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}
