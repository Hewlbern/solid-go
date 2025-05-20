package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"solid-go/internal/acl"
	"solid-go/internal/identity"
	"solid-go/internal/ldp"
	"solid-go/internal/storage"
)

// Operation represents an HTTP operation to be performed
type Operation struct {
	Method      string
	Target      string
	Body        []byte
	ContentType string
	Headers     http.Header
}

// Representation represents a resource representation
type Representation struct {
	Data     []byte
	Metadata map[string]string
}

// OperationHandler handles HTTP operations
type OperationHandler interface {
	Handle(ctx context.Context, op Operation) (*Representation, error)
}

// Authorizer handles authorization decisions
type Authorizer interface {
	Authorize(ctx context.Context, op Operation) error
}

// Server represents a Solid server
type Server struct {
	storage    storage.Storage
	webidStore *identity.WebIDStore
	aclStore   map[string]*acl.ACL
	containers map[string]*ldp.Container
}

// NewServer creates a new Solid server
func NewServer(storage storage.Storage) *Server {
	return &Server{
		storage:    storage,
		webidStore: identity.NewWebIDStore(),
		aclStore:   make(map[string]*acl.ACL),
		containers: make(map[string]*ldp.Container),
	}
}

// RegisterHandler registers an operation handler for a specific method
func (s *Server) RegisterHandler(method string, handler OperationHandler) {
	// Implementation needed
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Handle data directory queries
	if path == "/data" || path == "/data/" {
		s.handleDataDirectoryQuery(w, r)
		return
	}

	// Handle WebID requests
	if strings.HasPrefix(path, "/profile/") {
		s.handleWebIDRequest(w, r)
		return
	}

	// Handle ACL requests
	if strings.HasSuffix(path, ".acl") {
		s.handleACLRequest(w, r)
		return
	}

	// Handle LDP requests
	if s.isContainer(path) {
		s.handleContainerRequest(w, r)
		return
	}

	// Handle resource requests
	s.handleResourceRequest(w, r)
}

// DataDirectoryInfo represents information about the data directory
type DataDirectoryInfo struct {
	Path       string   `json:"path"`
	Resources  []string `json:"resources"`
	Containers []string `json:"containers"`
	ACLs       []string `json:"acls"`
}

// handleDataDirectoryQuery handles requests for data directory information
func (s *Server) handleDataDirectoryQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all resources
	resources, err := s.storage.List(r.Context(), "/")
	if err != nil {
		http.Error(w, "Failed to list resources", http.StatusInternalServerError)
		return
	}

	// Separate resources into different categories
	var containers []string
	var acls []string
	var regularResources []string

	for _, resource := range resources {
		switch {
		case strings.HasSuffix(resource, "/"):
			containers = append(containers, resource)
		case strings.HasSuffix(resource, ".acl"):
			acls = append(acls, resource)
		default:
			regularResources = append(regularResources, resource)
		}
	}

	info := DataDirectoryInfo{
		Path:       "/",
		Resources:  regularResources,
		Containers: containers,
		ACLs:       acls,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleWebIDRequest handles WebID profile requests
func (s *Server) handleWebIDRequest(w http.ResponseWriter, r *http.Request) {
	webid := strings.TrimPrefix(r.URL.Path, "/profile/")
	profile, err := s.webidStore.GetWebID(webid)
	if err != nil {
		http.Error(w, "WebID not found", http.StatusNotFound)
		return
	}

	// Return profile as RDF
	w.Header().Set("Content-Type", "text/turtle")
	// TODO: Convert profile to Turtle format
}

// handleACLRequest handles ACL requests
func (s *Server) handleACLRequest(w http.ResponseWriter, r *http.Request) {
	resourcePath := strings.TrimSuffix(r.URL.Path, ".acl")
	acl, exists := s.aclStore[resourcePath]
	if !exists {
		acl = acl.NewACL()
		s.aclStore[resourcePath] = acl
	}

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "text/turtle")
		// TODO: Convert ACL to Turtle format
	case http.MethodPut:
		// TODO: Parse ACL from request body
		s.aclStore[resourcePath] = acl
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleContainerRequest handles LDP container requests
func (s *Server) handleContainerRequest(w http.ResponseWriter, r *http.Request) {
	container, exists := s.containers[r.URL.Path]
	if !exists {
		container = ldp.NewContainer(s.storage, r.URL.Path)
		s.containers[r.URL.Path] = container
	}

	switch r.Method {
	case http.MethodGet:
		resources, err := container.ListResources(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// TODO: Return container listing
	case http.MethodPost:
		// TODO: Create new resource in container
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleResourceRequest handles resource requests
func (s *Server) handleResourceRequest(w http.ResponseWriter, r *http.Request) {
	// Check ACL
	if !s.checkAccess(r) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodGet:
		data, err := s.storage.Get(r.Context(), r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(data)
	case http.MethodPut:
		// TODO: Handle resource update
	case http.MethodDelete:
		err := s.storage.Delete(r.Context(), r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// checkAccess checks if the request has access to the resource
func (s *Server) checkAccess(r *http.Request) bool {
	// TODO: Implement proper access control
	return true
}

// isContainer checks if a path is a container
func (s *Server) isContainer(path string) bool {
	return strings.HasSuffix(path, "/")
}
