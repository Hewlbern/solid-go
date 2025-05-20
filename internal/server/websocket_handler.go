package server

import (
	"net/http"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler interface {
	// HandleWebSocket handles a WebSocket connection
	HandleWebSocket(w http.ResponseWriter, r *http.Request) error
}

// BaseWebSocketHandler provides a base implementation of WebSocketHandler
type BaseWebSocketHandler struct {
	upgrader http.Handler
}

// NewBaseWebSocketHandler creates a new BaseWebSocketHandler
func NewBaseWebSocketHandler(upgrader http.Handler) *BaseWebSocketHandler {
	return &BaseWebSocketHandler{
		upgrader: upgrader,
	}
}

// HandleWebSocket implements WebSocketHandler
func (h *BaseWebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) error {
	// Upgrade the HTTP connection to a WebSocket connection
	h.upgrader.ServeHTTP(w, r)
	return nil
}
