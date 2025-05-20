package server

import (
	"net/http"
)

// WacAllowHttpHandler handles WAC allow requests
type WacAllowHttpHandler struct {
	handler    HttpHandler
	authorizer Authorizer
}

// NewWacAllowHttpHandler creates a new WacAllowHttpHandler
func NewWacAllowHttpHandler(handler HttpHandler, authorizer Authorizer) *WacAllowHttpHandler {
	return &WacAllowHttpHandler{
		handler:    handler,
		authorizer: authorizer,
	}
}

// Handle implements HttpHandler
func (h *WacAllowHttpHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	// Convert request to operation
	op := Operation{
		Method:      r.Method,
		Target:      r.URL.Path,
		ContentType: r.Header.Get("Content-Type"),
		Headers:     r.Header,
	}

	// Check if the request is allowed
	if err := h.authorizer.Authorize(r.Context(), op); err != nil {
		// If not allowed, check if it's a HEAD request
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return nil
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return err
	}

	// Handle the request
	return h.handler.Handle(w, r)
}
