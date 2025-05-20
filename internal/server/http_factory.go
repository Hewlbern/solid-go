package server

import (
	"net/http"
)

// HttpServerFactory creates HTTP servers
type HttpServerFactory struct {
	configurators []ServerConfigurator
}

// NewHttpServerFactory creates a new HttpServerFactory
func NewHttpServerFactory(configurators ...ServerConfigurator) *HttpServerFactory {
	return &HttpServerFactory{
		configurators: configurators,
	}
}

// CreateServer creates a new HTTP server
func (f *HttpServerFactory) CreateServer() *http.Server {
	mux := http.NewServeMux()

	// Apply all configurators
	for _, configurator := range f.configurators {
		configurator.Configure(mux)
	}

	return &http.Server{
		Handler: mux,
	}
}
