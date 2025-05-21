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
	handlers map[string]http.Handler
}

// NewHandlerServerConfigurator creates a new HandlerServerConfigurator
func NewHandlerServerConfigurator() *HandlerServerConfigurator {
	return &HandlerServerConfigurator{
		handlers: make(map[string]http.Handler),
	}
}

// AddHandler adds a handler for a path
func (c *HandlerServerConfigurator) AddHandler(path string, handler http.Handler) {
	c.handlers[path] = handler
}

// Configure implements ServerConfigurator
func (c *HandlerServerConfigurator) Configure(mux *http.ServeMux) {
	for path, handler := range c.handlers {
		mux.Handle(path, handler)
	}
}
