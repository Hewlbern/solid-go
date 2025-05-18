package ldp

import (
	"net/http"

	"github.com/yourusername/solid-go/internal/wac"
)

// WACAuthenticator integrates the WAC system with LDP operations.
type WACAuthenticator struct {
	// wacHTTPHandler is the HTTP handler for WAC operations.
	wacHTTPHandler *wac.HTTPHandler
}

// NewWACAuthenticator creates a new WACAuthenticator.
func NewWACAuthenticator(wacHTTPHandler *wac.HTTPHandler) *WACAuthenticator {
	return &WACAuthenticator{
		wacHTTPHandler: wacHTTPHandler,
	}
}

// Middleware returns an HTTP middleware that checks if a request is authorized.
func (w *WACAuthenticator) Middleware() func(http.Handler) http.Handler {
	return w.wacHTTPHandler.ACLMiddleware
}

// AuthorizeRequest checks if a request is authorized.
func (w *WACAuthenticator) AuthorizeRequest(req *http.Request) (bool, error) {
	return w.wacHTTPHandler.AuthorizeRequest(req)
}

// RequireAuthentication returns an HTTP middleware that requires authentication.
func (w *WACAuthenticator) RequireAuthentication(handler http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// Check if the request is authorized
		authorized, err := w.AuthorizeRequest(req)
		if err != nil {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !authorized {
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			return
		}

		handler(rw, req)
	}
}

// WACHandler integrates the WAC system with LDP operations.
type WACHandler struct {
	// ldpHandler is the LDP handler.
	ldpHandler *Handler
	// wacAuthenticator is the WAC authenticator.
	wacAuthenticator *WACAuthenticator
}

// NewWACHandler creates a new WACHandler.
func NewWACHandler(ldpHandler *Handler, wacAuthenticator *WACAuthenticator) *WACHandler {
	return &WACHandler{
		ldpHandler:       ldpHandler,
		wacAuthenticator: wacAuthenticator,
	}
}

// RegisterRoutes registers the LDP routes with the given mux, protected by WAC.
func (h *WACHandler) RegisterRoutes(mux *http.ServeMux) {
	// Add the WAC middleware to all LDP routes
	middleware := h.wacAuthenticator.Middleware()

	// Register all LDP routes with WAC middleware
	mux.Handle("/", middleware(h.ldpHandler))
	mux.Handle("/containers/", middleware(h.ldpHandler))

	// For handling WebID profiles
	mux.Handle("/profiles/", middleware(h.ldpHandler))
}

// HandleResource handles requests for LDP resources, with WAC authorization.
func (h *WACHandler) HandleResource(w http.ResponseWriter, req *http.Request) {
	// Check if the request is authorized
	authorized, err := h.wacAuthenticator.AuthorizeRequest(req)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !authorized {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Handle the resource with the underlying LDP handler
	h.ldpHandler.ServeHTTP(w, req)
}

// HandleContainer handles requests for LDP containers, with WAC authorization.
func (h *WACHandler) HandleContainer(w http.ResponseWriter, req *http.Request) {
	// Check if the request is authorized
	authorized, err := h.wacAuthenticator.AuthorizeRequest(req)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !authorized {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Handle the container with the underlying LDP handler
	h.ldpHandler.ServeHTTP(w, req)
}

// HandleProfile handles requests for WebID profiles, with WAC authorization.
func (h *WACHandler) HandleProfile(w http.ResponseWriter, req *http.Request) {
	// Check if the request is authorized
	authorized, err := h.wacAuthenticator.AuthorizeRequest(req)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !authorized {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Handle the profile with the underlying LDP handler
	h.ldpHandler.ServeHTTP(w, req)
}
