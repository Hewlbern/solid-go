package server

import (
	"net/http"
)

// ServerConfigurator configures server settings
type ServerConfigurator interface {
	// Configure configures the server
	Configure(mux *http.ServeMux)
}

// HandlerServerConfigurator configures server handlers
type HandlerServerConfigurator struct {
	path    string
	handler http.Handler
}

// NewHandlerServerConfigurator creates a new HandlerServerConfigurator
func NewHandlerServerConfigurator(path string, handler http.Handler) *HandlerServerConfigurator {
	return &HandlerServerConfigurator{
		path:    path,
		handler: handler,
	}
}

// Configure implements ServerConfigurator
func (c *HandlerServerConfigurator) Configure(mux *http.ServeMux) {
	mux.Handle(c.path, c.handler)
}
