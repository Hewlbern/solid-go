package server

import (
	"net/http"
)

// AuthorizingHttpHandler handles HTTP requests with authorization
type AuthorizingHttpHandler interface {
	HttpHandler
	// Authorize checks if the request is authorized
	Authorize(r *http.Request) error
}

// BaseAuthorizingHandler provides a base implementation of AuthorizingHttpHandler
type BaseAuthorizingHandler struct {
	handler    HttpHandler
	authorizer Authorizer
}

// NewBaseAuthorizingHandler creates a new BaseAuthorizingHandler
func NewBaseAuthorizingHandler(handler HttpHandler, authorizer Authorizer) *BaseAuthorizingHandler {
	return &BaseAuthorizingHandler{
		handler:    handler,
		authorizer: authorizer,
	}
}

// Handle implements HttpHandler
func (h *BaseAuthorizingHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	// Convert request to operation
	op := Operation{
		Method:      r.Method,
		Target:      r.URL.Path,
		ContentType: r.Header.Get("Content-Type"),
		Headers:     r.Header,
	}

	// Authorize the request
	if err := h.authorizer.Authorize(r.Context(), op); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return err
	}

	// Handle the request
	return h.handler.Handle(w, r)
}

// Authorize implements AuthorizingHttpHandler
func (h *BaseAuthorizingHandler) Authorize(r *http.Request) error {
	op := Operation{
		Method:      r.Method,
		Target:      r.URL.Path,
		ContentType: r.Header.Get("Content-Type"),
		Headers:     r.Header,
	}
	return h.authorizer.Authorize(r.Context(), op)
}
