package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/solid-go/internal/auth"
	"github.com/yourusername/solid-go/internal/events"
	"github.com/yourusername/solid-go/internal/ldp"
	"github.com/yourusername/solid-go/internal/storage"
	"github.com/yourusername/solid-go/internal/wac"
)

// Config holds the configuration for the server.
type Config struct {
	// Port to run the server on
	Port int
	// Path to store data
	StoragePath string
	// Logging level
	LogLevel string
	// Default storage strategy (memory or file)
	StorageStrategy string
	// Default authentication strategy (webid or oidc)
	AuthStrategy string
}

// ServerFacade is the main entry point for the Solid server.
// It implements the Facade pattern to simplify the API.
type ServerFacade struct {
	config          *Config
	server          *http.Server
	storageFactory  storage.Factory
	authFactory     auth.Factory
	eventDispatcher events.Dispatcher
	ldpHandler      *ldp.Handler
	wacHandler      *wac.Handler
	authHandler     *auth.Handler
	mux             *http.ServeMux
}

// NewServerFacade creates a new ServerFacade instance.
func NewServerFacade(config *Config) (*ServerFacade, error) {
	// Set default storage strategy if not specified
	if config.StorageStrategy == "" {
		config.StorageStrategy = "file"
	}

	// Set default authentication strategy if not specified
	if config.AuthStrategy == "" {
		config.AuthStrategy = "webid"
	}

	// Create event dispatcher
	eventDispatcher := events.NewDispatcher()

	// Create storage factory
	storageFactory, err := storage.NewFactory(config.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage factory: %w", err)
	}

	// Create authentication factory
	authFactory := auth.NewFactory()

	// Create LDP handler
	ldpHandler, err := ldp.NewHandler(storageFactory, eventDispatcher)
	if err != nil {
		return nil, fmt.Errorf("failed to create LDP handler: %w", err)
	}

	// Create WAC handler with storage adapter
	storageAdapter := wac.NewStorageAdapter(storageFactory)
	wacHandler, err := wac.NewHandler(storageAdapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create WAC handler: %w", err)
	}

	// Create authentication handler
	authHandler, err := auth.NewHandler(authFactory, storageFactory)
	if err != nil {
		return nil, fmt.Errorf("failed to create authentication handler: %w", err)
	}

	// Create HTTP server mux
	mux := http.NewServeMux()

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: mux,
	}

	return &ServerFacade{
		config:          config,
		server:          server,
		storageFactory:  storageFactory,
		authFactory:     authFactory,
		eventDispatcher: eventDispatcher,
		ldpHandler:      ldpHandler,
		wacHandler:      wacHandler,
		authHandler:     authHandler,
		mux:             mux,
	}, nil
}

// setupRoutes configures the HTTP routes for the server.
func (s *ServerFacade) setupRoutes() {
	// Health check endpoint
	s.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	// Main handler that coordinates authentication, WAC, and LDP
	s.mux.HandleFunc("/", s.handleSolidRequest)
}

// handleSolidRequest is the main HTTP handler for Solid requests.
func (s *ServerFacade) handleSolidRequest(w http.ResponseWriter, r *http.Request) {
	// 1. Authenticate the request
	agent, err := s.authHandler.Authenticate(r)
	if err != nil {
		http.Error(w, "Authentication failed", http.StatusUnauthorized)
		return
	}

	// Convert auth.Agent to wac.Agent
	wacAgent := wac.Agent{
		WebID: agent.ID,
		Type:  wac.User,
	}

	// Convert HTTP method to AccessMode
	var accessMode wac.AccessMode
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		accessMode = wac.Read
	case http.MethodPut, http.MethodDelete:
		accessMode = wac.Write
	case http.MethodPost:
		accessMode = wac.Append
	case http.MethodPatch:
		accessMode = wac.Write
	default:
		accessMode = wac.Read
	}

	// 2. Check access control (WAC)
	allowed, err := s.wacHandler.CheckAccess(r.URL.Path, wacAgent, accessMode)
	if err != nil {
		http.Error(w, "Error checking access control", http.StatusInternalServerError)
		return
	}
	if !allowed {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// 3. Handle the LDP request
	s.ldpHandler.ServeHTTP(w, r)
}

// Start starts the server and blocks until it's shut down.
func (s *ServerFacade) Start() error {
	// Set up routes
	s.setupRoutes()

	// Set up shutdown handler
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on port %d", s.config.Port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	log.Println("Shutting down server...")

	// Create a timeout context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}

// Stop stops the server gracefully.
func (s *ServerFacade) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// GetEventDispatcher returns the event dispatcher for the server.
func (s *ServerFacade) GetEventDispatcher() *events.Dispatcher {
	return &s.eventDispatcher
}

// GetStorageFactory returns the storage factory for the server.
func (s *ServerFacade) GetStorageFactory() storage.Factory {
	return s.storageFactory
}

// GetAuthFactory returns the authentication factory for the server.
func (s *ServerFacade) GetAuthFactory() auth.Factory {
	return s.authFactory
}
