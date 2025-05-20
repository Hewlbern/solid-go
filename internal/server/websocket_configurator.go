package server

import (
	"net/http"
)

// WebSocketServerConfigurator configures WebSocket server settings
type WebSocketServerConfigurator struct {
	handler WebSocketHandler
	path    string
}

// NewWebSocketServerConfigurator creates a new WebSocketServerConfigurator
func NewWebSocketServerConfigurator(handler WebSocketHandler, path string) *WebSocketServerConfigurator {
	return &WebSocketServerConfigurator{
		handler: handler,
		path:    path,
	}
}

// Configure configures the WebSocket server
func (c *WebSocketServerConfigurator) Configure(mux *http.ServeMux) {
	mux.HandleFunc(c.path, func(w http.ResponseWriter, r *http.Request) {
		if err := c.handler.HandleWebSocket(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
