package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"solid-go/internal/acl"
	"solid-go/internal/identity"
	"solid-go/internal/ldp"
	"solid-go/internal/logging"
	"solid-go/internal/storage"
)

// Server represents a Solid server
type Server struct {
	http.Server
	storage    storage.Storage
	webidStore *identity.WebIDStore
	aclStore   map[string]*acl.ACL
	containers map[string]*ldp.Container
	logger     logging.Logger
}

// ServerOptions represents options for creating a server
type ServerOptions struct {
	Port     int
	HTTPS    bool
	CertFile string
	KeyFile  string
	Storage  storage.Storage
	Logger   logging.Logger
}

// NewServer creates a new server with the given options
func NewServer(options *ServerOptions) *Server {
	if options.Logger == nil {
		options.Logger = logging.NewBasicLogger(logging.Info)
	}

	srv := &Server{
		Server: http.Server{
			Addr:         fmt.Sprintf(":%d", options.Port),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		storage:    options.Storage,
		webidStore: identity.NewWebIDStore(),
		aclStore:   make(map[string]*acl.ACL),
		containers: make(map[string]*ldp.Container),
		logger:     options.Logger,
	}

	// Set up routes
	mux := http.NewServeMux()
	srv.Handler = mux

	// Add handlers
	mux.HandleFunc("/", srv.handleRequest)

	return srv
}

// handleRequest handles incoming HTTP requests
func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("Received %s request for %s", r.Method, r.URL.Path)

	// Handle different request types
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	case http.MethodPost:
		s.handlePost(w, r)
	case http.MethodPut:
		s.handlePut(w, r)
	case http.MethodDelete:
		s.handleDelete(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGet handles GET requests
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Check if path exists
	exists, err := s.storage.Exists(r.Context(), path)
	if err != nil {
		s.logger.Error("Error checking path existence: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.NotFound(w, r)
		return
	}

	// Get resource
	data, err := s.storage.Get(r.Context(), path)
	if err != nil {
		s.logger.Error("Error getting resource: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "text/turtle")
	w.Write(data)
}

// handlePost handles POST requests
func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement POST handling
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// handlePut handles PUT requests
func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Read request body
	data, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("Error reading request body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Store resource
	err = s.storage.Put(r.Context(), path, data)
	if err != nil {
		s.logger.Error("Error storing resource: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// handleDelete handles DELETE requests
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Delete resource
	err := s.storage.Delete(r.Context(), path)
	if err != nil {
		s.logger.Error("Error deleting resource: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Start starts the server
func (s *Server) Start() error {
	s.logger.Info("Starting server on %s", s.Addr)
	return s.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server")
	return s.Server.Shutdown(ctx)
}
